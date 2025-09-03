package main

import (
	"encoding/json"
	"os"
	"path"
)

func writeConf(_config rayconfig) error {
	config, err := json.MarshalIndent(_config, "", "    ")
	if err != nil {
		rlog.Notify("Cant format config file: " + err.Error(), "err")
		return err
	}
	return writeConfRaw(config)
}

func writeConfRaw(config []byte) error {
	err := os.WriteFile(path.Join(dotslash, "rayconfig.json"), config, 0666)
	if err != nil {
		rlog.Notify("Cant apply config changes: " + err.Error(), "err")
		return err
	}
	return nil
}

func readConfigRaw() []byte {
	_config, err := os.ReadFile(path.Join(dotslash, "rayconfig.json"))
	rerr.Fatal("Failed reading rayconfig: ", err, true)
	return _config
}

func readConfig() rayconfig {
	_config := readConfigRaw()
	var config rayconfig
	rerr.Fatal("Failed parsing rayconfig: ", json.Unmarshal(_config, &config), true)

	var modProjects []project
	for _, project := range config.Projects {
		if len(project.DeployOn) == 0 {
			project.DeployOn = []string{"local"}
		}
		modProjects = append(modProjects, project)
	}
	config.Projects = modProjects

	return config
}