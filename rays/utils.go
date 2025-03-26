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

func getProcessFromBranch(branch string) *process {
	for _, process := range processes {
		if (process.Branch == branch && process.Ghost == false && process.State != "drop") {
			return process
		}
	}
	return nil
}