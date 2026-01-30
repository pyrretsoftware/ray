package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func Install(installLocation string, forceFlag bool, repair bool) {
	if _, err := os.Stat(path.Join(installLocation, "rays"+fileEnding)); err == nil {
		fmt.Println("update")
		StopDaemon(installLocation, forceFlag, repair)
	} else {
		fmt.Println("install")
		os.Mkdir(filepath.Join(installLocation, "ray-env"), 0600)
		if _, err := os.Stat(filepath.Join(installLocation, "ray-env", "rayconfig.json")); errors.Is(err, os.ErrNotExist) {
			fmt.Println("Created default config.")
			os.WriteFile(filepath.Join(installLocation, "ray-env", "rayconfig.json"), []byte(defaultConfig), 0600)
		} else {
			fmt.Println("Config already exists, using existing one.")
		}
	}
	
	binPath := filepath.Join(installLocation, "rays" + fileEnding)
	err := os.WriteFile(binPath, Raysbinary, 0667)
	if err != nil {
		log.Fatalf("Failed to write rays binary to %s: %v", path.Join(installLocation, "rays"+fileEnding), err)
	}
	fmt.Println("Binary added/updated.")

	daemon := strings.ReplaceAll(systemdService, "${BinaryPath}", path.Join(installLocation, "rays"+fileEnding))
	cuser, err := user.Current()
	if err != nil {
		log.Fatal("Cant get current user: " + err.Error())
	}

	if runtime.GOOS == "windows" {
		fmt.Println("Binpath =", binPath)
		cmd := exec.Command("schtasks", "/create", "/f", "/tn", "rays", "/tr", binPath + " daemon", "/sc", "onstart")
		ba, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Could not register service with task scheduler. MAKE SURE YOU ARE RUNNING AS ADMINISTRATOR.")
			fmt.Println(string(ba), err)
			os.Exit(1)
		}
		fmt.Println("Created service.")

		cmd = exec.Command("schtasks", "/run", "/tn", "rays")
		ba, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Could not start the service with task scheduler. MAKE SURE YOU ARE RUNNING AS ADMINISTRATOR.")
			fmt.Println(string(ba), err)
			os.Exit(1)
		}
		fmt.Println("Started service.")
	} else {
		err = os.WriteFile("/etc/systemd/system/rays.service", []byte(strings.ReplaceAll(daemon, "${User}", cuser.Username)), 0644)
		if err != nil {
			fmt.Println("Cant create systemd daemon:")
			fmt.Println(err)
		} else {
			fmt.Println("Successfully registered as a systemd daemon!")

			cmd := exec.Command("systemctl", "enable", "rays", "--now")
			err := cmd.Run()
			if err != nil {
				fmt.Println("Could not enable and start rays.")
				fmt.Println()
				os.Exit(1)
			}
		}
	}

	fmt.Println("Installation complete!")
}