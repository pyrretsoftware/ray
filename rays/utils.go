package main

import (
	"os"
	"path"
)

var dotslash string = ""
func assignDotSlash() {
	exc, err := os.Executable()
	if err != nil {
		rlog.Fatal("Cant get current executable: " + err.Error())
	}

	dotslash = path.Dir(exc)
}