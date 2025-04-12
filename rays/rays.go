package main

import (
	"os"
	"strconv"
)

var _version = "v1.0.0"

func formatProcessList(process process) string {
	var state string
	if process.Ghost {
		state = " 👻"
	} else if (process.Active) {
		state = " ✅"
	} else {
		state = " ❌"
	}
	state += " (" + process.State + ")" 
	var content = process.Project.Name + state + ". " + strconv.Itoa(len(process.Processes)) + " active processes."

	for indx, process := range process.Processes {
		content += "\n- PID (process " + strconv.Itoa(indx + 1) + "): "+ strconv.Itoa(process)
	}

	content += "\n- Internal Port: "+ strconv.Itoa(process.Port)
	content += "\n- Log file: "+ process.LogFile
	content += "\n- Deployment: "+ process.Branch

	return content
}


func main() {
	if (!checkPerms()) {
		rlog.Fatal("To use the ray CLI or to launch rays you need to run as root or using sudo")
	}
	assignDotSlash()
	if (len(os.Args) == 1) {
		rlog.Fatal("No arguments passed!")
	}

	if (os.Args[1] == "--daemon") {
		rlog.Println("Ray server daemon launched.")
		rlog.Println("Setting up ray enviroument...")
		
		go daemonListen()
		SetupEnv()
		startProxy()
		select {}
	} else {
		handleCommand(os.Args)
	}
}