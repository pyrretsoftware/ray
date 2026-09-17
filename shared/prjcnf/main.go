package prjcnf

import (
	"encoding/json"
)

type baseSharedConfig struct {
	Version string
}

//Package prjcnf handles project configs and translates older ray project configs into modern ones

func TranslateAndMarshalConfig(configBa []byte) (ProjectConfig, error) {
	var base baseSharedConfig
	err := json.Unmarshal(configBa, &base)
	if err != nil {
		return ProjectConfig{}, err
	}

	switch base.Version {
	case "v1":
		var conf ProjectConfig_v1
		if err := json.Unmarshal(configBa, &conf); err != nil {return ProjectConfig{}, err}
		return Translate_v1(conf), nil
	}
	
	//unknown version or latest, parse as latest
	var conf ProjectConfig
	if err := json.Unmarshal(configBa, &conf); err != nil {return ProjectConfig{}, err}
	return conf, nil
}