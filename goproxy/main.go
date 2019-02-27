package main

/*
typedef void (*callback)(long long id);

static inline void FireCallback(void *ptr, long long id)
{
	callback p = (callback)ptr;
	p(id);
}

*/
import "C"

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/cloudveiltech/goproxy"
	vhost "github.com/inconshreveable/go-vhost"
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
)

//export SetOnBeforeRequestCallback
func SetOnBeforeRequestCallback(callback unsafe.Pointer) {
	beforeRequestCallback = callback
}

//export SetOnBeforeResponseCallback
func SetOnBeforeResponseCallback(callback unsafe.Pointer) {
	beforeResponseCallback = callback
}

//export Init
func Init(portHttp int16, portHttps int16, certFile string, keyFile string) {
	/*fd, _ := os.Create("C:\\err.txt")
	redirectStderr(fd)*/

	goproxy.SetDefaultTlsConfig(defaultTLSConfig)
	loadAndSetCa(certFile, keyFile)
	proxy = goproxy.NewProxyHttpServer()
	proxy.Verbose = true

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

	proxy.WebSocketHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		h, ok := w.(http.Hijacker)
		if !ok {
			return
		}

		client, _, err := h.Hijack()
		if err != nil {
			log.Printf("Websocket error Hijack %s", err)
			return
		}

		remote := dialRemote(req)

		defer remote.Close()
		defer client.Close()

		log.Printf("Got websocket request %s %s", req.Host, req.URL)

		req.Write(remote)
		go func() {
			for {
				n, err := io.Copy(remote, client)
				if err != nil {
					log.Printf("Websocket error request %s", err)
					return
				}
				if n == 0 {
					log.Printf("Websocket nothing requested close")
					return
				}
				time.Sleep(time.Millisecond) //reduce CPU usage due to infinite nonblocking loop
			}
		}()

		for {
			n, err := io.Copy(client, remote)
			if err != nil {
				log.Printf("Websocket error response %s", err)
				return
			}
			if n == 0 {
				log.Printf("Websocket nothing responded close")
				return
			}
			time.Sleep(time.Millisecond) //reduce CPU usage due to infinite nonblocking loop
		}
	})

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	config.portHttp = portHttp
	config.portHttps = portHttps

	if proxy.Verbose {
		log.Printf("Server inited")
	}
}

func dialRemote(req *http.Request) net.Conn {
	port := ""
	if !strings.Contains(req.URL.Host, ":") {
		if req.URL.Scheme == "https" {
			port = ":443"
		} else {
			port = ":80"
		}
	}

	if req.URL.Scheme == "https" {
		conf := tls.Config{
			//InsecureSkipVerify: true,
		}
		remote, err := tls.Dial("tcp", req.URL.Host+port, &conf)
		if err != nil {
			log.Printf("Websocket error connect %s", err)
			return nil
		}
		return remote
	} else {
		remote, err := net.Dial("tcp", req.URL.Host+port)
		if err != nil {
			log.Printf("Websocket error connect %s", err)
			return nil
		}
		return remote
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
			log.Printf("OnRequest() in go")

			startTime := time.Now()

			request := r
			var response *http.Response = nil
			if beforeRequestCallback != nil {
				session := session{r, nil, false}
				id := saveSessionToInteropMap(ctx.Session, &session)
				C.FireCallback(beforeRequestCallback, C.longlong(id))
				removeSessionFromInteropMap(id)

				request = session.request
				response = session.response
			}

			if time.Since(startTime) > 1 { // Cuts out all 0 second requests.
				//			fmt.Fprintf(os.Stderr, "OnRequest||%v||%s\n", time.Since(startTime), request.URL)
			}

			return request, response
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			startTime := time.Now()

			response := resp
			var isVerified bool = true

			if response != nil && response.TLS != nil {
				dnsName := ctx.Req.URL.Host

				for _, cert := range response.TLS.PeerCertificates {

					opts := x509.VerifyOptions {
						Roots: nil,
						DNSName: dnsName,
					}

					_, err := cert.Verify(opts)

					if err != nil {
						isVerified = false
						break
					}
				}
			} else {
				isVerified = false
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

			if time.Since(startTime) > 1 {
				//		fmt.Fprintf(os.Stderr, "OnResponse||%v||%s\n", time.Since(startTime), ctx.Req.URL)
			}

			return response
		})

	go runHttpsListener()

	if proxy.Verbose {
		log.Printf("Server started %d, %d", config.portHttp, config.portHttps)
	}
}

func runHttpsListener() {
	log.Printf("runHttpsListener() %d", config.portHttps)

	// listen to the TLS ClientHello but make it a CONNECT request instead
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.portHttps))

	if err != nil {
		log.Fatalf("Error listening for https connections - %v", err)
		return
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting new connection - %v", err)
			continue
		} else {
			log.Printf("Accepting new connection")
		}

		go func(c net.Conn) {
			tlsConn, err := vhost.TLS(c)
			if err != nil {
				log.Printf("Error accepting new connection - %v", err)
			}
			if tlsConn.Host() == "" {
				log.Printf("Cannot support non-SNI enabled clients")
				return
			}
			connectReq := &http.Request{
				Method: "CONNECT",
				URL: &url.URL{
					Opaque: tlsConn.Host(),
					Host:   net.JoinHostPort(tlsConn.Host(), "443"),
				},
				Host:   tlsConn.Host(),
				Header: make(http.Header),
			}

			resp := dumbResponseWriter{tlsConn}
			proxy.ServeHTTP(resp, connectReq)
		}(c)
	}
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

	Init(23500, 14301, "rootCertificate.pem", "rootPrivateKey.pem")
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
