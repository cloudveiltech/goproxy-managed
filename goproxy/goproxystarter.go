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
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/cloudveiltech/goproxy"
	"github.com/inconshreveable/go-vhost"
)

//import _ "net/http/pprof"

const DEFAULT_HTTPS_PORT uint16 = 443

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

	return goproxy.MitmConnect, host
}

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

	proxy.OnRequest().HandleConnect(handleConnectFunc)

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

func runHttpsListener() {
	log.Printf("runHttpsListener() %d", configuredPortHttps)

	// listen to the TLS ClientHello but make it a CONNECT request instead
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", configuredPortHttps))

	if err != nil {
		log.Fatalf("Error listening for https connections - %v", err)
		return
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting new connection - %v", err)
			continue
		}

		go func(c net.Conn) {
			helloBuffer := make([]byte, 2)
			n, err := c.Read(helloBuffer)

			port := DEFAULT_HTTPS_PORT
			if n > 0 {
				port = binary.BigEndian.Uint16([]byte{helloBuffer[1], helloBuffer[0]})
				log.Printf("Reading dest port for %d", port)
			}

			tlsConn, err := vhost.TLS(c)
			if err != nil {
				log.Printf("Assuming plain http connection - %v", err)
				chainReqToHttp(tlsConn)
				return
			}

			host := tlsConn.Host()
			if host == "" {
				log.Printf("Cannot support client")
				return
			}

			host = net.JoinHostPort(host, strconv.Itoa(int(port)))
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
	chainReqToHost(client, fmt.Sprintf("127.0.0.1:%d", configuredPortHttp))
}

func chainReqToHost(client net.Conn, hostPort string) {
	remote, err := net.Dial("tcp", hostPort)

	if err != nil {
		log.Printf("chainReqToHttp error connect %s", err)
		return
	}

	//	defer remote.Close()
	//	defer client.Close()

	go func() {
		io.Copy(remote, client)
	}()

	io.Copy(client, remote)
}

func startHttpServer(port int16) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port)}
	srv.Handler = proxy
	proxy.Http2Handler = serveHttp2Filtering

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

func startGoProxyServer(portHttp, portHttps, portConfigurationServer int16, certPath, certKeyPath string) {
	initGoProxy()
	loadAndSetCa(certPath, certKeyPath)
	configuredPortHttp = portHttp
	configuredPortHttps = portHttps
	configuredConfigurationServerPort = portConfigurationServer

	if proxy == nil {
		return
	}

	if proxy.Verbose {
		log.Printf("Server is about to start http: %d, https: %d", portHttp, portHttps)
	}

	server = startHttpServer(portHttp)

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			monitorLogFileSize()
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

					log.Printf("Page %s blocked by url, category %s", url, *category)

					r.URL.RawPath = HostPathForceSafeSearch(r.URL.Host, r.URL.RawPath)
					return r, goproxy.NewResponse(r,
						goproxy.ContentTypeHtml, http.StatusForbidden,
						adBlockMatcher.GetBlockPage(url, *category, isRelaxedPolicy))
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

			if resp.TLS != nil {
				var err error
				_, err = verifyCerts(ctx.Req.URL.Host, resp.TLS.PeerCertificates)
				if err != nil {
					certThumbPrint := base64.StdEncoding.EncodeToString(resp.TLS.PeerCertificates[0].Signature)

					if !isCertInException(certThumbPrint) {
						message := adBlockMatcher.GetBadCertPage(ctx.Req.URL.String(), ctx.Req.URL.Host, certThumbPrint)
						return goproxy.NewResponse(resp.Request, goproxy.ContentTypeHtml, http.StatusForbidden, message)
					}
				}
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

			bytesData = decodeResponseCompression(resp.Header.Get("Content-Encoding"), bytesData)

			category, matches := adBlockMatcher.TestContainsForbiddenPhrases(bytesData)

			if category != nil {
				log.Printf("Page %s blocked, category: %s, found forbidden phrases: %s", resp.Request.URL.String(), *category, strings.Join(matches, ", "))
				message := adBlockMatcher.GetBlockPage(resp.Request.URL.String(), *category, false)
				return goproxy.NewResponse(resp.Request, goproxy.ContentTypeHtml, http.StatusForbidden, message)
			}
			return resp
		})

	runConfigurationServerListener()
	go runHttpsListener()

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
