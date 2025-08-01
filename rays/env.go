package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)
var exiting = false
var processes []*process

func trackProcess(cmd *exec.Cmd, process *process, buffer *strings.Builder) {
	err := cmd.Wait()
	if exiting || process.State == "drop" {return}
	if (err != nil) {
		rlog.Notify("Process errored: ", "err")
		rlog.Notify(buffer.String(), "err")
		process.State = buffer.String()
		go triggerEvent("processError", *process)
		go taskAutofix(*process)
	} else {
		rlog.Println("Process exited.")
		process.State = "Exited without error."
		go triggerEvent("processError", *process)
		go taskAutofix(*process)
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
	for (len(ports) == 0 && waited < 100) {
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

func startUpdateCheck() {
	for {
		time.Sleep(time.Minute)
		updateProjects(false)
	}
}

func updateProjects(updateRollbacks bool) {//we wont print anything if no updates are found, as to not fill up log files
	for _, project := range rconf.Projects {
		branches := getBranches(project.Src)
		doUpdate := false

		dplymnt := project.Deployments
		dplymnt = append(dplymnt, deployment{
			Type: "prod",
			Branch: "prod", //removed this once because me stupid, NOT AGAIN!!!!!!!1
		})

		for _, deployment := range dplymnt {
			if (branches == nil || branches[deployment.Branch] == "") {
				continue
			}

			process := getProcessFromBranch(deployment.Branch, project)
			if (process == nil) {continue}

			//if we're rolled back and not on the faulty version
			if strings.HasPrefix(process.Hash, "rollback:") && strings.Replace(process.Hash, "rollback:", "", 1) == branches[deployment.Branch] && !updateRollbacks {
				continue
			}

			if (branches[deployment.Branch] != process.Hash) {
				doUpdate = true
			}
		}
		if (doUpdate) {
			rlog.Println("Performing update on " + project.Name)
			startProject(&project, "")
		}
	}
}

func finishLogSection(logBuffer *strings.Builder, file *logFile, si int, step pipelineStep, success bool) {
	logBuffer.WriteString("\nFinishing logging for this step\n")
	file.Steps = append(file.Steps, logSection{
		Name: step.Tool + " (Step " + strconv.Itoa(si + 1) + ")",
		Success: success,
		Log: logBuffer.String(),
	})
}

func finishProcess(logFile logFile, process process, project project, branch string, logPath string) {
	logFile.Name = project.Name + " (branch " + branch + ")"
	if (process.Active && process.State == "OK") {
		rlog.Notify(project.Name + ", branch " + branch + " was sucessfully deployed!", "done")
		logFile.Success = true
		go triggerEvent("newProcess", process)
	} else {
		rlog.Notify(project.Name + ", branch " + branch + " was not sucessfully deployed!", "err")
		logFile.Success = false
		go triggerEvent("processError", process)
		go taskAutofix(process)
	}
	logB, err := json.MarshalIndent(logFile, "", "    ")
	if err != nil {
		rlog.Println("Failed encoding log file.")
	} else {
		err := os.WriteFile(logPath, logB, 0600)
		rerr.Notify("Failed writing log file.", err)
	}
}

func deployLocalProcess(configPath string, dir string, project *project, swapfunction *func(), branch string, branchHash string, logDir string, envDir string, procId string, RLSHost string) {
	rlog.BuildNotify("Attempting to launch " + project.Name + " (deployment " + branch + ")", "info")
	var config projectConfig
	var process process

	process.Branch = branch
	process.Hash = branchHash
	process.Id = procId
	process.RLSInfo.IP = RLSHost
	process.Project = project
	process.Env = envDir
	process.Active = true
	process.State = "OK"
	if RLSHost == "127.0.0.1" {
		process.RLSInfo.Type = "local"
	} else {
		process.RLSInfo.Type = "adm"
	}

	logPath := path.Join(logDir, "log-" + getUuid() + ".json")
	var logFile logFile 
	process.LogFile = logPath

	
	//step 0: config validation and preperation
	var stepZeroLogBuffer strings.Builder
	stepZeroSuccess := false
	;func () {
		if _, err := os.Stat(configPath); err != nil {
			stepZeroLogBuffer.WriteString("No ray.config.json file found.")
			return
		}
		_config, err := os.ReadFile(configPath)
		if (err != nil) {
			stepZeroLogBuffer.WriteString("Could not read project config.")
			return
		}
		if err := json.Unmarshal(_config, &config); err != nil {
			stepZeroLogBuffer.WriteString("Failed parsing project config, json unmarshaling error: " + err.Error())
			return
		}
		
		verr := validateProjectConfig(config, *project)
		if verr != "" {
			stepZeroLogBuffer.WriteString("There is an issue with your project config: " + verr)
			return
		}
		stepZeroSuccess = true
		process.ProjectConfig = &config
	}();
	finishLogSection(&stepZeroLogBuffer, &logFile, -1, pipelineStep{Tool: "Initial preperation and validation"}, stepZeroSuccess)
	if !stepZeroSuccess {
		process.Active = false
		process.State = stepZeroLogBuffer.String()
		finishProcess(logFile, process, *project, branch, logPath)
		processes = append(processes, &process)
		return
	}

	if (!config.NotWebsite) {
		process.Port = pickPort()
	}
	
	for stepIndex, step := range config.Pipeline {
		var logBuffer strings.Builder //implements io.Writer

		BuiltIntool := builtIn(step)
		commandDir := dir
		if (step.Options.Dir != "") {
			commandDir = path.Join(commandDir, step.Options.Dir)
		}

		if (BuiltIntool == "rayserve" && !config.NotWebsite) {
			(*swapfunction)()
			staticServer(commandDir, process.Port, &process, step.Options.RayserveRedirects, step.Options.RayserveDisableDirListing)
			finishLogSection(&logBuffer, &logFile, stepIndex, step, true)
			continue
		}

		if (step.Options.IfAvailable) {
			_, errLocal := os.Stat(path.Join(commandDir, step.Tool))
			_, err := exec.LookPath(step.Tool)
			if (err != nil && errLocal != nil) {
				rlog.Println("Command " + step.Tool + " is not available on this system. Skipping...")
				logBuffer.WriteString("Command " + step.Tool + " is not available on this system. Skipping...\n")
				finishLogSection(&logBuffer, &logFile, stepIndex, step, true)
				continue
			}
		} 

		cmd := exec.Command(step.Tool, step.Command...)
		cmd.Dir = commandDir
		cmd.Env = cmd.Environ()

		process.remove = func() {
			makeGhost(&process)
			if cmd.Process != nil {
				err := cmd.Process.Kill()
				rerr.Notify("Process kill error: ", err, true)
			} else {
				rlog.Notify("Did not kill process, it did not exist in the first place.", "warn")
			}
		}

		for field, val := range project.EnvVars {
			cmd.Env = append(cmd.Env, field + "=" + val)
		}
		for field, val := range step.Options.EnvVar {
			cmd.Env = append(cmd.Env, field + "=" + val)
		}
		if (!config.NotWebsite) {
			cmd.Env = append(cmd.Env, "ray-port=" + strconv.Itoa(process.Port))
			cmd.Env = append(cmd.Env, "RAY_PORT=" + strconv.Itoa(process.Port))
		}
		cmd.Env = append(cmd.Env, "RAY_DEPLOYMENT=" + process.Branch)

		if (step.Type == "deploy") {
			(*swapfunction)()
		}

		cmd.Stdout = &logBuffer
		cmd.Stderr = &logBuffer

		commandError := cmd.Start()
		deployProcessExited := false

		if step.Type == "build" && commandError == nil {
			commandError = cmd.Wait()
			finishLogSection(&logBuffer, &logFile, stepIndex, step, commandError == nil)
		} else if commandError == nil {
			time.Sleep(2000 * time.Millisecond)
			go func () { //if the deploy process exits within 2100ms so we can check for it later, otherwise this goroutinue will keep running and doing nothing
				cmd.Wait()
				deployProcessExited = true
			}()
			time.Sleep(100 * time.Millisecond)
		}
		
		buildErr := func (message string)  {
			rlog.BuildNotify(message, "err")
			logBuffer.Write([]byte(message + "\n"))
		}

		//in case the step is of type build, commandError will be non nil if the os couldn't run the command or if the command errored
		//in case the step is of type deploy, commandError will be non nil if the os couldn't run the command, and deployProcessExited true if it exited withing 2100ms
		if (commandError != nil || (step.Type == "deploy" && deployProcessExited)) {
			if (commandError != nil && strings.Contains(commandError.Error(), exec.ErrNotFound.Error())) {
				buildErr("Failed to deploy " + project.Name + " (branch " + branch + "): the tool '" + step.Tool + "' used in the deployment pipeline may not be installed. Please install it and configure it in PATH.")
				process.State = "ToolNotFound"
			} else {
				lbString := logBuffer.String()
				if lbString == "" {
					lbString = "(no output)"
				}

				buildErr("Failed to deploy " + project.Name + " (branch " + branch +", step " + strconv.Itoa((stepIndex + 1)) + "), is there an issue with your command or code?")
				process.State = logBuffer.String()
				rlog.BuildNotify("Output:", "err")
				rlog.BuildNotify(lbString, "err")
			}
			process.Active = false
			finishLogSection(&logBuffer, &logFile, stepIndex, step, false)
			break
		} else {
			rlog.BuildNotify("Completed step " + strconv.Itoa((stepIndex + 1)) + ", " + step.Tool + " (" + strconv.Itoa(int((float32((stepIndex + 1)) / float32(len(config.Pipeline))) * 100)) + "%) (" + step.Type + ", deployment " + branch +")", "done") 
			if step.Type == "deploy" {
				process.Processes = append(process.Processes, cmd.Process.Pid)
				go trackProcess(cmd, &process, &logBuffer)

				if (config.NotWebsite) {break}
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

				break
			}
		}
	}

	finishProcess(logFile, process, *project, branch, logPath)
	processes = append(processes, &process)
}

func setupLocalProject(project *project, host string, hardCommit string) []process {
	validateDeployments(project.Deployments)

	var oldprocesses []*process
	for _, prc := range processes {
		if (prc.Project.Name == project.Name && !prc.Ghost && prc.RLSInfo.IP == host) {
			oldprocesses = append(oldprocesses, prc)
		}
	}

	var deployments = project.Deployments
	deployments = append(deployments, deployment{
		Type: "prod",
		Branch: "prod",
	})

	branchHashes := getBranches(project.Src)
	for _, deployment := range deployments {
		rlog.Println("Setting up enviroument for " + project.Name + " (deployment " + deployment.Branch + ")")

		procId := strings.ReplaceAll(project.Name, " ", "-") + "-" + getUuid()
		dir := rdata.RayEnv + "/" + procId
		os.Mkdir(dir, 0600)

		_cmd := []string{"clone", project.Src}
		if (deployment.Type != "prod") {
			_cmd = append(_cmd, "-b")
			_cmd = append(_cmd, deployment.Branch)
		}
		
		gitOutput := strings.Builder{}
		cmd := exec.Command("git", _cmd...)
		cmd.Dir = dir
		cmd.Stdout = &gitOutput
		cmd.Stderr = &gitOutput
		
		if err := cmd.Run(); err != nil {
			rlog.Println(_cmd)
			rlog.Println(gitOutput.String())
			rlog.Notify("Git cloning error: " + err.Error(), "err")
			continue //this wont fire any monitoring stuff and just silently fail (with the exception of the above log), could def be improved
		}
	
		content, err := os.ReadDir(dir)
		if err != nil {
			rlog.Fatal(err)
		}

		if hardCommit != "" {
			cmd := exec.Command("git", "reset", "--hard", hardCommit)
			cmd.Dir = path.Join(dir, content[0].Name())
			if err := cmd.Run(); err != nil {
				rlog.Println(cmd.Args)
				rlog.Notify("Git hard commit resetting error: " + err.Error(), "err")
				continue //this wont fire any monitoring stuff and just silently fail (with the exception of the above log), could def be improved
			}
		}

		os.Mkdir(path.Join(dir, "logs"), 0600)//Making sure to do this after we've cloned the repo
	
		projectConfig := dir + "/" + content[0].Name() + "/" + "ray.config.json" //the existance of this file is checked later dw
		rm := func() {
			for _, proc := range oldprocesses {
				if (!proc.Ghost) {
					proc.remove()
				}
			}
		}
	
		branch := deployment.Branch

		branchHash := ""
		if hardCommit != "" {
			branchHash = "rollback:" + branchHashes[branch] //the faulty version
		} else if (branchHashes != nil && branchHashes[branch] != "") {
			branchHash = branchHashes[branch]
		}
		go deployLocalProcess(projectConfig, dir + "/" + content[0].Name(), project, &rm, branch, branchHash, path.Join(dir, "logs"), dir, procId, host)
	}

	var newProcesses []process
	for _, process := range processes {
		if process.RLSInfo.IP == host {
			newProcesses = append(newProcesses, *process)
		}
	}
	return newProcesses
}

func startProject(project *project, hardCommit string) {
	if strings.Contains(hardCommit, "rollback:") {rlog.Notify("Cannot rollback to a rollback", "warn"); return}

	if slices.Contains(project.DeployOn, "local") {
		setupLocalProject(project, "127.0.0.1", hardCommit)
	}

	for _, deploymentTarget := range project.DeployOn {
		if deploymentTarget == "local" {continue}
		go setupRlspProject(project, deploymentTarget, hardCommit)
	}
}

var rconf *rayconfig
var rdata raydata
func SetupEnv() {
	os.Mkdir(dotslash, 0600)
	os.Mkdir(dotslash + "/projects", 0600)
	os.Mkdir(dotslash + "/ray-certs", 0600)

	rdata.RayEnv = dotslash + "/projects/ray-env-" + getUuid()
	os.Mkdir(rdata.RayEnv, 0600)
	go func() {
		chnl := make(chan os.Signal)
		signal.Notify(chnl, os.Interrupt)
		<- chnl

		rlog.Println("Cleaning up enviroument.")
		exiting = true
		os.RemoveAll(rdata.RayEnv)
		os.Remove(dotslash + "/clisocket.sock")
		os.Exit(0)
	}()	

	for _, project := range rconf.Projects {
		startProject(&project, "")
	}

	RLSinitalConnectionOver = true
	validateConfig(*rconf)
	go startUpdateCheck()
}