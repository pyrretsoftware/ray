package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
)

var hostsFile HostsFile

func readHostsFile() {
	exc, err := os.Executable()
	if err != nil {
		log.Fatal("Cant get current executable: " + err.Error())
	}

	_hosts, err := os.ReadFile(path.Dir(exc) + "/hosts.json")
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(_hosts, &hostsFile); err != nil {
		log.Fatal(err)
	}	
}

func writeHostsFile() {
	exc, err := os.Executable()
	if err != nil {
		log.Fatal("Cant get current executable: " + err.Error())
	}

	b, err := json.MarshalIndent(hostsFile, "", "	")
	if err != nil {
		log.Fatal(err)
	}	

	ferr := os.WriteFile(path.Dir(exc) + "/hosts.json", b, 0777)
	if ferr != nil {
		log.Fatal(ferr)
	}
}