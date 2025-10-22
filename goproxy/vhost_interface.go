package main

import (
	"net"
)

type Conn interface {
	net.Conn
	Host() string
	Free()
}

type CloseWriter interface {
	CloseWrite() error
}
