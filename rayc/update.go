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
	err, _ := makeRequest(cmd.String("remote"), comRequest{
		Action: "ray:update",
		Key: cmd.String("hardkey"),
		Payload: map[string]string{},
	})
	loading.Stop()
	fmt.Println(greenBold.Render("All updates that are available have been applied!"))
	return err
}