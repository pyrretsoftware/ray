package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
)

var Version = "1.0.0"
func about(context.Context, *cli.Command) error {
	fmt.Println(greenBold.Render("rayc"), "is a cli-based ray comline client, it is the offical/recommended way to manage ray servers with comlines.\nIt can talk to comlines using", greenBold.Render("UDS (Unix domain sockets)"), "or", greenBold.Render("HTTP") + ".")
	fmt.Println("By default, rayc will attempt to connect to a local UDS comline on this machine. You can use the", greenBold.Render("remote"), "command to add, remove and edit configured remote comlines.")
	fmt.Println("rayc also includes a", greenBold.Render("dev"), "command that allows you to build and run a project from a", greenBold.Render("ray.config.json"), "file")
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
var logStyles = map[string]lipgloss.Style{
	"ERR": redBold,
	"INFO": blueBold,
	"DONE": greenBold,
}

func Log(style string, a any) {
	fmt.Println("[" + logStyles[style].Render(style) + "]", a)
}

type LogWriter struct {
	Style string
	buf []byte
}


func (lw *LogWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if b == '\n' {
			Log(lw.Style, string(lw.buf))
			lw.buf = []byte{}
		} else {
			lw.buf = append(lw.buf, b)
		}
	}

	return len(p), nil
}

func badFormat() error {
	fmt.Println(redBold.Render("Comline request returned an unexpected format, try upgrading rayc and rays to their latest versions."))
	return errors.New("comline request returned unknown format")
}

var UseAccesible = os.Getenv("ACCESSIBLE") != ""

func main() {
	cli := &cli.Command{
		Name: "rayc",
		Usage: "cli-based ray comline client",
		Authors: []any{"axell (https://axell.me)"},
		ExitErrHandler: func(ctx context.Context, c *cli.Command, err error) {
			Log("ERR", err)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "remote",
				Value: "",
				Usage: "a URL to a remote HTTP comline, or the name of a configured remote comline",
				Aliases: []string{"r"},
			},
			&cli.StringFlag{
				Name: "hardkey",
				Value: "",
				Usage: "a static authentication key",
				Aliases: []string{"hk"},
			},
			&cli.BoolFlag{
				Name: "debug-local-rays",
				Aliases: []string{"debug"},
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
				Name: "remote",
				Usage: "commands to add, remove and edit configured remote servers",
				Commands: []*cli.Command{
					{
						Name: "add",
						Usage: "adds a remote to remotes.json",
						Action: AddRemote,
					},
				},
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
			{
				Name: "dev",
				Usage: "dev allows you to run a ray project locally for development",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name: "port",
						Aliases: []string{"p"},
						Value: -1,
						Usage: "the port to use, by default a sensible available port is selected",
					},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "directory",
						UsageText: "the directory to use, it should be the root of a git repository and have a ray.config.json file",
					},
				},
				Action: dev,
			},
		},
	}

	cli.Run(context.Background(), os.Args);
}