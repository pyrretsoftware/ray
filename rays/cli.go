package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"time"
)

type cliCommand struct {
	Command string
	Args    []string
}

func formatProcessList(process process) string {
	var state string
	if process.Ghost {
		state = " 👻"
	} else if (process.Active) {
		state = " ✅ (" + process.State + ")" 
	} else {
		state = " ❌ (error)"
	}
	var content = process.Project.Name + state + ". " + strconv.Itoa(len(process.Processes)) + " active processes."

	for indx, process := range process.Processes {
		content += "\n- PID (process " + strconv.Itoa(indx + 1) + "): "+ strconv.Itoa(process)
	}

	content += "\n- Internal Port: "+ strconv.Itoa(process.Port)
	content += "\n- RLS Type: "+ process.RLSInfo.Type
	content += "\n- Log file: "+ process.LogFile
	content += "\n- Enviroument: "+ process.Env
	content += "\n- Hash: "+ process.Hash
	content += "\n- Deployment: "+ process.Branch
	content += "\n- ID (Used for RLS): "+ process.Id

	return content
}

func RrayFormat(json []byte) {
	if len(os.Args) > 2 && os.Args[2] == "rray" {
		fmt.Println(string(json))
		os.Exit(0)
	}
}

func removeComponent(dir string) {
	if comp, err := os.Stat(dir); err == nil {
		rlog.Println("Removing " + dir)

		rmfunc := os.Remove
		if (comp.IsDir()) {
			rmfunc = os.RemoveAll
		}
		err := rmfunc(dir)
		rerr.Fatal("Couldn't remove component: ", err, true)
	} else {
		rlog.Notify("Component " + dir +" not found, ray might not have been properly installed.", "warn")
	}
}

func handleCommand(args []string) {
	switch (args[1]) {
	case "uninstall":
		rlog.Println("Uninstalling rays")
		fileEnding := ""
		if (runtime.GOOS == "windows") {
			fileEnding = ".exe"
		}

		installLocation := "/usr/bin"
		if runtime.GOOS == "windows" {
			dir, err := os.UserHomeDir()
			rerr.Fatal("", err, true)
			installLocation = dir
		}
	
		if (runtime.GOOS == "linux") {
			exec.Command("systemctl", "stop", "rays").Run()
			removeComponent("/etc/systemd/system/rays.service")
		}
		removeComponent(path.Join(installLocation, "rays" + fileEnding))
		removeComponent(path.Join(installLocation, "ray-env"))
		
	case "rray-edit-config":
		if len(os.Args) > 2 {
			ba, err := base64.StdEncoding.DecodeString(os.Args[2])
			rerr.Fatal("Invalid b64 config string: ", err, true)

			applyChangesRaw(ba)
		} else {
			rlog.Fatal("No b64 config provided.")
		}
	case "version":
		fmt.Println(Version)
		os.Exit(0)
	case "rray-read-config":
		fmt.Println(string(readConfigRaw()))
	case "list":
		jsonRes := cliSendCommand("LISTPROCESS", nil)
		RrayFormat(jsonRes)

		var response []process
		err := json.Unmarshal(jsonRes, &response)
		if (err != nil) {
			rlog.Fatal(err)
		}

		for _, process := range response {
			fmt.Println(formatProcessList(process))
		}
	case "reload":
		rlog.Println("Updating config file and restarting processes")
		data := cliSendCommand("RELOAD", nil)
		if (string(data) == "success\n") {
			rlog.Notify("Reloaded successfully.", "done")
			os.Exit(0)
		} else {
			rlog.Println("Rays did not indicate success in reloading processes, please check the status of the server.")
			os.Exit(1)
		}
	case "update":
		rlog.Println("Updating config file and restarting necessary processes")
		data := cliSendCommand("UPDATE", nil)
		if (string(data) == "success\n") {
			rlog.Notify("Reloaded successfully.", "done")
			os.Exit(0)
		} else {
			rlog.Println("Rays did not indicate success in reloading processes, please check the status of the server.")
			os.Exit(1)
		}
	case "force-renrollment":
		rlog.Println("Forcing a renrollment onto all users who were enrolled into a channel before this point...")
		data := cliSendCommand("FORCE_RE", nil)
		if (string(data) == "\n") {
			rlog.Notify("Applied changed to config.", "done")
		} else {
			rlog.Fatal("Failed applying changes to config.")
		}
	case "exit":
		rlog.Println("Exiting...")
		data := cliSendCommand("EXIT", nil)

		if (string(data) == "\n") {
			rlog.Notify("Exited!", "done")
		}
		os.Exit(0)
	case "dev-auth":
		if !*flag.Bool("rray", false, "use rray formatting") {
			rlog.Println("Generating new credentials for development channels... (all old credentials will be invalidated)")
		}
		
		jsonRes := cliSendCommand("GETDEVAUTH", nil)
		RrayFormat(jsonRes)

		var response auth
		err := json.Unmarshal(jsonRes, &response)
		if (err != nil) {
			rlog.Fatal(err)
		}

		rlog.Notify("Success!", "done")
		rlog.Println("Token: " + response.Token)
	case "edit-config":
		nano := exec.Command("nano", path.Join(dotslash, "rayconfig.json"))
		nano.Stdout = os.Stdout
		nano.Stderr = os.Stderr
		nano.Stdin = os.Stdin

		err := nano.Run()
		if (err == exec.ErrNotFound) {
			rlog.Println("Tried opening config file with nano, but it dosen't appear to be installed. Please install it or open the following file with another editor:")
			rlog.Println(path.Join(dotslash, "rayconfig.json"))
		} else {
			rlog.Notify("Exited nano without error", "done")
		}
	default:
		rlog.Fatal("Command not found")
	}
}

func daemonHandleCommand(command cliCommand) []byte {
	switch command.Command {
	case "LISTPROCESS":
		json, err := json.Marshal(processes)
		rerr.Notify("Failed marshaling json: ", err, true)

		return append(json, byte('\n'))
	case "RELOAD":
		config := readConfig()
		rconf = &config
		
		for _, project := range rconf.Projects {
			startProject(&project, "")
		}
		return []byte("success\n")
	case "UPDATE":
		config := readConfig()
		rconf = &config
		
		updateProjects(true)
		return []byte("success\n")
	case "FORCE_RE":
		rconf.ForcedRenrollment = time.Now().Unix()
		err := applyChanges(*rconf)

		var data string
		if (err == nil) {
			return []byte("")
		} else {
			data = err.Error()
		}
		return []byte(data + "\n")
	case "EXIT":
		rlog.Println("Exiting...")
		os.Exit(0)
		return []byte("\n")
	case "GETDEVAUTH":
		generateAuth()
		json, err := json.Marshal(devAuth)
		rerr.Notify("Failed marshaling json: ", err, true)

		return append(json, byte('\n'))
	default:
		return []byte("\n")
	}
}

func cliSendCommand(command string, args []string) []byte {
	socketPath := dotslash + "/clisocket.sock"

	conn, err := net.Dial("unix", socketPath)
	rerr.Notify("Failed connecting to rays, it may not have finished initalization or it may not be started at all.\nTip: try running with sudo/with elevated permissions", err)
	
	defer conn.Close()

	var jsonCommand cliCommand
	jsonCommand.Command = command
	jsonCommand.Args = args

	jsonData, err := json.Marshal(jsonCommand)
	rerr.Fatal("Failed marshaling json: ", err, true)
	jsonData = append(jsonData, byte('\n'))

	_, err = conn.Write(jsonData)
	rerr.Fatal("Failed to send command: ", err, true)

	if (command == "EXIT") {return []byte("\n")}
	buffer := make([]byte, 4096)
	_command := make([]byte, 0)
	for {
		n, _err := conn.Read(buffer)
		if _err != nil {
			rlog.Println("Error reading from connection:" + _err.Error())
			break
		}

		_command = append(_command, buffer[:n]...)

		if n > 0 && buffer[n-1] == '\n' {
			break
		}
	}

	return _command
}

func daemonListen() {
	socketPath := dotslash + "/clisocket.sock"

	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		rlog.Notify("Could not remove existing socket file:" + err.Error(), "warn")
	}

	listener, err := net.Listen("unix", socketPath)
	rerr.Fatal("Failed listenting to cli socket", err, true)

	for {
		conn, err := listener.Accept()
		if err != nil {
			rlog.Notify("Invalid cli connection: " + err.Error(), "err")
			continue
		}

		buffer := make([]byte, 4096)
		_command := make([]byte, 0)

		for {
			n, _err := conn.Read(buffer)
			if _err != nil {
				rlog.Notify("Error reading from connection:" + _err.Error(), "err")
				break
			}

			_command = append(_command, buffer[:n]...)

			if n > 0 && buffer[n-1] == '\n' {
				break
			}
		}

		var command cliCommand
		__err := json.Unmarshal(_command, &command)
		if __err != nil {
			rlog.Notify("Issue decoding CLI command: " + __err.Error(), "err")
			rlog.Notify(string(_command), "err")
			continue
		}

		conn.Write(daemonHandleCommand(command))
	}
}
