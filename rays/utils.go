package main

import (
	"math/rand/v2"
	"net"
	"os"
	"path"
	"strconv"
)

func pickPort() int {
	port := rand.IntN(16383) + 49152

	ln, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	if err != nil {
		return pickPort()
	}
  
	ln.Close()
	rlog.Println("Using available port " + strconv.Itoa(port) + ".")
	return port
}

var dotslash string = ""
func assignDotSlash() {
	exc, err := os.Executable()
	if err != nil {
		rlog.Fatal("Cant get current executable: " + err.Error())
	}

	dotslash = path.Join(path.Dir(exc), "ray-env")
}

func getProcessFromBranch(branch string) *process {
	for _, process := range processes {
		if (process.Branch == branch && process.Ghost == false && process.State != "drop") {
			return process
		}
	}
	return nil
}