package main

import (
	"encoding/json"
	"strconv"
	"strings"
)

//Register any plugins here
var plugins = map[string]func(project project) string{
	"raystatus" : generateStatus,
}

//plugin code
// #region plugins
func generateStatus(project project) string {
	var status rayStatus
	status.Name = project.Options["RSName"]
	status.Desc = project.Options["RSDesc"]
	
	up := true
	for _, proc := range processes {
		if proc.Ghost {
			continue
		}

		if !proc.Active {
			up = false
		}

		status.Processes = append(status.Processes, statusItem{
			Running: proc.Active,
			Text:    proc.Project.Name + " (" + proc.Branch + " channel)",
			Subtext: "Git hash: " + strings.TrimLeft(proc.Hash, "0")[:8] + `, Listenting on ` + proc.Project.Domain + `, ` + strconv.Itoa(len(proc.Processes)) + " Running processes.",
		})
	}
	status.EverythingUp = up

	statusJson, err := json.Marshal(status)
	if err != nil {
		rlog.Notify("could not generate status.", "err")
		return ""
	}

	return string(statusJson)
}
// #endregion 

func invokePlugin(project project) (string, bool) {
	if (plugins[project.PluginImplementation] != nil) {
		return plugins[project.PluginImplementation](project), true
	}
	return "", false
}