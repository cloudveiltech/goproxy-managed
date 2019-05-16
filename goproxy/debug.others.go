// +build !windows

package main

import (
	"log"
	"os"
)

func redirectStderr(f *os.File) {
	os.Stderr = f
	log.SetOutput(f)
}

