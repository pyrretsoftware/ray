package main

import (
	"runtime"
	_ "embed"
)

//go:embed version
var Version string
//go:embed rays
var Raysbinary []byte

var fileEnding string
func main() {
	fileEnding = ""
	if (runtime.GOOS == "windows") {
		fileEnding = ".exe"
	}
	installPack(Raysbinary)
}