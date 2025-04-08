package main

import "encoding/json"

func generateStatus() string {
	var status rayStatus
	status.Name = rconf.RayStatus.Name
	status.Name = rconf.RayStatus.Desc

	up := true
	for _, proc := range processes {
		if (proc.Ghost) {
			continue
		}

		if (!proc.Active) {
			up = false
		}

		status.Processes = append(status.Processes, statusItem{
			Running: proc.Active,
			Text: proc.Project.Name + " (" + proc.Branch + " channel)",
			Subtext: "Git hash: "+ proc.Hash,
		})
	}
	status.Status.Running = up
	if (up) {
		status.Status.Text = "All systems operational"
	} else {
		status.Status.Text = "Experiencing issues"
	}

	statusJson, err := json.Marshal(status)
	if (err != nil) {
		rlog.Notify("could not generate status.", "err")
		return ""
	}

	rlog.Println(string(statusJson))
	return string(statusJson)
}