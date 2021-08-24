package main

/*
typedef int (*callback)(long long id);
typedef int (*adBlockCallback)(long long id, _GoString_ url, int* categories, int categoryLen);

static inline int FireCallback(void *ptr, long long id)
{
	callback p = (callback)ptr;
	return p(id);
}

static inline int FireAdblockCallback(void* ptr, long long id, _GoString_ url, int* categories, int categoryLen)
{
	adBlockCallback p = (adBlockCallback)ptr;
	return p(id, url, categories, categoryLen);
}

*/
import "C"

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/cloudveiltech/goproxy"
	"github.com/inconshreveable/go-vhost"
)

type Config struct {
	portHttp  int16
	portHttps int16
}

var (
	proxy  *goproxy.ProxyHttpServer
	server *http.Server
	config = Config{8080, 8081}

	beforeRequestCallback  unsafe.Pointer
	beforeResponseCallback unsafe.Pointer

	portWriteLock sync.Mutex
	portMap       = make(map[int]int)
)

const proxyNextActionKey string = "__proxyNextAction__"
const DEFAULT_HTTPS_PORT = 443

//export SetOnBeforeRequestCallback
func SetOnBeforeRequestCallback(callback unsafe.Pointer) {
	beforeRequestCallback = callback
}

//export SetOnBeforeResponseCallback
func SetOnBeforeResponseCallback(callback unsafe.Pointer) {
	beforeResponseCallback = callback
}

//export SetProxyLogFile
func SetProxyLogFile(logFile string) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}

	redirectStderr(file)
}

//export Init
func Init(portHttp int16, portHttps int16, certFile string, keyFile string) {
	loadAndSetCa(certFile, keyFile)
	proxy = goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.Http2Handler = serveHttp2Filtering

	if proxy.Verbose {
		log.Printf("certFilePath %s", certFile)
	}

	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("NonproxyHandler fired.")

		if req.Host == "" {
			fmt.Fprintln(w, "Cannot handle requests without Host header, e.g., HTTP 1.0")
			return
		}

		req.URL.Scheme = "http"
		req.URL.Host = req.Host
		proxy.ServeHTTP(w, req)
	})

	proxy.Tr = &http.Transport{
		MaxIdleConnsPerHost: 10,
		MaxIdleConns:        1000,
		IdleConnTimeout:     time.Minute * 10,
		TLSClientConfig: &tls.Config{
			NextProtos:               []string{"http/1.1"},
			InsecureSkipVerify:       true,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			Renegotiation:            tls.RenegotiateFreelyAsClient,
		},
	}
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	config.portHttp = portHttp
	config.portHttps = portHttps

	if proxy.Verbose {
		log.Printf("Server inited")
	}
}

func startHttpServer() *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", config.portHttp)}
	srv.Handler = proxy
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
			server = nil
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

//export SetDestPortForLocalPort
func SetDestPortForLocalPort(localPort int, destPort int) {
	if destPort == DEFAULT_HTTPS_PORT {
		return
	}
	portWriteLock.Lock()
	defer portWriteLock.Unlock()
	portMap[localPort] = destPort
}

//export Start
func Start() {
	if proxy == nil {
		return
	}

	if proxy.Verbose {
		log.Printf("Server is about to start")
	}

	server = startHttpServer()

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			userData := make(map[string]interface{})
			ctx.UserData = userData

			//dumpRequest(r)
			request := r
			var response *http.Response = nil
			session := session{r, nil, false}
			id := saveSessionToInteropMap(ctx.Session, &session)
			defer removeSessionFromInteropMap(id)

			if beforeRequestCallback != nil {
				C.FireCallback(beforeRequestCallback, C.longlong(id))

				request = session.request
				response = session.response
			}

			if response != nil {
				return request, response
			}

			// Now run our matching engine.
			if AdBlockMatcherAreListsLoaded() {

				url := request.URL.String()
				host := request.URL.Hostname()

				// adBlockMatcher is in adblock_interop.go
				categories, matchTypes := adBlockMatcher.TestUrlBlockedWithMatcherCategories(url, host, request.Referer())
				if len(categories) > 0 {
					for index, category := range categories {
						if category.ListType == Whitelist || matchTypes[index] == Excluded {
							userData["blocked"] = false

							log.Printf("Whitelisted categories matched %s", request.URL.String())
							if onWhitelistCallback != nil {
								categoryInts := TransformMatcherCategoryArrayToIntArray(categories)

								C.FireAdblockCallback(onWhitelistCallback, C.longlong(id), url, (*C.int)(&categoryInts[0]), C.int(len(categoryInts)))

								request = session.request
							}

							return request, nil
						}
					}

					if categories[0].ListType == Blacklist || categories[0].ListType == BypassList {
						userData["blocked"] = true

						log.Printf("Blacklisted categories matched %s", request.URL.String())
						if onBlacklistCallback != nil {
							categoryInts := TransformMatcherCategoryArrayToIntArray(categories)

							C.FireAdblockCallback(onBlacklistCallback, C.longlong(id), url, (*C.int)(&categoryInts[0]), C.int(len(categoryInts)))

							request = session.request
							response = session.response
						}

						return request, response //goproxy.NewResponse(request, "text/plain", 401, "Blocked by rules")
					}
				} else {
					log.Printf("No categories matched %s", request.URL.String())
				}
			}

			request.URL.RawPath = HostPathForceSafeSearch(request.URL.Host, request.URL.RawPath)
			return request, response
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			response := resp

			var isVerified bool = true

			if response != nil && response.TLS != nil {
				var err error
				isVerified, err = verifyCerts(ctx.Req.URL.Host, response.TLS.PeerCertificates)
				if err != nil {
					isVerified = false
				}
			} else {
				isVerified = false
			}

			if ctx.UserData != nil {
				userData, ok := ctx.UserData.(map[string]interface{})

				if ok {
					blocked, valueOk := userData["blocked"].(bool)
					if valueOk {
						if !blocked {
							return response
						}
					}
				}
			}

			// TODO: Call x509.Certificate.Verify
			// We should be able to glean from that whether or not we do bad SSL page.
			// A couple of things here:
			// 1. Need a boolean that says IsVerified for Response
			// 2. Need a block page that allows us to bypass it directly from the block page.
			if beforeResponseCallback != nil {
				session := session{ctx.Req, resp, isVerified}
				session.isCertVerified = isVerified
				id := saveSessionToInteropMap(ctx.Session, &session)
				C.FireCallback(beforeResponseCallback, C.longlong(id))
				removeSessionFromInteropMap(id)

				response = session.response
				//log.Printf("OnBeforeResponse overhead time: %v, %v", time.Since(startTime), id)
			}

			return response
		})

	go runHttpsListener()

	if proxy.Verbose {
		log.Printf("Server started %d, %d", config.portHttp, config.portHttps)
	}

	monitorMemoryUsage()
}

func runHttpsListener() {
	log.Printf("runHttpsListener() %d", config.portHttps)

	// listen to the TLS ClientHello but make it a CONNECT request instead
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", config.portHttps))

	if err != nil {
		log.Printf("Error listening for https connections - %v", err)
		return
	}

	for {
		if !IsRunning() {
			log.Printf("Stopping https listener")
			ln.Close()
			return
		}

		c, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting new connection - %v", err)
			continue
		}

		go func(c net.Conn) {
			tlsConn, err := vhost.TLS(c)
			if err != nil {
				log.Printf("Assuming plain http connection - %v", err)
				chainReqToHttp(tlsConn)
				return
			}

			localPort := tlsConn.RemoteAddr().(*net.TCPAddr).Port
			port, exists := portMap[localPort]

			if !exists {
				port = DEFAULT_HTTPS_PORT
			}

			if proxy.Verbose {
				log.Printf("Reading dest port for %d is %d", localPort, port)
			}

			host := tlsConn.Host()
			if host == "" {
				log.Printf("Cannot support client")
				return
			}

			if proxy.Verbose {
				log.Printf("Https handler called for %s:%s", host, strconv.Itoa(port))
			}

			host = net.JoinHostPort(host, strconv.Itoa(port))
			resp := dumbResponseWriter{tlsConn}
			connectReq := &http.Request{
				Method: "CONNECT",
				URL: &url.URL{
					Opaque: host,
					Host:   host,
				},
				Host:   host,
				Header: make(http.Header),
			}

			proxy.ServeHTTP(resp, connectReq)
		}(c)
	}
}

func chainReqToHttp(client net.Conn) {
	localAddress := client.LocalAddr().(*net.TCPAddr).IP

	log.Printf("chainReqToHttp addr %s %s", localAddress.String(), client.RemoteAddr().(*net.TCPAddr).IP.String())
	remote, err := net.Dial("tcp", net.JoinHostPort(localAddress.String(), strconv.Itoa(int(config.portHttp))))
	if err != nil {
		log.Printf("chainReqToHttp error connect %s", err)
		return
	}

	defer remote.Close()
	defer client.Close()

	go func() {
		nonBlockingCopy(remote, client)
	}()

	nonBlockingCopy(client, remote)
}

func nonBlockingCopy(from, to net.Conn) {
	log.Printf("Non blocking copy fired")

	buf := make([]byte, 10240)
	for {
		from.SetDeadline(time.Now().Add(time.Second * 10))
		if server == nil {
			log.Printf("Break chain on server stop")
			break
		}

		n, err := from.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("error request %s", err)
			break
		}
		if n == 0 {
			break
		}

		if _, err := to.Write(buf[:n]); err != nil {
			log.Printf("error response %s", err)
			break
		}

	}
}

type httpResponseWriter struct {
	net.Conn
	req         *http.Request
	header      http.Header
	status      int
	headerWrote bool
}

func (h *httpResponseWriter) Header() http.Header {
	//	panic("Header() should not be called on this ResponseWriter")

	return h.header
}

func (h *httpResponseWriter) Write(buf []byte) (int, error) {
	if !h.headerWrote {
		h.WriteHeader(200)
	}

	h.Conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", h.status, http.StatusText(h.status))))
	h.header.Write(h.Conn)
	h.Conn.Write([]byte("\r\n"))
	return h.Conn.Write(buf)
}

func (h *httpResponseWriter) WriteHeader(code int) {
	//	panic("WriteHeader() should not be called on this ResponseWriter")
	h.status = code
	h.headerWrote = true
}

func (h *httpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h, bufio.NewReadWriter(bufio.NewReader(h), bufio.NewWriter(h)), nil
}

type dumbResponseWriter struct {
	net.Conn
}

func (dumb dumbResponseWriter) Header() http.Header {
	//	panic("Header() should not be called on this ResponseWriter")

	return make(http.Header)
}

func (dumb dumbResponseWriter) Write(buf []byte) (int, error) {
	if bytes.Equal(buf, []byte("HTTP/1.0 200 OK\r\n\r\n")) {
		return len(buf), nil // throw away the HTTP OK response from the faux CONNECT request
	}

	return dumb.Conn.Write(buf)
}

func (dumb dumbResponseWriter) WriteHeader(code int) {
	//	panic("WriteHeader() should not be called on this ResponseWriter")

}

func (dumb dumbResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return dumb, bufio.NewReadWriter(bufio.NewReader(dumb), bufio.NewWriter(dumb)), nil
}

//export Stop
func Stop() {
	context, _ := context.WithTimeout(context.Background(), 1*time.Second)
	server.Shutdown(context)
	server = nil
}

//export IsRunning
func IsRunning() bool {
	return server != nil
}

//export GetCert
func GetCert(res *[]byte) {
	*res = caCert
}

func main() {
	test()
}

func test() {
	log.Printf("main: starting HTTP server")

	Init(14500, 14501, "rootCertificate.pem", "rootPrivateKey.pem")
	Start()

	log.Printf("main: serving for 1000 seconds")

	var quit = false
	var line = ""

	reader := bufio.NewReader(os.Stdin)

	for !quit {
		line, _ = reader.ReadString('\n')
		if strings.TrimSpace(line) == "quit" {
			quit = true
		}
	}

	//	Stop()
	//	log.Printf("main: done. exiting")
}
