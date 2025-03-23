package main

import (
	"encoding/json"
	"log"
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
	content += "\n- Deployment: "+ process.Branch

	return content
}

func main() {
	if (len(os.Args) == 1) {
		rlog.Fatal("No arguments passed!")
	}

	if (os.Args[1] == "--daemon") {
		rlog.Println("Ray server daemon launched.")
		rlog.Println("Setting up ray enviroument...")
		SetupEnv()
		go daemonListen()
		startProxy()
		select {}
	} else {
		switch (os.Args[1]) {
		case "list":
			var response []process
			err := json.Unmarshal(cliSendCommand("LISTPROCESS", nil), &response)
			if (err != nil) {
				log.Fatal(err)
			}

			for _, process := range response {
				rlog.Println(formatProcessList(process))
			}
		case "reload":
			rlog.Println("Reloading processes")
			data := cliSendCommand("RELOAD", nil)
			if (string(data) == "success\n") {
				rlog.Println("Reloaded successfully.")
				os.Exit(0)
			} else {
				rlog.Println("Rays did not indicate success in reloading processes, please check the status of the server.")
				os.Exit(1)
			}
		case "force-renrollment":
			rlog.Println("Forcing a renrollment onto all users who were enrolled into a channel before this point...")
			data := cliSendCommand("FORCE_RE", nil)
			if (string(data) == "\n") {
				rlog.Notify("Applied changed to config.", "done")
			} else {
				rlog.Fatal("Failed applying changes to config.")
			}
		}
	}
}