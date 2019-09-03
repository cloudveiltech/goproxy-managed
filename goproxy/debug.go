package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
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
