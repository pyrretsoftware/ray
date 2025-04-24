package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

func validateRemote(remote string) {
	_, hostExists := hostsFile.StoredHosts[remote]
	_, authExists := hostsFile.StoredAuth[remote]
	if (!hostExists || !authExists) {
		log.Fatal("Invalid remote name")
	}
}

var listProp = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true)
var listStyle = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1).
	Border(lipgloss.RoundedBorder())

func main() {	
	readHostsFile()
	remoteFlag := cli.StringFlag{
		Name: "remote",
		Aliases: []string{"r"},
		Usage: "specifies the remote server to connect to",
		Required: true,
	}

	app := &cli.App{
	 Name: "rray",
	  Usage: "connect to remote ray servers",
	  Description: "utility for connecting and interacting with remote ray servers",
	  Version: "1.0.0",
	  Authors: []*cli.Author{
		  {
			  Name: "axell",
			  Email: "mail@axell.me",
		  },
	  },
	  Suggest: true,
	  Commands: []*cli.Command{
		  {
			  Name: "list",
			  Description: "list the currently running ray processes on the remote server",
			  Usage: "provide the -r flag to specify the remote server",
			  Flags: []cli.Flag{&remoteFlag},
			  Action: func(ctx *cli.Context) error {
				validateRemote(ctx.String("remote"))
				fmt.Println(formatList(getOutputSpin("sudo rays list rray", ctx.String("remote"))))
				return nil
			  },
		  },
		  {
			Name: "remote",
			Description: "manage the list of remotes",
			Usage: "provide a subcommand to specify an action",
			Action: func(ctx *cli.Context) error {
				
				log.Fatal("Please provide a subcommand to specify an action.")
				return nil
			},
			Subcommands: []*cli.Command{
				{
					Name: "add",
					Aliases: []string{"edit"},
					Description: "adds or edit a remote",
					Action: func(ctx *cli.Context) error {
						var remoteName string
						var remoteHost string
						var remoteAuth []string
						var remoteUser string
						var remotePw string

						form := huh.NewForm(
							huh.NewGroup(
								huh.NewInput().
									Title("What name would like to give this remote?").
									Value(&remoteName),
								huh.NewInput().
									Title("What hostname/address and optionally port does the remote have?").
									Value(&remoteHost),
								huh.NewInput().
									Title("What user on the remote server do you want to use?").
									Value(&remoteUser),
								huh.NewInput().
									Title("What password does this user have (used for sudo)?").
									Value(&remotePw).
									EchoMode(huh.EchoModePassword),
								huh.NewMultiSelect[string]().
									Title("What authentication methods would you like to use for this remote?").Options(
										huh.Option[string]{
											Key: "Password",
											Value: "password",
										},
										huh.Option[string]{
											Key: "Public/private key",
											Value: "publickey",
										},
										huh.Option[string]{
											Key: "Interactive keyboard challenge",
											Value: "keyboard-interactive",
										},
									).Validate(func(s []string) error {
										if len(s) == 0 {
											return errors.New("Select at least one authentication method. (use x)")
										}
										return nil
									}).
									Value(&remoteAuth),
							),
						)
						ferr := form.Run()
						if ferr != nil {
							log.Fatal(ferr)
						}

						if !strings.Contains(remoteHost, ":") {
							remoteHost += ":22"
						}

						var authentication []StoredAuth
						for _, authMethod := range remoteAuth {
							var storedAuth StoredAuth 
							storedAuth.Type = authMethod

							if authMethod == "password" {
								var pw string
								if huh.NewInput().
									Title("What password do you want to authenticate with?").
									EchoMode(huh.EchoModePassword).
									Value(&pw).
									Run() != nil {
										log.Fatal("Failed asking for password")
								}

								storedAuth.Value = pw
							}
							authentication = append(authentication, storedAuth)
						}
						if (hostsFile.StoredHosts == nil) {
							hostsFile.StoredHosts = map[string]StoredHost{}
						}
						if (hostsFile.StoredAuth == nil) {
							hostsFile.StoredAuth = map[string][]StoredAuth{}
						}

						hostsFile.StoredAuth[remoteName] = authentication
						hostsFile.StoredHosts[remoteName] = StoredHost{
							Host: remoteHost,
							User: remoteUser,
							Password: remotePw,
						}

						writeHostsFile()
						return nil
					},
				},
				{
					Name: "list",
					Description: "lists all remotes",
					Action: func(ctx *cli.Context) error {
						for name, host := range hostsFile.StoredHosts {
							fmt.Println(listStyle.Render(
								listProp.Render("Name: ") + name,
								listProp.Render("\nHost: ") + host.Host,
								listProp.Render("\nUser: ") + host.User,
							))
						}

						return nil
					},
				},
				{
					Name: "remove",
					Description: "removes a remote",
					Action: func(ctx *cli.Context) error {
						var options []huh.Option[string]

						for option, _ := range hostsFile.StoredHosts {
							options = append(options, huh.NewOption(option, option))
						}

						var rremove string
						if huh.NewSelect[string]().Options(options...).Title("What remote would you like to remove?").Value(&rremove).Run() != nil {
							log.Fatal("Failed asking what remotes to remove")
						}

						delete(hostsFile.StoredAuth, rremove)
						delete(hostsFile.StoredHosts, rremove)
						writeHostsFile()
						return nil
					},
				},
			},
		},
	  },
	  
  }

  app.Run(os.Args)
}
