package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli/v3"
)

func restart(cc context.Context, cmd *cli.Command) error {
	fmt.Println(redBold.Render("WARNING:"), "This command is experimental and may cause issues with ray server not starting back up properly. Please", redBold.Render("do not use this"), "if you cant get a shell on the server to fix any problems.")
	if (!cmd.Bool("fr")) {
		fmt.Println("Please", blueBold.Render("provide the -fr flag"), "when running this command to confirm you understand this.")
		return errors.New("not fr")
	}
	loading := spinner.New(spinner.CharSets[14], 100 * time.Millisecond)
	loading.Start()
	fmt.Println("Alright, attempting to restart rays with systemctl")
	fmt.Println()

	err, _ := makeRequest(cmd.String("remote"), comRequest{
		Action: "ray:systemctl:restart",
		Key: cmd.String("hardkey"),
		Payload: map[string]string{},
	})
	loading.Stop()
	return err
}