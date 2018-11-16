package main

/*
typedef void (*callback)(int id);

static inline void FireCallback(void *ptr, int id)
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
	// debug.SetTraceback("all")
	// fd, _ := os.Create("d:/work/Filter-Windows/GoProxyDotNet/testapp/err.txt")
	// redirectStderr(fd)

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
			//			r.Header.Set("X-GoProxy", "yxorPoG-X")
			if beforeRequestCallback != nil {
				log.Printf("Request call")
				id := saveRequestToInteropMap(r)
				C.FireCallback(beforeRequestCallback, C.int(id))
				removeRequestFromInteropMap(id)
				log.Printf("Request call end")
			}
			return r, nil
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if beforeResponseCallback != nil {
				//	C.FireCallback(beforeResponseCallback, unsafe.Pointer(resp))
			}
			return resp
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
