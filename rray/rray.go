package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v2"
)

func validateRemote(remote string) {
	_, hostExists := hostsFile.StoredHosts[remote]
	_, authExists := hostsFile.StoredAuth[remote]
	if (!hostExists || !authExists) {
		log.Fatal("Invalid remote name")
	}
}

func main() {	
	readHostsFile()

	raysBinary := "sudo rays"
	if runtime.GOOS == "windows" {
		raysBinary = "%%USERPROFILE%%/rays.exe"
	}

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
				fmt.Println(formatList(getOutputSpin(raysBinary + " list rray", ctx.String("remote"))))
				return nil
			  },
		  },
		  {
			Name: "reload",
			Description: "reloads the currently running ray processes on the remote server",
			Usage: "provide the -r flag to specify the remote server",
			Flags: []cli.Flag{&remoteFlag},
			Action: func(ctx *cli.Context) error {
			  validateRemote(ctx.String("remote"))
			  fmt.Println(getOutputSpin(raysBinary + " reload rray", ctx.String("remote")))
			  return nil
			},
		  },
		  {
			Name: "logs",
			Description: "logs returns the log of a process on the remote server",
			Usage: "provide the -r flag to specify the remote server",
			Flags: []cli.Flag{&remoteFlag, &cli.StringFlag{
				Name: "process",
				Aliases: []string{"p"},
				Usage: "specifies the process to inspect",
				Required: true,
			},
			},
			Action: func(ctx *cli.Context) error {
			  validateRemote(ctx.String("remote"))
			  fmt.Println(handleLogs(getOutputSpin(raysBinary + " list rray", ctx.String("remote")), ctx.String("process"), ctx.String("remote")))
			  return nil
			},
		  },
		  {
			Name: "config",
			Description: "config allows you to edit the remote ray config",
			Usage: "provide the -r flag to specify the remote server",
			Flags: []cli.Flag{&remoteFlag},
			Action: func(ctx *cli.Context) error {
			  validateRemote(ctx.String("remote"))
			  
			  file, err := os.CreateTemp(os.TempDir(), "rayconfig-*.json")
			  if err != nil {
				log.Fatal("Failed creating local config file.")
			  }

			  _, readerr := file.WriteString(getOutputSpin(raysBinary + " rray-read-config", ctx.String("remote")))
			  if readerr != nil {
				log.Fatal("Failed copying over config file from remote server.")
			  }

			  fmt.Println("You may now edit the remote servers configuration using the file located at " + linkStyle.Render(file.Name()))
			  fmt.Println(greyedOut.Render("s - save changes • d - discard changes"))

			  reader := bufio.NewReader(os.Stdin)

			  for {
				  response, err := reader.ReadByte()
				  if err != nil {
					  log.Fatal(err)
				  }

				  if response == byte('s') {
					fmt.Println("Saving changes.")
					b, err := os.ReadFile(file.Name())
					if err != nil {
						log.Fatal("Failed reading local config file.")
					}

					fmt.Print(getOutput(raysBinary + ` rray-edit-config ` + base64.StdEncoding.EncodeToString(b), ctx.String("remote") ))
					fmt.Println("Remember, you'll also need to run " + linkStyle.Render("rray reload") + " for the changes to take effect!")
					file.Close(); os.Remove(file.Name())
					return nil
				  } else if response == byte('d') {
					fmt.Println("Discarded changes.")
					file.Close(); os.Remove(file.Name())
					return nil
				  }
			  }
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
