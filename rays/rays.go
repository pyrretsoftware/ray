package main

import (
	"os"
)

var Version = "unknown version"

func main() {
	if (!checkPerms()) {
		rlog.Fatal("To use the ray CLI or to launch rays you need to run as root or using sudo")
	}
	assignDotSlash()
	if (len(os.Args) == 1) {
		rlog.Fatal("No arguments passed!")
	}

	if (os.Args[1] == "daemon") {
		_cnf := readConfig()
		rconf = &_cnf
		
		rlog.Println("Ray server daemon launched.")
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