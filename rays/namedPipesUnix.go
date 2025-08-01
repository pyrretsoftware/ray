//go:build !windows
// +build !windows

package main

import (
	"errors"
	"net"
)

func DialNamedPipe(addr string) (net.Conn, error) {
	return nil, errors.New("Named pipes only available on windows. How did you get here??")
}