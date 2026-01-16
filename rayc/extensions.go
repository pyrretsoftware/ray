package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func extensions(cc context.Context, cmd *cli.Command) error {
	err, resp := makeRequest(cmd.String("remote"), comRequest{
		Action: "extensions:read",
		Key:    cmd.String("hardkey"),
	}, cmd.Bool("debug-local-rays"))
	if err != nil {
		return err
	}

	pl, ok := resp.Data.Payload.(map[string]any)
	if !ok {
		return badFormat()
	}

	for name, extData := range pl {
		ext, ok := extData.(map[string]any)
		if !ok {
			fmt.Println(seperatedContent.Render(
				redBold.Render("Failed parsing extension: " + name),
			))
			continue
		}

		desc, descok := ext["Description"].(string)
		url, urlok := ext["URL"].(string)

		if !descok || !urlok {
			fmt.Println(seperatedContent.Render(
				redBold.Render("Failed parsing this extension"),
			))
			continue
		}
		fmt.Println(seperatedContent.Render(
			greenBold.Render(name),
			"\n" + desc,
			"\n" + link.Render(url),
		))
	}

	return nil
}