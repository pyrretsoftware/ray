package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func parseJsonArray(data string) []map[string]any {
	var result []map[string]any
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func handleLogs(output string, processName string, remote string) string {
	var foundProcess map[string]any
	for _, process := range parseJsonArray(output) {
		envPath := strings.Split(process["Env"].(string), "/")
		if envPath[len(envPath)-1] == processName {
			foundProcess = process
		}
	}

	if (foundProcess == nil) {
		return serror.Render("Process with that name not found.")
	}
	_log := getOutputSpin("sudo cat " + foundProcess["LogFile"].(string), remote)

	var logf logFile
	err := json.Unmarshal([]byte(_log), &logf)
	if err != nil {
		log.Fatal(err)
	}

	style := linkStyle
	if !logf.Success {
		style = serror
	}
	var steps string
	for _, step := range logf.Steps {
		sstyle := linkStyle
		if !step.Success {
			sstyle = serror
		}
		steps += "\n" +  sstyle.Render(step.Name) + "\n" + listStyle.Render(greyedOut.Render(step.Log))
	}

	return style.Render(logf.Name) + "\n" + steps
}

func formatList(output string) string {
	for _, process := range parseJsonArray(output) {
		var state string
		if process["Ghost"].(bool) {
			state = " 👻"
		} else if (process["Active"].(bool)) {
			state = " ✅ (" + process["State"].(string) + ")" 
		} else {
			state = " ❌ (error)"
		}

		fmt.Println(listStyle.Render(
			listProp.Render(process["Project"].(map[string]any)["Name"].(string) + state + "\n"),
			listProp.Render("\nInternal Port: ") + strconv.Itoa(int(process["Port"].(float64))),
			listProp.Render("\nLog file: ") + process["LogFile"].(string),
			listProp.Render("\nEnviroument: ") + process["Env"].(string),
			listProp.Render("\nHash: ") + process["Hash"].(string),
			listProp.Render("\nDeployment: ") + process["Branch"].(string),
		))
	}



	return ""
}
