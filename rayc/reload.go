package main

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli/v3"
)

func reload(cc context.Context, cmd *cli.Command) error {
	fmt.Println(blueBold.Render("Did you know:"), "When reloading, ray router will hold all requests of a process until the new one is ready, ensuring not a single request fails.")
	loading := spinner.New(spinner.CharSets[14], 100 * time.Millisecond)
	loading.Start()
	err, _ := makeRequest(cmd.String("remote"), comRequest{
		Action: "ray:reload",
		Key: cmd.String("hardkey"),
		Payload: map[string]string{},
	})
	loading.Stop()
	fmt.Println()
	if err == nil {
		fmt.Println(greenBold.Render("All processes are now building,"), "it might take a while for all of them to be reloaded.")
	}
	return err
}