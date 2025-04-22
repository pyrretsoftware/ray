package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

var platformCommand = map[string][]string{
	"windows": {"netstat", "-ano"},
	"linux":   {"ss", "-tunlp"},
}

var platformListen = map[string]string{
	"windows": "LISTENING",
	"linux":   "LISTEN",
}

var platformPidSeperator = map[string]string{
	"windows": platformListen["windows"],
	"linux":   "pid=",
}

func parse(content string) map[string][]string {
	table := make(map[string][]string)

	for _, line := range strings.Split(strings.ReplaceAll(content, "\r", ""), "\n") {
		if !strings.Contains(line, platformListen[runtime.GOOS]) {
			continue
		}

		_p1 := strings.Split(line, ":")[1]
		port := strings.ReplaceAll(strings.Split(_p1, " ")[0], " ", "")
		pid := strings.Split(strings.ReplaceAll(strings.Split(line, platformPidSeperator[runtime.GOOS])[1], " ", ""), ",")[0]

		table[pid] = append(table[pid], port)
	}
	return table
}

func getProcessPorts(pid int) []string {
	val, contains := platformCommand[runtime.GOOS]
	if !contains {
		fmt.Println("WARNING: Rays is running on an unsupported platform, and were not able the resolve certain process information.")
		return []string{}
	}

	cmd := exec.Command(val[0], val[1])
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("WARNING: Rays were not able to resolve certain process information.")
		return []string{}
	}

	table := parse(string(out))
	return table[strconv.Itoa(pid)]
}
