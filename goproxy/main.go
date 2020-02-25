package main

import "C"

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime/debug"
	"unsafe"
)

const (
	SUCCESS                = 1
	ERROR_PORTS_BUSY       = -1
	ERROR_CERTS_GENERATION = -2
)

func checkPortAvailable(port int16) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	defer l.Close()

	if err != nil {
		// Log or report the error here
		return false
	}
	return true
}

//export SetProxyLogFile
func SetProxyLogFile(logFile *C.char) {
	logPath := C.GoString(logFile)
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}

	redirectStderr(file)
}

//export AdBlockMatcherSetBlacklistCallback
func AdBlockMatcherSetBlacklistCallback(callback unsafe.Pointer) {
	adBlockBlacklistCallback = callback
}

//export StartGoServer
func StartGoServer(portHttp int16, portHttps int16, certFileC *C.char, keyFileC *C.char) int16 {
	debug.SetTraceback("all")
	debug.SetPanicOnFault(true)

	if !checkPortAvailable(portHttp) || !checkPortAvailable(portHttps) {
		return ERROR_PORTS_BUSY
	}

	certFile := C.GoString(certFileC)
	keyFile := C.GoString(keyFileC)

	_, err := os.Stat(certFile)
	if os.IsNotExist(err) {
		if !GenerateCerts(certFile, keyFile) {
			return ERROR_CERTS_GENERATION
		}
	}

	startGoProxyServer(portHttp, portHttps, certFile, keyFile)
	return SUCCESS
}

//export StopGoServer
func StopGoServer() {
	stopGoProxyServer()
}

func main() {
	test()
}

func test() {
	log.Printf("main: starting HTTP server")
	startGoProxyServer(14500, 14501, "rootCertificate.pem", "rootPrivateKey.pem")

	log.Printf("main: serving for 1000 seconds")

	var quit = false

	for !quit {
		//line, _ = reader.ReadString('\n')
		//if strings.TrimSpace(line) == "quit" {
		//	quit = true
		//}
	}

	//	Stop()
	//	log.Printf("main: done. exiting")
}
