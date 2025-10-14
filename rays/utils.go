package main

import (
	"encoding/hex"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
)

func checkPerms() bool {
	if runtime.GOOS == "linux" {
		return os.Geteuid() == 0
	}
	return true
}

func getUuid() string { //technically not a uuid ig
	uuid := ""
	for range 5 {
		section := []byte{}
		for range 4 {
			section = append(section, byte(rand.Uint32N(255)))
		}
		uuid = uuid + hex.EncodeToString(section) + "-"
	}

	return uuid
}

func makeGhost(process *process) {
	process.Ghost = true
	process.Active = false
	process.State = "drop"
	if process.Hash != "" { //hash is only unset for oci processes
		latestWorkingCommit[process.Project.Name] = process.Hash
	}
	os.RemoveAll(process.Env)
}

type ReadWaiter struct {
	w io.Writer
	close chan bool
}

func (w ReadWaiter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w ReadWaiter) Close() (err error) {
	w.close <- true
	return 
}

func (w ReadWaiter) YieldClose() {
	<- w.close
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

func portUsed(port int) bool {
	conn, err := net.Dial("tcp", ":" + strconv.Itoa(port))

	if err != nil {
		return false
	}
  
	conn.Close()
	return true
}
var dotslash string = ""
func assignDotSlash() {
	exc, err := os.Executable()
	rerr.Fatal("Cant get current executable: ", err, true)

	dotslash = path.Join(filepath.Dir(exc), "ray-env")
}

func getProcessFromBranch(branch string, project project) *process {
	for _, process := range processes {
		if (process.Project.Name == project.Name && process.Branch == branch && !process.Ghost && process.State != "drop") {
			return process
		}
	}
	return nil
}