package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"time"
)
var exiting = false
var processes []*process

func pickPort() int {
	return rand.IntN(16383) + 49152
}

func trackProcess(cmd *exec.Cmd, process *process, stderr *io.ReadCloser) {
	err := cmd.Wait()
	if exiting || process.State == "drop" {return}
	if (err != nil) {
		rlog.Println("Process errored: ")
		slurp, _ := io.ReadAll(*stderr)
		rlog.Println(string(slurp))
		process.State = string(slurp)
	} else {
		rlog.Println("Process exited.")
		process.State = "Exited"
	}
	process.Active = false
}

func incorrectPortUsage(process *process, actualPort string) {
	process.remove()
	rlog.Notify(process.Project.Name + " was instructed to listen for connections on port " + strconv.Itoa(process.Port) + ", but instead started listenting on port " + actualPort + ", and was forcefully terminated.", "err")
	rlog.Notify("Please make sure your application listens to connections according to the ray-port enviroument variable.", "err")
}

func waitForPortOpen(process *process) {
	waited := 0
	var ports []string
	for (len(ports) > 0 || waited >= 100) {
		time.Sleep(500 * time.Millisecond)
		waited += 1
		ports = getProcessPorts(process.Processes[0])
	}

	if (len(ports) == 0) {
		rlog.Notify("The application has not yet started listening for connections on any port, even after waiting 50 seconds, Terminating...", "warn")
		process.remove()
		return
	}

	if (ports[0] != strconv.Itoa(process.Port)) {
		incorrectPortUsage(process, ports[0])
		return
	}
}

func launchProject(configPath string, dir string, project *project, swapfunction *func(), branch string) {
	rlog.BuildNotify("Attempting to launch " + project.Name + " (deployment " + branch + ")", "info")
	_config, err := os.ReadFile(configPath)
	if (err != nil) {
		rlog.Fatal(err)
	}

	var config projectConfig
	var process process
	process.Branch = branch
	if err := json.Unmarshal(_config, &config); err != nil {
		rlog.Fatal(err)
	}

	process.Project = project
	process.Active = true
	process.State = "OK"
	process.Port = pickPort()
	
	for stepIndex, step := range config.Pipeline {
		if (step.Tool == "rayserve" && step.Type == "deploy") {
			(*swapfunction)()
			staticServer(dir, process.Port, &process)

			continue
		} else if (step.Tool == "rayserve") {
			rlog.Notify("ray.config.json error: rayserve is a built in ray tool that requires type deploy.", "err")
		}

		cmd := exec.Command(step.Tool, step.Command)
		cmd.Dir = dir
		cmd.Env = cmd.Environ()
		stderr, _ := cmd.StderrPipe()
		if (err != nil) {
			log.Panic("error getting stderr.")
		}

		process.remove = func() {
			process.State = "drop"
			process.Ghost = true
			err := cmd.Process.Kill()
			if err != nil {
				rlog.Notify("Process kill error " + err.Error(), "err")
			}
		}

		for field, val := range project.EnvVars {
			cmd.Env = append(cmd.Env, field + "=" + val)
		}
		cmd.Env = append(cmd.Env, "ray-port=" + strconv.Itoa(process.Port))
		if (step.Type == "deploy") {
			(*swapfunction)()
		}

		err := cmd.Start()
		if step.Type == "build" {
			cmd.Wait()
		} else {
			time.Sleep(1000 * time.Millisecond)
		}

		if (err != nil) {
			if (strings.Contains(err.Error(), exec.ErrNotFound.Error())) {
				rlog.BuildNotify("Failed to deploy " + project.Name + " (branch " + branch + "): the tool '" + step.Tool + "' used in the deployment pipeline may not be installed. Please install it and configure it in PATH.", "err")
			} else {
				rlog.BuildNotify("Failed to deploy " + project.Name + " (branch " + branch +"), is there and issue with your command?", "err")
				rlog.BuildNotify("", "err")
				rlog.BuildNotify(cmd.Stdout, "err")
			}
			process.Active = false
			process.State = err.Error()
		} else {
			rlog.BuildNotify("Completed step " + strconv.Itoa((stepIndex + 1)) + ", " + step.Tool + " (" + strconv.Itoa(int((float32((stepIndex + 1)) / float32(len(config.Pipeline))) * 100)) + "%) (" + step.Type + ", deployment " + branch +")", "done") 
			if step.Type == "deploy" {
				process.Processes = append(process.Processes, cmd.Process.Pid)
				go trackProcess(cmd, &process, &stderr)

				ports := getProcessPorts(cmd.Process.Pid)
				if (len(ports) > 0) {
					if (!slices.Contains(ports, strconv.Itoa(process.Port))) {
						incorrectPortUsage(&process, ports[0])
					} else {
						rlog.Notify("Process correctly listens to connections from the instructed port.", "done")
					}
				} else {
					rlog.Println("Process has not yet started listening for connections.")
					go waitForPortOpen(&process)
				}
			}
		}
	}
	if (process.Active && process.State == "OK") {
		rlog.Notify(project.Name + ", branch " + branch + " was sucessfully deployed!", "done")
	} else {
		rlog.Notify(project.Name + ", branch " + branch + " was not sucessfully deployed!", "err")
	}

	processes = append(processes, &process)
}

func startProject(project *project, env string) {
	var oldprocesses []*process
	for _, prc := range processes {
		if (prc.Project.Name == project.Name && !prc.Ghost) {
			oldprocesses = append(oldprocesses, prc)
		}
	}

	var deployments = project.Deployments
	deployments = append(deployments, deployment{
		Type: "prod",
	})

	for _, deployment := range deployments {
		var _dpl = deployment.Branch
		if (_dpl == "") {
			_dpl = "prod"
		}
		rlog.Println("Setting up enviroument for " + project.Name + " (deployment " + _dpl + ")")

		dir := env + "/" + project.Name + "-" + strconv.Itoa(rand.IntN(10000000))
		os.Mkdir(dir, 0600)

		_cmd := []string{"clone", project.Src}
		if (deployment.Type != "prod") {
			_cmd = append(_cmd, "-b")
			_cmd = append(_cmd, deployment.Branch)
		}
		
		cmd := exec.Command("git", _cmd...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			rlog.Println(cmd.Args)
			rlog.Fatal("git cloning error: " + err.Error())
		}
	
		content, err := os.ReadDir(dir)
		if (err != nil) {
			rlog.Fatal(err)
		}
	
		projectConfig := dir + "/" + content[0].Name() + "/" + "ray.config.json"
		if _, err := os.Stat(projectConfig); err != nil {
			rlog.Fatal("No ray.config.json file found in project.")
		}
		rm := func() {
			for _, proc := range oldprocesses {
				if (!proc.Ghost) {
					proc.remove()
				}
			}
		}
	
		branch := deployment.Branch
		if branch == "" {
			branch = "prod"
		}
		launchProject(projectConfig, dir + "/" + content[0].Name(), project, &rm, branch)
	}
}

var rconf *rayconfig
var rdata raydata
func SetupEnv() {
	_cnf := readConfig()
	rconf = &_cnf
	
	rdata.RayEnv = rconf.EnvLocation + "/ray-env-" + strconv.Itoa(rand.IntN(10000000))
	os.Mkdir(rdata.RayEnv, 0600)
	go func() {
		chnl := make(chan os.Signal)
		signal.Notify(chnl, os.Interrupt)
		<- chnl

		rlog.Println("Cleaning up enviroument.")
		exiting = true
		os.RemoveAll(rdata.RayEnv)
		os.Exit(0)
	}()	

	for _, project := range rconf.Projects {
		startProject(&project, rdata.RayEnv)
	}
}