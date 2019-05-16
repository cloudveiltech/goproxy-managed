// +build !windows

package main

import (
	"log"
	"os"
)

var (
	defaultStderr *os.File = nil
)

func redirectStderr(f *os.File) {
	if defaultStderr == nil {
		defaultStderr = os.Stderr
	}

	os.Stderr = f
	log.SetOutput(f)
}
