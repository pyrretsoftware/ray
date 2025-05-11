package main

import (
	"encoding/hex"
	"math/rand/v2"
	"net"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
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

	return uuid[:len(strings.Split(uuid, ""))-1]
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
	rerr.Fatal("Cant get current executable: ", err, true)

	dotslash = path.Join(path.Dir(exc), "ray-env")
}

func getProcessFromBranch(branch string, project project) *process {
	for _, process := range processes {
		if (process.Project.Name == project.Name && process.Branch == branch && !process.Ghost && process.State != "drop") {
			return process
		}
	}
	return nil
}