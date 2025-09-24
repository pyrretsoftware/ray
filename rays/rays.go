package main

import (
	_ "embed"
	"os"

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
		InitalizeRls()
		SetupEnv()
		startProxy()
		select {}
	} else if os.Args[1] == "reload" {
		SendComlineRequest(LocalComline, RawComRequest{
			Action: "ray:reload",
			Key: "ext:Systemd;This extension is used by systemd to manage the local ray server;https://systemd.io/",
		})
	} else if os.Args[1] == "exit" {
		SendComlineRequest(LocalComline, RawComRequest{
			Action: "ray:shutdown",
			Key: "ext:Systemd;This extension is used by systemd to manage the local ray server;https://systemd.io/",
		})
	}
}