package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func auth(cc context.Context, cmd *cli.Command) error {
	err, resp := makeRequest(cmd.String("remote"), comRequest{
		Action: "channel:auth",
		Key: cmd.String("hardkey"),
	})
	if err != nil {return err}

	key, ok := resp.Data.Payload.(string)
	if !ok {
		fmt.Println(redBold.Render("Comline request returned an unexpected format, try upgrading rayc and rays to their latest versions."))
	}

	fmt.Println("Use the key", blueBold.Render(key), "to log into development channels.")
	return err
}