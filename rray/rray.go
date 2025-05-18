package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func validateDomain(input string) error {
	if _, found := strings.CutPrefix(input, "http://"); found {return errors.New("Provide only the name and not a protocol, for example, 'example.com' instead of 'https://example.com'")}
	if _, found := strings.CutPrefix(input, "https://"); found {return errors.New("Provide only the name and not a protocol, for example, 'example.com' instead of 'https://example.com'")}
	return nil
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
	  Version: "2.0.0",
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
			Name: "deploy",
			Description: "deploy gets the remote associated with the git repository in the current directory, and deploys that to a remote ray server.",
			Usage: "provide the -r flag to specify the remote server",
			Flags: []cli.Flag{&remoteFlag},
			Action: func(ctx *cli.Context) error {
				validateRemote(ctx.String("remote"))

				cmd := exec.Command("git", "remote", "-v")
				ba, err := cmd.Output()
				if err != nil {
					fmt.Println(ba)
					return err
				}

				var remoteOptions []huh.Option[string]
				for line := range strings.SplitSeq(string(ba), "\n") {
					if !strings.Contains(line, "fetch") {continue}
					
					remoteOptions = append(remoteOptions, huh.NewOption(strings.Split(line, "\t")[0], strings.Split(strings.Split(line, "\t")[1], " ")[0]))
				}
				if len(remoteOptions) == 0 {
					log.Fatal("No remotes found!")
					return nil
				}
				
				var selectedRemote string
				serr := huh.NewSelect[string]().Title("Pick the remote you would like to deploy from:").Options(remoteOptions...).Value(&selectedRemote).Run()
				if serr != nil {
					log.Fatal("Error showing form.")
					return nil
				}

				var selectedName string
				var selectedDomain string
				ferr := huh.NewForm(huh.NewGroup(
					huh.NewInput().Title("Choose a name for your project:").Value(&selectedName),
					huh.NewInput().Title("Choose the domain/host to accept traffic for your project:").Validate(validateDomain).Value(&selectedDomain),
				)).Run()
				if ferr != nil {
					log.Fatal("Error showing form.")
					return nil
				}

				generatedConfig := map[string]string{
					"Name" : selectedName,
					"Domain" : selectedDomain,
					"Src" : selectedRemote,
				}
				ba, gerr := json.MarshalIndent(generatedConfig, "", "    ")
				if gerr != nil {
					log.Fatal(err)
				}

				fmt.Println("Generated project config:")
				fmt.Println(linkStyle.Render(string(ba)))
				if !confirm("Would you like to add the above configuration entry to the remote's ray config?") {return nil}

				var result map[string]any
				jerr := json.Unmarshal([]byte(getOutputSpin(raysBinary + " rray-read-config", ctx.String("remote"))), &result)
				if jerr != nil {
					log.Fatal(err)
				}
				if _, ok := result["Projects"].([]any); !ok {
					log.Fatal("failed asserting projects in remote config")
				}

				result["Projects"] = append(result["Projects"].([]any), generatedConfig)

				ba, jeerr := json.MarshalIndent(result, "", "	")
				if jeerr != nil {
					log.Fatal(err)
				}

				fmt.Println("Saving changes.")
				fmt.Print(getOutput(raysBinary + ` rray-edit-config ` + base64.StdEncoding.EncodeToString(ba), ctx.String("remote") ))
				fmt.Println("Remember, you'll also need to run " + linkStyle.Render("rray reload") + " for the changes to take effect!")
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

			  confReadOut := getOutputSpin(raysBinary + " rray-read-config", ctx.String("remote"))
			  _, readerr := file.WriteString(strings.ReplaceAll(confReadOut, "\r", ""))
			  if readerr != nil {
				log.Fatal("Failed copying over config file from remote server.")
			  }

			  fmt.Println("You may now edit the remote server's configuration using the file located at " + linkStyle.Render(file.Name()))
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
					fmt.Print(getOutput(raysBinary + ` rray-edit-config ` + base64.StdEncoding.EncodeToString(b), ctx.String("remote")))
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

						for option := range hostsFile.StoredHosts {
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
