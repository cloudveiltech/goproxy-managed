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
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/cloudveiltech/goproxy"
	"github.com/inconshreveable/go-vhost"
	"github.com/things-go/go-socks5"
)

type Config struct {
	portHttp  int16
	portHttps int16
}

const VERSION_SOCKS5 = byte(0x05)

var (
	proxy        *goproxy.ProxyHttpServer
	server       *http.Server
	socks5Server *socks5.Server
	config       = Config{8080, 8081}

	beforeRequestCallback  unsafe.Pointer
	beforeResponseCallback unsafe.Pointer

	portWriteLock sync.RWMutex
	portMap       = make(map[int]int)
	addressMap    = make(map[int]string)
)

//var fileData []byte

type HttpsHandler func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string)

func (f HttpsHandler) HandleConnect(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return f(host, ctx)
}

var handleConnectFunc HttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	hostWithoutPort := host
	parts := strings.Split(hostWithoutPort, ":")
	if len(parts) > 1 {
		hostWithoutPort = strings.ReplaceAll(hostWithoutPort, ":"+parts[len(parts)-1], "")
	}

	if adBlockMatcher.IsDomainWhitelisted(hostWithoutPort) {
		log.Printf("Whitelisting host %s", host)
		return goproxy.OkConnect, host
	}

	ips, err := net.LookupIP(hostWithoutPort)
	if err == nil {
		for _, ip := range ips {
			isPrivateNetwork := ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalMulticast()
			if isPrivateNetwork {
				log.Printf("Private network host %s", host)
				return goproxy.OkConnect, host
			}
		}
	} else {
		log.Printf("Host lookup err: %v", err)
	}

	return goproxy.MitmConnect, host
}

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
		if req.Host == "" {
			fmt.Fprintln(w, "Cannot handle requests without Host header, e.g., HTTP 1.0")
			return
		}
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
		proxy.ServeHTTP(w, req)
	})

	proxy.Tr = &http.Transport{
		MaxIdleConnsPerHost:   10,
		MaxIdleConns:          1000,
		IdleConnTimeout:       time.Minute * 10,
		ResponseHeaderTimeout: time.Minute * 10,
	}

	proxy.OnRequest().HandleConnect(handleConnectFunc)

	config.portHttp = portHttp
	config.portHttps = portHttps

	if proxy.Verbose {
		log.Printf("Server inited")
	}
}

func startHttpServer() *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", config.portHttp)}
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
func SetDestPortForLocalPort(localPort int, destPort int, remoteIp string) {
	portWriteLock.Lock()
	defer portWriteLock.Unlock()
	addressMap[localPort] = remoteIp
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

	debug.SetTraceback("all")

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
			ip := net.ParseIP(r.RemoteAddr)
			if ip != nil {
				isPrivateNetwork := ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalMulticast()
				if isPrivateNetwork {
					userData["blocked"] = false
					log.Printf("Http whiltelisting private ip")
					return request, nil
				}
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
				}

			}

			if strings.Contains(request.URL.Host, "vimeo") {
				request.Header.Set("cookie", CookiePatchSafeSearch(request.URL.Host, request.Header.Get("cookie")))
			}

			request.URL.RawPath = HostPathForceSafeSearch(request.URL.Host, request.URL.RawPath)
			return request, response
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if resp == nil {
				return nil
			}
			response := resp

			var isVerified bool = true

			contentType := response.Header.Get("Content-Type")
			isContentTypeFilterable := isContentTypeFilterable(contentType)
			if !isContentTypeFilterable {
				return resp
			}
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
			}

			return response
		})

	go runHttpsListener()
	createSocksHandler()

	if proxy.Verbose {
		log.Printf("Server started %d, %d", config.portHttp, config.portHttps)
	}

	monitorMemoryUsage()
}

func createSocksHandler() {
	option1 := socks5.WithDialAndRequest(func(ctx context.Context, network, addr string, request *socks5.Request) (net.Conn, error) {
		log.Printf("Socks 5 dialer called %v", addr)

		conn, err := net.Dial("tcp", net.JoinHostPort("::1", strconv.Itoa(int(config.portHttps))))
		SetDestPortForLocalPort(conn.LocalAddr().(*net.TCPAddr).Port, request.DstAddr.Port, addr)
		return conn, err
	})
	option2 := socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags)))
	socks5Server = socks5.NewServer(option1, option2)

	log.Printf("Socks5 handler created")
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
			buffered := newBufferedConn(c)
			tmp, err := buffered.Peek(2)
			if err != nil {
				log.Printf("Can't read from %v", err)
				return
			}

			if tmp[0] == VERSION_SOCKS5 {
				log.Printf("Socks request detected")
				err = socks5Server.ServeConn(buffered)
				if err != nil {
					log.Printf("Can't serve socks %v", err)
				}
				return
			}
			tlsConn, err := vhost.TLS(buffered)
			localPort := tlsConn.RemoteAddr().(*net.TCPAddr).Port

			if err != nil {
				log.Printf("Assuming plain http connection - %v", err)
				httpConn, err := vhost.HTTP(tlsConn)
				if err != nil {
					log.Printf("Not http either, dropping - %v", err)
					httpConn.Close()
					return
				}

				chainReqToLocalServer(httpConn, int(config.portHttp))
				return
			}

			port := DEFAULT_HTTPS_PORT
			ipString := "127.0.0.1"
			remoteAddr := tlsConn.LocalAddr().String()
			if strings.Count(remoteAddr, ":") > 1 {
				//ipv6
				ipString = "::1"
			}
			exists := false
			attempts := 0
			for attempts < 3000 {
				portWriteLock.RLock()
				port, exists = portMap[localPort]
				ipString, exists = addressMap[localPort]
				portWriteLock.RUnlock()
				if !exists {
					time.Sleep(1 * time.Millisecond)
					port = DEFAULT_HTTPS_PORT
					//	log.Printf("Waiting for port data")
				} else {
					break
				}
				attempts = attempts + 1
			}

			ip := net.ParseIP(ipString)

			isPrivateNetwork := ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalMulticast()
			if isPrivateNetwork {
				if proxy.Verbose {
					log.Printf("Chain local IP wihout filtering %s", ipString)
				}
				chainReqWithoutFiltering(tlsConn, ip.String(), port)
				return
			}

			if proxy.Verbose {
				log.Printf("Read port: %d for ip %v", port, ipString)
			}

			host := tlsConn.Host()
			if host == "" {
				log.Printf("Error reading tls host")
				host = ipString
			}

			if adBlockMatcher.IsDomainWhitelisted(host) {
				log.Printf("Early whitelisting https host %s", host)
				chainReqWithoutFiltering(tlsConn, host, port)
				return
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

func chainReqToLocalServer(client net.Conn, port int) {
	localAddress := client.LocalAddr().(*net.TCPAddr).IP

	remote, err := net.Dial("tcp", net.JoinHostPort(localAddress.String(), strconv.Itoa(port)))
	if err != nil {
		log.Printf("chainReqToLocalServer error connect %s", err)
		return
	}

	defer remote.Close()
	defer client.Close()

	go func() {
		nonBlockingCopy(remote, client)
	}()

	nonBlockingCopy(client, remote)
}

func chainReqWithoutFiltering(client net.Conn, host string, port int) {
	defer client.Close()
	remote, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		log.Printf("chainReqWithoutFiltering error connect %s", err)
		return
	}

	defer remote.Close()

	go func() {
		io.Copy(remote, client)
	}()

	io.Copy(client, remote)
}

func nonBlockingCopy(from, to net.Conn) {
	buf := make([]byte, 10240)
	for {
		from.SetDeadline(time.Now().Add(time.Minute * 10))
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
	//testUtls()
	log.Printf("main: starting HTTP server")

	Init(14500, 14501, "rootCertificate.pem", "rootPrivateKey.pem")
	Start()
	//SetProxyLogFile("text.log")

	AdBlockMatcherInitialize()
	AdblockMatcherLoadingFinished()

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
