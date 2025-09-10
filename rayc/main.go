package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var Version = "1.0.0"
func about(context.Context, *cli.Command) error {
	fmt.Println(greenBold.Render("rayc"), "is a cli-based ray comline client, it is the offical/recommended way to manage ray servers with comlines.\nIt can talk with comlines that use", greenBold.Render("UDS (Unix domain sockets)"), "and", greenBold.Render("HTTP") + ".")
	fmt.Println("By default, rayc will attempt to connect to a local UDS comline on this machine. You can use the", greenBold.Render("-r flag"), "to specify a remote HTTP comline to use.")
	fmt.Println()
	fmt.Println("Running rayc version", greenBold.Render(Version))
	return nil
}

func main() {
	cli := &cli.Command{
		Name: "rayc",
		Usage: "cli-based ray comline client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "remote",
				Value: "",
				Usage: "a URL to a remote http comline",
				Aliases: []string{"r"},
			},
			&cli.StringFlag{
				Name: "hardkey",
				Value: "",
				Usage: "a hardcoded key to use for remote comlines",
				Aliases: []string{"hk"},
			},
		},
		Commands: []*cli.Command{
			{
				Name: "about",
				Usage: "returns information about rayc",
				Action: about,
			},
			{
				Name: "reenroll",
				Usage: "re-enroll people already to enrolled to a channel for a project",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "project",
						Value: "",
						Usage: "the name of the project you would like to use. Use apostrophes for names with spaces.",
					},
				},
				Aliases: []string{"renroll", "re-enroll"},
				Action: renroll,
			},
			{
				Name: "config",
				Usage: "edit the config",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "project",
						Value: "",
						Usage: "",
					},
				},
				Action: config,
			},
			{
				Name: "auth",
				Usage: "generates an authentication token for accessing dev channels",
				Action: auth,
			},
			{
				Name: "reload",
				Usage: "reads and updates the server to changes in the config file, including restarting all processes.",
				Action: reload,
			},
			{
				Name: "update",
				Usage: "manually checks for updates on all projects, and updates those that are out dated.",
				Description: "This is automatically done every minute, though using this command also updates rolled-backed processes.",
				Action: update,
			},
			{
				Name: "systemctl-restart",
				Aliases: []string{"sctl-restart"},
				Usage: "restarts ray server with systemctl restart, this only works on linux with systemctl",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "fr",
						Usage: "run the command fr fr",
					},
				},
				Action: restart,
			},
		},
	}

	cli.Run(context.Background(), os.Args);
}