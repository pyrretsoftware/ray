package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/urfave/cli/v3"
)

func list(cc context.Context, cmd *cli.Command) error {
	err, resp := makeRequest(cmd.String("remote"), comRequest{
		Action: "process:read",
		Key:    cmd.String("hardkey"),
	}, cmd.Bool("debug-local-rays"))
	if err != nil {
		return err
	}

	pl, ok := resp.Data.Payload.([]any)
	if !ok {
		return badFormat()
	}
	
	table := table.New()

	table.Border(lipgloss.RoundedBorder()).Headers("Name", "State", "Identifier", "Internal Port", "Deployment", "Hash", "PID(s)", "RLS Type")
	for _, proc := range pl {
		process, ok := proc.(map[string]any)
		if !ok {
			return badFormat()
		}

		state, ok := process["State"].(string)
		isBadFormat(ok)

		stateStyle := redBold
		if state == "OK" {
			stateStyle = greenBold
		} else if state == "drop" {
			stateStyle = yellowBold
			if !cmd.Bool("ghost") {continue}
		} else {
			state = "Errored"
		}

		project, ok := process["Project"].(map[string]any)
		isBadFormat(ok)
		rlsInfo, ok := process["RLSInfo"].(map[string]any)
		isBadFormat(ok)
		
		pids, pidsOk := process["Processes"].([]float64)
		pidsStr := "N/A"
		if pidsOk {
			pidsStr = ""
			for _, pid := range pids {
				pidsStr += ", " + strconv.FormatFloat(pid, 'g', -1, 64)
			}
		}

		table.Row(project["Name"].(string), stateStyle.Render(state), process["Id"].(string), strconv.FormatFloat(process["Port"].(float64), 'g', -1, 64), process["Branch"].(string), process["Hash"].(string), pidsStr, rlsInfo["Type"].(string))
	}
	fmt.Println(table.Render())

	return nil
}