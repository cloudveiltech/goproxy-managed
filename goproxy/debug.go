package main

import (
	"fmt"
	"os"
)

func d(msg string) {
	fmt.Fprint(os.Stderr, msg)
}
