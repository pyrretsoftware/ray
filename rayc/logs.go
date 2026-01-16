package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func logs(cc context.Context, cmd *cli.Command) error {
	ac := "log"
	if cmd.Name == "build-logs" {
		ac = "build_log"
	}

	if cmd.String("process") == "" {
		return errors.New("Please supply a process id with the --process flag.")
	}

	err, resp := makeRequest(cmd.String("remote"), comRequest{
		Action: "process:" + ac,
		Key:    cmd.String("hardkey"),
		Payload: map[string]string{
			"process" : cmd.String("process"),
		},
	}, cmd.Bool("debug-local-rays"))

	if err != nil {
		return err
	}

	pl, ok := resp.Data.Payload.([]byte)
	if !ok {
		fmt.Println(resp.Data.Payload)
		return badFormat()
	}

	name := "ray-log-*"
	if cmd.Name == "build-logs" {
		name += ".json"
	}
	file, err := os.CreateTemp("", name)
	if err != nil {
		return err
	}

	_, err = file.Write(pl)
	if err != nil {
		return err
	}

	file.Close()
	fmt.Println("The log file has been dumped to ", greenBold.Render(file.Name()))

	return nil
}