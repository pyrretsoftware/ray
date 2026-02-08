package main

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Register any plugins here
var plugins = map[string]func(project project) string{
	"raystatus": generateStatus,
}

// plugin code
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
		listentingOn := `, Listenting on ` + proc.Project.Domain + `, `
		if proc.ProjectConfig.NonNetworked {
			listentingOn = ", "
		}

		trimmedHash := strings.TrimLeft(proc.Hash, "0")
		if len(trimmedHash) > 8 {
			trimmedHash = trimmedHash[:8]
		}
		
		status.Processes = append(status.Processes, statusItem{
			Running: proc.Active,
			Text:    proc.Project.Name + " (" + proc.Branch + " channel)",
			Subtext: "Git hash: " + trimmedHash + listentingOn + strconv.Itoa(len(proc.Processes)) + " Running process.",
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

func invokePlugin(process process, project project) (string, bool) {
	if process.ProjectConfig == nil {
		return "", false
	}

	if plugins[process.ProjectConfig.PluginImplementation] != nil {
		return plugins[process.ProjectConfig.PluginImplementation](project), true
	} else if process.ProjectConfig.PluginImplementation != "" {
		rlog.Notify("Unknown plugin '" + process.ProjectConfig.PluginImplementation + "'.", "warn")
	}
	return "", false
}
