package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func installPack(raysBin []byte) {
	installLocation := "/usr/bin"
	if runtime.GOOS == "windows" {
		dir, err := os.UserHomeDir()
		log.Println("Since you're on Windows, you will need to add %%USERPROFILE%% to your path variable if you want to use the 'rays' command globally.")
		if err != nil {
			log.Fatal(err)
		}
		installLocation = dir
	}

	if _, err := os.Stat(path.Join(installLocation, "rays")); err == nil {
		log.Println("Rays is already installed, updating...")

		if (runtime.GOOS == "linux") {
			log.Println("Stopping rays...")
			exec.Command("systemctl", "stop", "rays").Run()
		}
	} else {
		os.Mkdir(path.Join(installLocation, "ray-env"), 0600)
		if _, err := os.Stat(path.Join(installLocation, "ray-env", "rayconfig.json")); errors.Is(err, os.ErrNotExist) {
			log.Println("Created default config.")
			os.WriteFile(path.Join(installLocation, "ray-env", "rayconfig.json"), []byte(defaultConfig), 0600)
		} else {
			log.Println("Config already exists, using existing one.")
		}
	}

	err := os.WriteFile(path.Join(installLocation, "rays" + fileEnding), raysBin, 0667)
	if err != nil {
		log.Fatalf("Failed to write rays binary to %s: %v", path.Join(installLocation, "rays" + fileEnding), err)
	}
	
	registerDaemon(path.Join(installLocation, "rays" + fileEnding))
}

var systemdService string = `[Unit]
Description=ray server (rays)
After=network.target

[Service]
User=${User}
Type=idle
Restart=always
ExecStart=${BinaryPath} daemon
ExecReload=${BinaryPath} reload
ExecStop=${BinaryPath} stop

[Install]
WantedBy=multi-user.target`

var defaultConfig string = `{
    "EnableRayUtil" : true,
    "Projects": [
        {
            "Name": "ray demo",
            "Src": "https://github.com/pyrretsoftwarelabs/ray-demo",
            "Domain": "localhost"
        }
    ]
}`