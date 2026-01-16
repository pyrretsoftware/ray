package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var Version = "1.0.0"
func about(context.Context, *cli.Command) error {
	fmt.Println(greenBold.Render("rayc"), "is a cli-based ray comline client, it is the offical/recommended way to manage ray servers with comlines.\nIt can talk to comlines using", greenBold.Render("UDS (Unix domain sockets)"), "or", greenBold.Render("HTTP") + ".")
	fmt.Println("By default, rayc will attempt to connect to a local UDS comline on this machine. You can use the", greenBold.Render("-r flag"), "to specify a remote HTTP comline to use.")
	fmt.Println()
	fmt.Println("Running rayc version", greenBold.Render(Version))
	return nil
}

func isBadFormat(ok bool) {
	if !ok {
		fmt.Println(redBold.Render("Comline request returned an unexpected format, try upgrading rayc and rays to their latest versions."))
		os.Exit(1)
	}
}

func badFormat() error {
	fmt.Println(redBold.Render("Comline request returned an unexpected format, try upgrading rayc and rays to their latest versions."))
	return errors.New("comline request returned unknown format")
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
			&cli.BoolFlag{
				Name: "debug-local-rays",
				Value: false,
				Usage: "for debugging use, do not use!",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "about",
				Usage: "returns information about rayc",
				Action: about,
			},
			{
				Name: "logs",
				Usage: "allows you to view process logs of processes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "process",
						Value: "",
						Usage: "the id of the process you would like to view",
					},
				},
				Action: logs,
			},
			{
				Name: "build-logs",
				Usage: "allows you to view build logs of processes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "process",
						Value: "",
						Usage: "the id of the process you would like to view",
					},
				},
				Aliases: []string{"blogs", "buildlogs"},
				Action: logs,
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
				Usage: "manually checks for updates on all projects, and updates those that are outdated.",
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
			{
				Name: "extensions",
				Aliases: []string{"ext"},
				Usage: "lists all active extensions",
				Action: extensions,
			},
			{
				Name: "list",
				Usage: "lists all processes",
				Action: list,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "ghost",
						Aliases: []string{"gh"},
						Usage: "also show ghost processes",
					},
				},
			},
		},
	}

	cli.Run(context.Background(), os.Args);
}