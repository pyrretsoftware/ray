package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func Uninstall(installLocation string, forceFlag bool) {
	fmt.Println("uninstall")
	fmt.Print("Are you sure you want to uninstall? (y/n)")
	continueStr := ""
	_, err := fmt.Scan(&continueStr)
	if err != nil || (continueStr != "y" && continueStr != "yes")  {
		fmt.Println("Aborting...")
		return
	}
	fmt.Println()

	fmt.Println("Stopping rays...")
	StopDaemon(installLocation, forceFlag, false)
	fmt.Println("Removing service...")
	switch runtime.GOOS {
	case "linux":
		err := os.Remove("/etc/systemd/system/rays.service")
		if err != nil {
			fmt.Println("Could not remove the service file.")
			fmt.Println(err)
			return
		}
	case "windows":
		cmd := exec.Command("schtasks", "/delete", "/tn", "rays", "/f")

		ba, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Could not delete the service with task scheduler. MAKE SURE YOU ARE RUNNING AS ADMINISTRATOR.")
			fmt.Println(string(ba), err)
			os.Exit(1)
		}
	}
	binPath := filepath.Join(installLocation, "rays" + fileEnding)
	for {
	fmt.Println("Removing binary...")
		err = os.Remove(binPath)
		if err != nil {
			fmt.Println("Could not remove binary: " + err.Error() + ", trying again in 3 seconds...")
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	fmt.Println("Binary removed")
	fmt.Println("Uninstalled successfully. Configuration was not removed and is stored in", filepath.Join(installLocation, "ray-env"))
}