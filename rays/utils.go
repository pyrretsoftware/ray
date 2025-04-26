package main

import (
	"math/rand/v2"
	"net"
	"os"
	"path"
	"runtime"
	"strconv"
)

func checkPerms() bool {
	if runtime.GOOS == "linux" {
		return os.Geteuid() == 0
	}
	return true
}

func makeGhost(process *process) {
	process.Ghost = true
	process.State = "drop"
	os.RemoveAll(process.Env)
}

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
	rlog.Println("Debugging notice: now assigned dotslash to " + dotslash)
}

func getProcessFromBranch(branch string, project project) *process {
	for _, process := range processes {
		if (process.Project.Name == project.Name && process.Branch == branch && !process.Ghost && process.State != "drop") {
			return process
		}
	}
	return nil
}