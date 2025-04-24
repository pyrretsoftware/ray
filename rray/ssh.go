package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"golang.org/x/crypto/ssh"
)

func getOutputSpin(command string, remote string) string {
	var out string
	action := func ()  {
		out = getOutput(command, remote)
	}
	spinner.New().Title("Contacting server...").Action(action).Run()

	return out
}

func getOutput(command string, remote string) string {
	authMethods := getAuthMethods(remote)

	config := &ssh.ClientConfig{
		User: hostsFile.StoredHosts[remote].User,
		Auth: authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			var keyChanged = false
			for _, host := range hostsFile.KnownHosts {
				if host.Host == hostname && base64.StdEncoding.EncodeToString(key.Marshal()) == host.Key && key.Type() == host.KeyType {
					return nil
				} else if (host.Host == hostname) {
					keyChanged = true
				}
			}
			notice := "The public key provided by the server is not in your list of known hosts, and so the authenticity cannot be verified."
			if (keyChanged) {
				notice = warning.Render("Warning: Remote host identification has changed, and the authenticity of the server cannot be verified.")
			}
			if confirm(notice + "\nDo you want to continue?") {
				hostsFile.KnownHosts = append(hostsFile.KnownHosts, KnownHost{
					Host: hostname,
					KeyType: key.Type(),
					Key: base64.StdEncoding.EncodeToString(key.Marshal()),
				})
				fmt.Println("Alright, added to list of known hosts.")

				writeHostsFile()
				return nil
			} else {
				os.Exit(0)
				return errors.New("Cancelled by user")
			}
		},
	}
	client, err := ssh.Dial("tcp", hostsFile.StoredHosts[remote].Host, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatal("request for terminal failed: ", err)
	}

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	session.Stderr = os.Stderr

	var outBuf bytes.Buffer
	go io.Copy(&outBuf, stdout)

	// Start the command
	err = session.Start(command)
	if err != nil {
		log.Fatal("Failed to start command: ", err)
	}

	// Send the password followed by a newline
	_, err = stdin.Write([]byte(hostsFile.StoredHosts[remote].Password + "\n"))
	if err != nil {
		log.Fatal("Failed to write password: ", err)
	}

	err = session.Wait()
	if err != nil {
		log.Fatal("Command finished with error: ", err)
	}

	return strings.Join(strings.Split(outBuf.String(), "\n")[1:], "\n")
}