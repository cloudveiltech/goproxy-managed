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
	"fmt"
	"time"
	"unsafe"

	"log"
	"net/http"

	"gopkg.in/elazarl/goproxy.v1"
)

type Config struct {
	port int16
}

var (
	proxy  *goproxy.ProxyHttpServer
	server *http.Server
	config = Config{8080}

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
func Init(port int16) {
	//fd, _ := os.Create("err.txt")
	//redirectStderr(fd)

	loadAndSetCa()
	proxy = goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	config.port = port
}

func startHttpServer() *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", config.port)}
	srv.Handler = proxy

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
			Stop()
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

	server = startHttpServer()

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
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
			return request, response
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			response := resp
			if beforeResponseCallback != nil {
				session := session{ctx.Req, resp}
				id := saveSessionToInteropMap(ctx.Session, &session)
				C.FireCallback(beforeResponseCallback, C.longlong(id))
				removeSessionFromInteropMap(id)

				response = session.response
			}
			return response
		})
}

//export Stop
func Stop() {
	server.Shutdown(nil)
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

	Init(8081)
	Start()

	log.Printf("main: serving for 10 seconds")

	time.Sleep(10 * time.Second)

	Stop()
	log.Printf("main: done. exiting")
}
