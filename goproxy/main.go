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
	"fmt"
	"net"
	"os"
	"time"
	"unsafe"

	"log"
	"net/http"
	"net/url"

	vhost "github.com/inconshreveable/go-vhost"
	"gopkg.in/elazarl/goproxy.v1"
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
	//fd, _ := os.Create("err.txt")
	//redirectStderr(fd)

	loadAndSetCa(certFile, keyFile)
	proxy = goproxy.NewProxyHttpServer()
	proxy.Verbose = false

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

	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
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
		if err := srv.ListenAndServe(); err != nil {
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

	proxy.OnRequest().HijackConnect(func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
		defer func() {
			if e := recover(); e != nil {
				ctx.Logf("error connecting to remote: %v", e)
				client.Write([]byte("HTTP/1.1 500 Cannot reach destination\r\n\r\n"))
			}
			client.Close()
		}()
		clientBuf := bufio.NewReadWriter(bufio.NewReader(client), bufio.NewWriter(client))
		remote, err := connectDial(proxy, "tcp", req.URL.Host)
		panicOnError(err)
		remoteBuf := bufio.NewReadWriter(bufio.NewReader(remote), bufio.NewWriter(remote))

		for {
			request, err := http.ReadRequest(clientBuf.Reader)
			panicOnError(request.Write(remoteBuf))
			panicOnError(remoteBuf.Flush())

			response, err := http.ReadResponse(remoteBuf.Reader, request)
			panicOnError(err)
			panicOnError(response.Write(clientBuf.Writer))
			panicOnError(clientBuf.Flush())
		}
	})

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			startTime := time.Now()

			request := r
			var response *http.Response = nil
			if beforeRequestCallback != nil {
				session := session{r, nil}
				id := saveSessionToInteropMap(ctx.Session, &session)
				C.FireCallback(beforeRequestCallback, C.longlong(id))
				removeSessionFromInteropMap(id)

				request = session.request
				response = session.response
			}

			if time.Since(startTime) > 1 { // Cuts out all 0 second requests.
				fmt.Fprintf(os.Stderr, "OnRequest||%v||%s\n", time.Since(startTime), request.URL)
			}

			return request, response
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			startTime := time.Now()

			response := resp
			if beforeResponseCallback != nil {
				session := session{ctx.Req, resp}
				id := saveSessionToInteropMap(ctx.Session, &session)
				C.FireCallback(beforeResponseCallback, C.longlong(id))
				removeSessionFromInteropMap(id)

				response = session.response
			}

			if time.Since(startTime) > 1 {
				fmt.Fprintf(os.Stderr, "OnResponse||%v||%s\n", time.Since(startTime), ctx.Req.URL)
			}

			return response
		})

	go runHttpsListener()

	if proxy.Verbose {
		log.Printf("Server started")
	}
}

func runHttpsListener() {
	// listen to the TLS ClientHello but make it a CONNECT request instead
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", config.portHttps))
	if err != nil {
		log.Fatalf("Error listening for https connections - %v", err)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting new connection - %v", err)
			continue
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

// copied/converted from https.go
func dial(proxy *goproxy.ProxyHttpServer, network, addr string) (c net.Conn, err error) {
	if proxy.Tr.Dial != nil {
		return proxy.Tr.Dial(network, addr)
	}
	return net.Dial(network, addr)
}

// copied/converted from https.go
func connectDial(proxy *goproxy.ProxyHttpServer, network, addr string) (c net.Conn, err error) {
	if proxy.ConnectDial == nil {
		return dial(proxy, network, addr)
	}
	return proxy.ConnectDial(network, addr)
}

type dumbResponseWriter struct {
	net.Conn
}

func (dumb dumbResponseWriter) Header() http.Header {
	panic("Header() should not be called on this ResponseWriter")
}

func (dumb dumbResponseWriter) Write(buf []byte) (int, error) {
	if bytes.Equal(buf, []byte("HTTP/1.0 200 OK\r\n\r\n")) {
		return len(buf), nil // throw away the HTTP OK response from the faux CONNECT request
	}
	return dumb.Conn.Write(buf)
}

func (dumb dumbResponseWriter) WriteHeader(code int) {
	panic("WriteHeader() should not be called on this ResponseWriter")
}

func (dumb dumbResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return dumb, bufio.NewReadWriter(bufio.NewReader(dumb), bufio.NewWriter(dumb)), nil
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
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
	//test()
}

func test() {
	log.Printf("main: starting HTTP server")

	Init(14300, 14301, "cert.pem", "key.pem")
	Start()

	log.Printf("main: serving for 1000 seconds")

	time.Sleep(1000 * time.Second)

	Stop()
	log.Printf("main: done. exiting")
}
