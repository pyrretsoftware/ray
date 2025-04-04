package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type cliCommand struct {
	Command string
	Args    []string
}


func daemonHandleCommand(command cliCommand) []byte {
	switch command.Command {
	case "LISTPROCESS":
		json, err := json.Marshal(processes)
		if err != nil {
			log.Println(err)
		}

		return append(json, byte('\n'))
	case "RELOAD":
		config := readConfig()
		rconf = &config
		for _, project := range rconf.Projects {
			startProject(&project, rdata.RayEnv)
		}
		return []byte("success\n")
	case "FORCE_RE":
		rconf.ForcedRenrollment = time.Now().Unix()
		err := applyChanges(*rconf)

		var data string
		if (err == nil) {
			data = ""
		} else {
			data = err.Error()
		}
		return []byte(data + "\n")
	case "STOP":
		rlog.Println("Exiting...")
		os.Exit(0)
		return []byte("\n")
	case "GETDEVAUTH":
		generateAuth()
		json, err := json.Marshal(devAuth)
		if err != nil {
			log.Println(err)
		}

		return append(json, byte('\n'))
	default:
		return []byte("\n")
	}
}

func cliSendCommand(command string, args []string) []byte {
	socketPath := dotslash + "/clisocket.sock"

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("Failed connecting to rays, it may not have finished initalization or it may not be started at all.")
		log.Fatal(err.Error())
	}
	defer conn.Close()

	var jsonCommand cliCommand
	jsonCommand.Command = command
	jsonCommand.Args = args

	jsonData, err := json.Marshal(jsonCommand)
	if err != nil {
		log.Fatal(err)
	}
	jsonData = append(jsonData, byte('\n'))

	_, err = conn.Write(jsonData)
	if err != nil {
		log.Fatal("Failed to send command: " + err.Error())
	}

	if (command == "STOP") {return []byte("\n")}
	buffer := make([]byte, 4096)
	_command := make([]byte, 0)
	for {
		n, _err := conn.Read(buffer)
		if _err != nil {
			log.Println("Error reading from connection:", _err)
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
		log.Println("Warning: could not remove existing socket file:", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Invalid cli connection: " + err.Error())
			continue
		}

		buffer := make([]byte, 4096)
		_command := make([]byte, 0)

		for {
			n, _err := conn.Read(buffer)
			if _err != nil {
				log.Println("Error reading from connection:", _err)
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
			log.Println("Issue decoding CLI command: " + __err.Error())
			log.Println(string(_command))
			continue
		}

		conn.Write(daemonHandleCommand(command))
	}
}
