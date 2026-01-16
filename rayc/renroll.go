package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func renroll(cc context.Context, cmd *cli.Command) error {
	err, _ := makeRequest(cmd.String("remote"), comRequest{
		Action: "channel:renroll",
		Key: cmd.String("hardkey"),
		Payload: map[string]string{
			"project" : cmd.String("project"),
		},
	}, cmd.Bool("debug-local-rays"))
	if err != nil {
		return err
	}

	return nil
}