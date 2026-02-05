package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func systemdDaemonExists() bool {
	_, err := os.Stat("/etc/systemd/system/rays.service")
	return !errors.Is(err, os.ErrNotExist)
}

func StopDaemon(installLocation string, forceFlag bool, repair bool) {
	if runtime.GOOS == "linux" && !systemdDaemonExists() {
		if _, err := os.Stat(filepath.Join(installLocation, "ray-env", "comsock.sock")); err == nil {
			if !forceFlag {
				fmt.Println("Please manually stop ray server before attempting to update. If ray server is actually shut down, use the force flag.")
				os.Exit(0)
			}
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("systemctl", "stop", "rays")

		output, err := cmd.CombinedOutput()
		if err != nil && !repair {
			fmt.Println("Could not stop the service with systemd.")
			fmt.Println(string(output), err)
			os.Exit(1)
		}
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("schtasks", "/end", "/tn", "rays")

		output, err := cmd.CombinedOutput()
		if err != nil && !repair {
			fmt.Println("Could not stop the service with task scheduler. MAKE SURE YOU ARE RUNNING AS ADMINISTRATOR.")
			fmt.Println(string(output), err)
			os.Exit(1)
		}
		time.Sleep(3 * time.Second)
	}
}

var systemdService string = `[Unit]
Description=ray server (rays)
After=network-online.target

[Service]
User=${User}
Type=idle
Restart=always
ExecStart=${BinaryPath} daemon
ExecReload=${BinaryPath} reload
ExecStop=${BinaryPath} exit

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
