package main

import "C"

import (
	"fmt"
	"io"
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
	MAX_LOG_SIZE           = 10 * 1024 * 1024
)

var certsException = make(map[string]bool)
var logFilePath = ""
var logFileHandle *os.File

//export AddCertException
func AddCertException(thumbPrintC *C.char) {
	thumbPrint := C.GoString(thumbPrintC)
	_, ok := certsException[thumbPrint]
	if !ok {
		certsException[thumbPrint] = true
	}
}

func isCertInException(thumbPrint string) bool {
	_, ok := certsException[thumbPrint]
	return ok
}

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
	logFilePath = C.GoString(logFile)
	setProxyLogFileInternal(logFilePath)
}

func setProxyLogFileInternal(logFile string) {
	logFilePath = logFile
	logFileHandle, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}

	redirectStderr(logFileHandle)
}

func monitorLogFileSize() {
	stat, err := os.Stat(logFilePath)
	if err != nil {
		log.Printf("Can't stat log file %v", err)
		return
	}

	if stat.Size() > MAX_LOG_SIZE {
		log.Printf("Rotate log file")
		destination, err := os.OpenFile(logFilePath+".1", os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			log.Printf("Can't open temp file %v", err)
			return
		}
		fileHandle, err := os.OpenFile(logFilePath, os.O_RDONLY, 0666)
		if err != nil {
			log.Printf("Can't open log file %v", err)
			return
		}
		fileHandle.Seek(MAX_LOG_SIZE/2, 0)

		defer destination.Close()
		_, err = io.Copy(destination, fileHandle)
		if err != nil {
			log.Printf("Can't copy log file %v", err)
			return
		}

		logFileHandle.Close()
		os.Rename(logFilePath+".1", logFilePath)
		setProxyLogFileInternal(logFilePath)

		log.Printf("Rotate log file done.")
	}
}

//export AdBlockMatcherSetBlacklistCallback
func AdBlockMatcherSetBlacklistCallback(callback unsafe.Pointer) {
	adBlockBlacklistCallback = callback
}

//export StartGoServer
func StartGoServer(portHttp, portHttps, portConfigurationServer int16, certFileC *C.char, keyFileC *C.char) int16 {
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

	startGoProxyServer(portHttp, portHttps, portConfigurationServer, certFile, keyFile)
	monitorLogFileSize()
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
	startGoProxyServer(14600, 14501, 14502, "rootCertificate.pem", "rootPrivateKey.pem")

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
