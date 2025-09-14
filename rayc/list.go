package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/urfave/cli/v3"
)

func list(cc context.Context, cmd *cli.Command) error {
	err, resp := makeRequest(cmd.String("remote"), comRequest{
		Action: "process:read",
		Key:    cmd.String("hardkey"),
	})
	if err != nil {
		return err
	}

	pl, ok := resp.Data.Payload.([]any)
	if !ok {
		fmt.Println(redBold.Render("Comline request returned an unexpected format, try upgrading rayc and rays to their latest versions."))
		return errors.New("comline request returned unknown format")
	}
	
	table := table.New()

	table.Border(lipgloss.RoundedBorder()).Headers("State", "Identifier")
	for _, proc := range pl {
		process, ok := proc.(map[string]any)
		if !ok {continue}

		state := process["State"].(string)
		stateStyle := redBold
		if state == "OK" {
			stateStyle = greenBold
		} else if state == "drop" {
			stateStyle = yellowBold
			if !cmd.Bool("ghost") {continue}
		} else {
			state = "Errored"
		}

		project := process["Project"].(map[string]any)
		
		table.Row(project["Name"].(string), stateStyle.Render(state), process["Id"].(string))
	}
	fmt.Println(table.Render())

	return nil
}