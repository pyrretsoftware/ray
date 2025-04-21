package main

import (
	"encoding/json"
	"os"
)

func applyChanges(_config rayconfig) error {
	config, err := json.MarshalIndent(_config, "", "    ")
	if err != nil {
		rlog.Notify("Cant format config file: " + err.Error(), "err")
		return err
	}

	err2 := os.WriteFile(dotslash + "/rayconfig.json", config, 0666)
	if err2 != nil {
		rlog.Notify("Cant apply config changes: " + err2.Error(), "err")
		return err
	}
	return nil
}

func readConfig() rayconfig {
	_config, err := os.ReadFile(dotslash + "/rayconfig.json")
	if err != nil {
		rlog.Fatal(err)
	}

	var config rayconfig
	if err := json.Unmarshal(_config, &config); err != nil {
		rlog.Fatal(err)
	}

	return config
}

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