package main

import "C"

import (
	"fmt"
	"unsafe"

	"log"
	"net/http"

	"github.com/dkwiebe/goproxy/callbacks"
	"gopkg.in/elazarl/goproxy.v1"
)

type Config struct {
	port int16
}

var proxy *goproxy.ProxyHttpServer
var server *http.Server
var config = Config{8080}

//export SayHello
func SayHello(out *string) int {
	*out = "Hello From GO!"
	return 1
}

//export StringCallbackFunction
func StringCallbackFunction(callback unsafe.Pointer) {
	callbacks.FireCallback(callback, "HEY CALLBACK GO")
}

//export Init
func Init(port int16) {
	proxy = goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	config.port = port

	loadAndSetCa()
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
func GetCert() []byte {
	return caCert
}

func main() {
	/*log.Printf("main: starting HTTP server")

	Init(8081)
	Start()

	log.Printf("main: serving for 10 seconds")

	time.Sleep(10 * time.Second)

	Stop()
	log.Printf("main: done. exiting")*/
}
