//go:build windows
// +build windows

package main

import (
	"net"

	"github.com/natefinch/npipe"
)

func DialNamedPipe(addr string) (net.Conn, error) {
	return npipe.Dial(addr)
}