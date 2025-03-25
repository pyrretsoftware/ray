package main

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

func systemdRegisterDaemon(service string) {
	os.WriteFile("/etc/systemd/system/rays.service", []byte(service), 0644)
	cmd2 := exec.Command("systemctl", "start", "rays")
	cmd := exec.Command("systemctl", "enable", "rays")

	out1, err := cmd.Output()
	out2, err2 := cmd2.Output()
	if (err != nil && err2 != nil){
		rlog.Notify("Cant create systemd daemon:", "err")
		rlog.Notify(err, "err")
		rlog.Notify(string(out1), "err")
		rlog.Notify(err2, "err")
		rlog.Fatal(string(out2))
	} else {
		rlog.Notify("Successfully registered as a systemd daemon!", "done")
	}
}

func registerDaemon() {
	if (runtime.GOOS == "windows") {
		rlog.Notify("Rays has not been daemonized since you're on windows. If you want rays to automatically start on boot you will have to register it as a service manually.", "warn")
		return
	}

	rlog.Println("Now registering daemon.")
	path, err := os.Executable()
	if (err != nil) {
		rlog.Fatal("Cant get current executable: " + err.Error())
	}

	daemon := systemdService

	daemon = strings.ReplaceAll(daemon, "${BinaryPath}", path)
	cuser, err := user.Current()
	if (err != nil) {
		rlog.Fatal("Cant get current user: " + err.Error())
	}

	daemon = strings.ReplaceAll(daemon, "${User}", cuser.Username)
	systemdRegisterDaemon(daemon)	
}

func install() {
	rlog.Println("Welcome to the ray server installed!👋")

	if (runtime.GOOS != "windows" && runtime.GOOS != "linux") {
		rlog.Fatal("Only linux and windows is supported.")
	}
	if (runtime.GOOS == "linux") {
		if _, err := os.Stat("/etc/systemd/system/"); errors.Is(err, os.ErrNotExist) {
			rlog.Fatal("Could not find /etc/systemd/system/ directory. Rays only supports linux distros with systemd (Ubuntu, Arch, Debian, Fedora, etc).")
		}
	}
	
	path, err := os.Getwd()
	if err != nil {
		rlog.Fatal(err)
	}

	rlog.Println("Ray will be installed in the current directory (" + path + "). You might want to cancel here and move the rays binary to another place, like your user/home directory.")
	rlog.Println("Note that once you have continued with the installation, you can't move the ray binary from it's current place without performing a reinstallation.")
	rlog.Println("Do you want to continue installation in this directory? (y,N)")

	r := bufio.NewReader(os.Stdin)
	answer, err := r.ReadByte()
	if err != nil {
		rlog.Fatal(err)
	}

	if (answer != byte('y')) {
		rlog.Println("Alright, exiting...")
		os.Exit(1)
		return
	}

	if _, err := os.Stat("./rayconfig.json"); errors.Is(err, os.ErrNotExist) {
		rlog.Notify("Created default config.", "done")
		os.WriteFile("./rayconfig.json", []byte(defaultConfig), 0600)
	} else {
		rlog.Println("Config already exists, using existing one.")
	}

	registerDaemon()
	rlog.Notify("Installation done!", "done")
	rlog.Notify("Note: if you want to be able to use the rays command globally, you'l need to register this directory in your PATH variable.", "info")
}