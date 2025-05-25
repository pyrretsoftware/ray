package main

import (
	"os"
)

var Version = "unknown version"

func main() {
	assignDotSlash()
	if (!checkPerms()) {
		rlog.Fatal("To use the ray CLI or to launch rays you need to run as root or using sudo")
	}
	if (len(os.Args) == 1) {
		rlog.Fatal("No arguments passed!")
	}

	if (os.Args[1] == "daemon") {
		_cnf := readConfig()
		rconf = &_cnf
		
		rlog.Println("Ray server daemon launched.")
		go triggerEvent("raysStart", nil)
		rlog.Println("Setting up ray enviroument...")
		initRLS()
		go daemonListen()
		SetupEnv()
		startProxy()
		select {}
	} else {
		handleCommand(os.Args)
	}
}