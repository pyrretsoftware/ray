package main

import (
	"os"
	_ "embed"
)

//go:embed version
var Version string
var DebugLogsEnabled bool

func main() {
	assignDotSlash()
	if (!checkPerms()) {
		rlog.Fatal("To launch rays you need to run as root or using sudo")
	}
	if (len(os.Args) == 1) {
		rlog.Fatal("No arguments passed!")
	}

	if (os.Args[1] == "daemon") {
		for _, arg := range os.Args {
			if arg == "-d" || arg == "--show-debug" {
				DebugLogsEnabled = true
				break
			}
		}
		_cnf := readConfig()
		rconf = &_cnf
		
		rlog.Println("Ray server daemon launched.")
		rlog.Debug("Debug messages are shown.")
		go triggerEvent("raysStart", nil)
		rlog.Println("Setting up ray enviroument...")
		initRLS()
		SetupEnv()
		startProxy()
		select {}
	} else if os.Args[1] == "reload" {
		rayReload([]string{"special:all"})
	} else if os.Args[1] == "exit" {
		cleanUpAndExit()
	}
}