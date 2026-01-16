package main

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/urfave/cli/v3"
)

func update(cc context.Context, cmd *cli.Command) error {
	fmt.Println(blueBold.Render("Note:"), "For this command and for automatic updates to work you need to have supplied git http authentication.")
	loading := spinner.New(spinner.CharSets[14], 100 * time.Millisecond)
	loading.Start()
	err, resp := makeRequest(cmd.String("remote"), comRequest{
		Action: "ray:update",
		Key: cmd.String("hardkey"),
		Payload: map[string]string{},
	}, cmd.Bool("debug-local-rays"))
	loading.Stop()

	failed, ok := resp.Data.Payload.([]any)
	if !ok {
		failed = []any{}
	}
	for _, project := range failed {
		fmt.Println(yellowBold.Render("Warning:"), "Failed checking for updates on project '" + project.(string) + "'. Git http authentication may not be configured properly. (also, updates are not available in DCM)")
	}
	fmt.Println(greenBold.Render("All updates that are available have been applied!"))
	return err
}