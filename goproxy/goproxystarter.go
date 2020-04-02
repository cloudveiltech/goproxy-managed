package main

/*
#include <stdlib.h>

typedef int (*adBlockCallback)(char* url, char* category);

static inline int FireAdblockCallback(void* ptr, char* url, char* category)
{
	adBlockCallback p = (adBlockCallback)ptr;
	return p(url, category);
}
*/
import "C"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
	"unsafe"

	"github.com/elazarl/goproxy"
	"github.com/inconshreveable/go-vhost"
)

//import _ "net/http/pprof"

var (
	proxy  *goproxy.ProxyHttpServer
	server *http.Server
)

func initGoProxy() {
	proxy = goproxy.NewProxyHttpServer()
	proxy.Verbose = true

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

	if proxy.Verbose {
		log.Printf("Server inited")
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

func runHttpsListener(port int16) {
	// listen to the TLS ClientHello but make it a CONNECT request instead
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		log.Printf("Error listening for https connections - %v", err)
		return
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

			if proxy.Verbose {
				log.Printf("Https handler called for %s", tlsConn.Host())
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

func startHttpServer(port int16) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	srv.Handler = proxy

	// go func() {
	// 	http.ListenAndServe(":6060", nil)
	// }()
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

func startGoProxyServer(portHttp, portHttps int16, certPath, certKeyPath string) {
	initGoProxy()
	loadAndSetCa(certPath, certKeyPath)

	if proxy == nil {
		return
	}

	if proxy.Verbose {
		log.Printf("Server is about to start http: %d, https: %d", portHttp, portHttps)
	}

	server = startHttpServer(portHttp)

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if adBlockMatcher != nil {
				category, matchType, isRelaxedPolicy := adBlockMatcher.TestUrlBlocked(r.URL.String(), r.Host, r.Referer())
				if category != nil && matchType == Included {
					url := r.URL.String()
					if adBlockBlacklistCallback != nil {
						unsafeUrl := C.CString(url)
						unsafeCategory := C.CString(*category)
						C.FireAdblockCallback(adBlockBlacklistCallback, unsafeUrl, unsafeCategory)
						C.free(unsafe.Pointer(unsafeUrl))
						C.free(unsafe.Pointer(unsafeCategory))
					}

					return r, goproxy.NewResponse(r,
						goproxy.ContentTypeHtml, http.StatusForbidden,
						adBlockMatcher.GetBlockPage(url, *category, isRelaxedPolicy, false))
				}
			}
			return r, nil
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if resp == nil {
				return resp
			}

			if resp.StatusCode > 400 { //ignore errors
				return resp
			}

			if adBlockMatcher == nil {
				return resp
			}

			if !adBlockMatcher.TestContentTypeIsFiltrable(resp.Header.Get("Content-Type")) {
				return resp
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)

			bytesData := buf.Bytes()

			//since we'd read all body - we need to recreate reader for client here
			resp.Body.Close()
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(bytesData))

			if !adBlockMatcher.IsContentSmallEnoughToFilter(int64(len(bytesData))) {
				return resp
			}

			category := adBlockMatcher.TestContainsForbiddenPhrases(bytesData)

			if category != nil {
				message := adBlockMatcher.GetBlockPage(resp.Request.URL.String(), *category, false, true)
				return goproxy.NewResponse(resp.Request, goproxy.ContentTypeHtml, http.StatusForbidden, message)
			}
			return resp
		})

	go runHttpsListener(portHttps)

	if proxy.Verbose {
		log.Printf("Server started")
	}
}

func stopGoProxyServer() {
	if server != nil {
		context, _ := context.WithTimeout(context.Background(), 1*time.Second)
		server.Shutdown(context)
		server = nil
	}
}
