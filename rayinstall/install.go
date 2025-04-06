package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
)

func systemdRegisterDaemon(service string) {
	err := os.WriteFile("/etc/systemd/system/rays.service", []byte(service), 0644)

	if (err != nil){
		fmt.Println("Cant create systemd daemon:")
		fmt.Println(err)
	} else {
		fmt.Println("Successfully registered as a systemd daemon!")
	}
}

func registerDaemon(path string) {
	if (runtime.GOOS == "windows") {
		fmt.Println("Rays has not been daemonized since you're on Windows. If you want rays to automatically start on boot you will have to register it as a service manually.")
		return
	}

	daemon := strings.ReplaceAll(systemdService, "${BinaryPath}", path)
	cuser, err := user.Current()
	if (err != nil) {
		log.Fatal("Cant get current user: " + err.Error())
	}

	daemon = strings.ReplaceAll(daemon, "${User}", cuser.Username)
	systemdRegisterDaemon(daemon)	
}

func installPack(pack inPack) {
	if pack.Metadata.Platform != runtime.GOOS {
		log.Fatal("Installation package os dosen't match this OS")
	}

	installLocation := "/usr/bin"
	if runtime.GOOS == "windows" {
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		installLocation = dir
	}

	for _, file := range pack.Binaries {
		blob, err := base64.StdEncoding.DecodeString(file.Blob)
		if (err != nil) {
			log.Fatal(err)
		}
		os.WriteFile(path.Join(installLocation, file.Name), blob, 0667)
	}
	registerDaemon(path.Join(installLocation, "rays" + fileEnding))
}

var systemdService string = `[Unit]
Description=ray server (rays)
After=network.target

[Service]
User=${User}
Restart=always
ExecStart=${BinaryPath} --daemon
ExecReload=${BinaryPath} reload
ExecStop=${BinaryPath} stop

[Install]
WantedBy=multi-user.target`