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
	var content = process.Project.Name + state + ": On port " + strconv.Itoa(process.Port) + ", " + strconv.Itoa(len(process.Processes)) + " active processes."

	for _, process := range process.Processes {
		content += "\n- PID: "+ strconv.Itoa(process)
	}
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
		rlog.Println("Setup and deployed all projects.")
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
		}
	}
}