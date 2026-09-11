package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
	"pyrret.com/rays/prjcnf"
	"pyrret.com/rays/rayserve"
)

func initEnv(wd string) (env string, err error) {
	Log("INFO", "performing inital enviroument checks")
	_, err = os.Stat(filepath.Join(wd, "./.git"))
	if err != nil {
		return "", errors.New("The working directory does not appear to be tracked in a git repository, which is required for ray deployment and development. You can use git init to start a new repository.")
	}

	if _, err := exec.LookPath("git"); err != nil {
		return "", errors.New("Could not find git, make sure it's installed.")
	}

	_, err = os.Stat(filepath.Join(wd, "./ray.config.json"))
	if err != nil {
		return "", errors.New("Could not find ray.config.json in working directory. See the docs for writing that file.")
	}

	Log("INFO", "removing old .raydev contents")
	err = os.RemoveAll(filepath.Join(wd, "./.raydev"))
	if err != nil {
		return "", err
	}

	err = os.Mkdir(filepath.Join(wd, "./.raydev"), 0666)
	if err != nil {
		return "", err
	}

	Log("INFO", "Cloning directory")
	cmd := exec.Command("git", "clone", wd, filepath.Join(wd, "./.raydev"))
	cmd.Stdout = &LogWriter{
		Style: "INFO",
	}
	cmd.Stderr = &LogWriter{
		Style: "ERR",
	}

	err = cmd.Run()
	if err != nil {
		return "", errors.New("Could not clone repository: " + err.Error())
	}

	return filepath.Join(wd, "./.raydev"), nil
}

func pickPort() (string, error) {
	ports := []string{"7292", "3000", "8080", "3001", "5678", "5000", "8000"} //7292 for telephone keypad -> RAYC
	for _, v := range ports {
		l, err := net.Listen("tcp", ":" + v)
		if err != nil {
			Log("INFO", "port " + v + "busy: " + err.Error() + ", trying another port")
			continue
		}

		l.Close()
		Log("INFO", "port " + v + " ok")
		return v, nil
	}

	return "", errors.New("all ports busy")
}

func deployEnv(env string, cmd *cli.Command) (error, string) {
	ba, err := os.ReadFile("ray.config.json")
	if err != nil {
		return err, ""
	}

	config, err := prjcnf.TranslateAndMarshalConfig(ba)
	if err != nil {
		return err, ""
	}

	port := cmd.String("port")
	if !config.NonNetworked && port == "" {
		port, err = pickPort()
		if err != nil {
			return err, ""
		}
	}

	msg := prjcnf.ValidateProjectConfig(config)
	if msg == "" {
		Log("INFO", "validated project config, everything alright.")
	} else {
		return errors.New("The static validator detected a problem with your project config: " + msg), ""
	}

	Log("INFO", "now going through pipeline.")
	for stepIndex, step := range config.Pipeline {
		commandDir := env
		if step.Options.Dir != "" {
			commandDir = filepath.Join(commandDir, step.Options.Dir)
		}

		if step.Tool == "rayserve" {
			if step.Type == "build" {
				return errors.New("Rayserve cannot be used with step type build"), ""
			} else if config.NonNetworked {
				return errors.New("Rayserve cannot be used with NonNetworked configured"), ""
			}

			notFoundPage, err := os.ReadFile(filepath.Join(commandDir, "404.html"))
			if err != nil {
				Log("INFO", "No 404 page specified for rayserve")
				notFoundPage = []byte("Rayserve: 404 page not found")
			}
			
			handler := rayserve.RayserveFileServer(commandDir, notFoundPage, step.Options.RayserveDisableDirListing, step.Options.RayserveRedirects, "rayc dev " + Version)
			go func ()  {
				Log("ERR", http.ListenAndServe(":" + port, handler))
			}()
			break
		}

		if step.Options.IfAvailable {
			_, errLocal := os.Stat(filepath.Join(commandDir, step.Tool))
			_, err := exec.LookPath(step.Tool)
			if err != nil && errLocal != nil {
				Log("INFO", "Command " + step.Tool + " is not available on this system. Skipping...")
				continue
			}
		}

		cmd := exec.Command(step.Tool, step.Command...)
		cmd.Dir = commandDir
		cmd.Env = cmd.Environ()

		for field, val := range step.Options.EnvVar {
			cmd.Env = append(cmd.Env, field+"="+val)
		}

		if !config.NonNetworked {
			cmd.Env = append(cmd.Env, "ray-port="+port)
			cmd.Env = append(cmd.Env, "RAY_PORT="+port)
		}

		cmd.Stdout = &LogWriter{
			Style: "INFO",
		}
		cmd.Stderr = &LogWriter{
			Style: "ERR",
		}

		commandError := cmd.Start()
		deployProcessExited := false

		if step.Type == "build" && commandError == nil {
			commandError = cmd.Wait()
		} else if commandError == nil {
			go func() { //if the deploy process exits within 2000ms so we can check for it later, otherwise this goroutinue will keep running and doing nothing (really hacky)
				cmd.Wait()
				deployProcessExited = true
			}()
			time.Sleep(2000 * time.Millisecond)
		}

		//in case the step is of type build, commandError will be non nil if the os couldn't run the command or if the command errored
		//in case the step is of type deploy, commandError will be non nil if the os couldn't run the command, and deployProcessExited true if it exited withing 2100ms
		if commandError != nil || (step.Type == "deploy" && deployProcessExited) {
			if commandError != nil && strings.Contains(commandError.Error(), exec.ErrNotFound.Error()) {
				return errors.New("failed to deploy: the tool '" + step.Tool + "' used in the deployment pipeline may not be installed. Please install it and make it accessible through PATH."), ""
			}

			return errors.New("failed to deploy, step " + strconv.Itoa((stepIndex + 1)) + ": is there an issue with your command or code?"), ""
		} else {
			Log("INFO", "Completed step "+  strconv.Itoa((stepIndex+1)) + ", " + step.Tool + " (" + strconv.Itoa(int((float32((stepIndex+1))/float32(len(config.Pipeline)))*100))+"%) ("+step.Type+")")
			if step.Type == "deploy" {
				break
			}
		}
	}
	return nil, port
}

var doneStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("#0dbc79")).Padding(1,4)


func dev(cc context.Context, cmd *cli.Command) error {
	wd := cmd.StringArg("directory")
	if wd == "" {
		wd = "./"
	}

	env, err := initEnv(wd)
	if err != nil {
		return err
	}

	err, port := deployEnv(env, cmd)
	if err != nil {
		return err
	}

	fmt.Println(doneStyle.Render("Running!","- Local: http://ray.localhost:" + port))
	select{}
}