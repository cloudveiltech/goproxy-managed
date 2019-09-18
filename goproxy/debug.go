package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

//import _ "net/http/pprof"

func d(msg string) {
	fmt.Fprint(os.Stderr, msg)
}

func monitorMemoryUsage() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			printMemUsage()
			debug.FreeOSMemory()
			if !IsRunning() {
				return
			}
		}
	}()

	//	go func() {
	//	http.ListenAndServe(":6060", nil)
	//	}()
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	log.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	log.Printf("\tSys = %v MiB", bToMb(m.Sys))
	log.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func dumpRequest(r *http.Request) {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	log.Printf("Request: %s", strings.Join(request, "\n"))
}
