package main

//Docker compatability mode (DWR)

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pyrret.com/rays/prjcnf"
)

func deployLocalDockerProcess(project *project, swapfunction *func(), branch string, branchHash string, logDir string, envDir string, procId string, RLSHost string) {
	rlog.BuildNotify("Attempting to launch "+project.Name+" (deployment " + branch  + ") using Docker", "info")
	var process process

	process.Branch = branch
	process.Hash = branchHash
	process.Id = procId
	process.RLSInfo.IP = RLSHost
	process.Project = project
	process.Env = envDir
	process.Active = true
	process.State = "OK"
	process.ProjectConfig = &prjcnf.ProjectConfig{}
	process.remove = func() {
		rlog.Println("Did not remove process, process never started.")
	}
	if RLSHost == "127.0.0.1" {
		process.RLSInfo.Type = "local"
	} else {
		process.RLSInfo.Type = "adm"
	}

	logPath := filepath.Join(logDir, "log-" + getUuid() + ".json")
	var logFile logFile
	process.LogFile = logPath

	//step 0: config validation and preperation

	if !project.DockerOptions.NonNetworked {
		process.Port = pickPort()
	}

	//pull image
	rlog.BuildNotify("Now pulling container image '"+project.Src+"' for " + project.Name, "info")
	var pullLogBuffer strings.Builder
	pull := exec.Command("docker", "pull", project.Src)
	pull.Stdout = &pullLogBuffer
	pull.Stderr = &pullLogBuffer

	pullError := pull.Run()
	finishLogSection(&pullLogBuffer, &logFile, -1, prjcnf.PipelineStep{Tool: "Pull container image"}, pullError == nil)
	if pullError != nil {
		process.Active = false
		process.State = pullLogBuffer.String() + ", " + pullError.Error()
		rlog.BuildNotify("Failed pulling image '"+project.Src+" for " + project.Name, "err")
		rlog.BuildNotify(pullLogBuffer.String(), "err")
		rlog.BuildNotify("OS Error:" + pullError.Error(), "err")

		finishProcess(logFile, &process, *project, branch, logPath)
		processes = append(processes, &process)
		return
	} else {
		rlog.BuildNotify("Successfully pulled image for " + project.Name, "done")
	}


	//add files
	var addFilesLogBuffer strings.Builder
	addFilesLogBuffer.WriteString("Adding files...\n")
	err := AddFiles(project.Files, envDir, &addFilesLogBuffer)
	if err != nil {
		addFilesLogBuffer.WriteString("Failed adding files: " + err.Error() + "\n")
		process.Active = false
		process.State = err.Error()
		finishProcess(logFile, &process, *project, branch, logPath)
		processes = append(processes, &process)
	} else {
		addFilesLogBuffer.WriteString("All files to be added have been added.")
	}
	finishLogSection(&addFilesLogBuffer, &logFile, -1, prjcnf.PipelineStep{Tool: "Add files"}, err == nil)
	if err != nil {return}

	var deployLogBuffer strings.Builder 

	envs := project.EnvVars
	if envs == nil {
		envs = map[string]string{}
	}
	envs["RAY_DEPLOYMENT"] = branch //since branches dont exist for containers and maintaining one container for each channel would be cumbersome, this is what differs channels for DWR

	containerName := "ray-" + getUuid()
	args := []string{"run", "-i", "--rm"}
	for key := range envs {
		args = append(args, "-e", key)
	}

	if !project.DockerOptions.NonNetworked {
		args = append(args, "-p", strconv.Itoa(process.Port) + ":" + strconv.Itoa(project.DockerOptions.ContainerPort))
	}

	for source, dest := range project.DockerOptions.Volumes {
		args = append(args, "-v", filepath.Join(envDir, source) + ":" + dest)
	}

	args = append(args, "--init", "--name", containerName)
	args = append(args, project.Src)

	cmd := exec.Command("docker", args...)
	cmd.Env = cmd.Environ()
	for key, val := range envs {
		cmd.Env = append(cmd.Env, key + "=" + val)
	}

	cmd.Stdout = &deployLogBuffer
	cmd.Stderr = &deployLogBuffer

	process.remove = func() {
		makeGhost(&process)
		stopCmd := exec.Command("docker", "stop", containerName) //docker stop sends SIGTERM, waits for up to 10 seconds, and then SIGKILL, so this wont hang for more than 10 sec
		ba, err := stopCmd.CombinedOutput()
		if err != nil {
			rlog.Notify("Process kill error: ", "err")
			rlog.Notify(string(ba), "err")
		}
	}
		

	//TODO: write to a file for deploy steps, keeping everything the program logs in a buffer in memory is a terrible idea.
	cmd.Stdout = &deployLogBuffer
	cmd.Stderr = &deployLogBuffer

	(*swapfunction)()

	commandError := cmd.Start()
	deployProcessExited := false

	if commandError == nil {
		time.Sleep(2000 * time.Millisecond)
		go func() { //if the deploy process exits within 2100ms so we can check for it later, otherwise this goroutinue will keep running and do nothing (really hacky)
			cmd.Wait()
			deployProcessExited = true
		}()
		time.Sleep(100 * time.Millisecond)
	}

	buildErr := func (message string)  {
		rlog.BuildNotify(message, "err")
		deployLogBuffer.Write([]byte(message + "\n"))
	}

	//commandError will be non nil if the os couldn't run the command, and deployProcessExited true if it exited withing 2100ms
	if commandError != nil || deployProcessExited {
		if commandError != nil {
			buildErr("Failed to deploy " + project.Name + " (branch " + branch + "): the OS encountered an error starting the container.")
			process.State = "DockerOSError"
		} else { //deployProcessExited
			lbString := deployLogBuffer.String()
			if lbString == "" {
				lbString = "(no output)"
			}

			buildErr("Failed to deploy and start container for " + project.Name + " (branch " + branch + "), is there an issue with the container?")
			process.State = deployLogBuffer.String()
			rlog.BuildNotify("Output:", "err")
			rlog.BuildNotify(lbString, "err")
		}
		process.Active = false
		finishLogSection(&deployLogBuffer, &logFile, 0, prjcnf.PipelineStep{Tool: "Running container (deploy step)"}, false)
	} else {
		rlog.BuildNotify("Successfully started container '" + project.Src + "' for " +project.Name+" (deployment " + branch  + ")", "done")

		process.Processes = append(process.Processes, cmd.Process.Pid)
		process.log = &deployLogBuffer
		go trackProcess(cmd, &process, &deployLogBuffer)
			//go waitForProcessListen(&process, "DOCKER", true) //maybe not use this for now
	}
	
	finishProcess(logFile, &process, *project, branch, logPath)
	processes = append(processes, &process)
}