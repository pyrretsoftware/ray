package main

import (
	"encoding/base64"
	"encoding/json"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type ComError string
const (
	NotPermitted ComError = "PermissonError"
	Success ComError = ""
	TypeError = "TypeError"
)
func ComErrorString(err ComError) string {
	switch err {
	case NotPermitted:
		return "this key does not have the required permissons for this action!"
	case TypeError:
		return "the request payload does not match the expected format."
	}

	return string(err)
}

//com actions
func processesGet(permissons []string) ([]process, ComError) {
	if permOk(permissons, "process:read", "process:all", "special:all", "special:ext") {
		procs := []process{}
		for _, procptr := range processes {
			procs = append(procs, *procptr)
		}
		return procs, Success
	}
	return nil, NotPermitted
}

func permOk(has []string, needs ...string) bool {
	for _, perm := range needs {
		if slices.Contains(has, perm) {return true}
	}
	return false
}

func routerRegister(route string, dest string, permissons []string) ComError {
	if !permOk(permissons, "router:registration", "router:all", "special:all", "special:ext") {
		return NotPermitted
	}
	internalRouteTable[route] = dest
	return Success
}
func routerDeregister(route string, permissons []string) ComError {
	if !permOk(permissons, "router:registration", "router:all", "special:all", "special:ext") {
		return NotPermitted
	}
	delete(internalRouteTable, route)
	return Success
}
func configReadRaw(permissons []string) ([]byte, ComError) {
	if permOk(permissons, "config:read", "config:all", "special:all", "special:ext") {
		return readConfigRaw(), Success
	}
	return nil, NotPermitted
}
func configRead(permissons []string) (rayconfig, ComError) {
	if permOk(permissons, "config:read", "config:all", "special:all", "special:ext") {
		return readConfig(), Success
	}
	return rayconfig{}, NotPermitted
}
func configWrite(permissons []string, ba []byte) (ComError) {
	if permOk(permissons, "config:write", "config:all", "special:all") {
		var _conf rayconfig
		jerr := json.Unmarshal(ba, &_conf)
		if jerr != nil {
			return ComError("Invalid json, refusing to write config: " + jerr.Error())
		}
		err := writeConfRaw(ba)
		if err != nil {
			return ComError("Could not write config: " + err.Error())
		}
		return Success
	}
	return NotPermitted
}
func channelForceRenrollment(permissons []string, projectName string) ComError {
	if permOk(permissons, "channels:reenroll", "channels:all", "special:all", "special:ext") {
		//this is more like a config change
		newProj := []project{}
		for _, proj := range rconf.Projects {
			if proj.Name == projectName {
				proj.ForcedRenrollment = time.Now().Unix()
			}
			newProj = append(newProj, proj)
		}
		rconf.Projects = newProj
		err := writeConf(*rconf)
		if err != nil {
			return ComError(err.Error())
		}

		return Success
	}
	return NotPermitted
}
func channelAuth(permissons []string) (string, ComError) {
	if permOk(permissons, "channels:auth", "channels:all", "special:all", "special:ext") {
		generateAuth()
		return devAuth.Token, Success
	}
	return "", NotPermitted
}
func rayReload(permissons []string) ComError {
	if permOk(permissons, "ray:reload", "ray:all", "special:all", "special:ext") {
		config := readConfig()
		rconf = &config
		
		for _, project := range rconf.Projects {
			startProject(&project, "")
		}
		return Success 
	}
	return NotPermitted
}
func rayUpdate(permissons []string) ComError {
	if permOk(permissons, "ray:reload", "ray:all", "special:all", "special:ext") {
		updateProjects(true)
		return Success
	}
	return NotPermitted
}
func raySystemctlRestart(permissons []string) ComError {
	if permOk(permissons, "ray:restart", "ray:all", "special:all", "special:ext") {
		return ComError(exec.Command("systemctl", "restart", "rays.service").Run().Error())
	}
	return NotPermitted
}
func extensionsRead(permissons []string) (map[string]Extension, ComError) {
	if permOk(permissons, "extensions:read", "extensions:all", "special:all", "special:ext") {
		return extensions, ""
	}
	return map[string]Extension{}, NotPermitted
}

func HandleRequest(r comRequest, l ComLine) comResponse {
	if r.Key == "" {
		return comResponse{
			Data: comData{
				Error: "no key provided!",
			},
		}
	}

	comKeyConf := Key{}
	keyFound := false

	if strings.HasPrefix(r.Key, "ext:") {
		if !l.AllowExtensions() {
			return comResponse{
				Data: comData{
					Error: "this comline does not accept extensions!",
				},
			}
		}

		sections := strings.Split(strings.TrimPrefix(r.Key, "ext:"), ";")
		img := ""
		if len(sections) != 4 && len(sections) != 3 {
			return comResponse{
				Data: comData{
					Error: "inproperly formatted extension declaration: should have 3 or 4 sections!",
				},
			}
		} else if len(sections) == 4 {
			img = sections[3] 
		}

		keyFound = true
		comKeyConf.DisplayName = sections[0]
		comKeyConf.Permissons = []string{"special:ext"}
		extensions[sections[0]] = Extension{
			Description: sections[1],
			URL: sections[2],
			ImageBlob: img,
		}
	} else {
		for _, key := range rconf.Com.Keys {
			if key.Type == "hardcode" && key.Key == r.Key {
				comKeyConf = key
				keyFound = true
			}
		}
	}
	if !keyFound {
		return comResponse{
			Data: comData{
				Error: "invalid key!",
			},
		}
	}
	comKey := comKeyInfo{
		Holder: comKeyConf.DisplayName,
		Permissions: comKeyConf.Permissons,
	}

	response := comData{}
	rpl := r.Payload

	switch r.Action{
	case "process:read":
		pl, err := processesGet(comKey.Permissions)
		response.Error = ComErrorString(err)
		response.Payload = pl
	case "router:register":
		if rpl["route"] == "" || rpl["dest"] == "" {
			response.Error = ComErrorString(TypeError)
		} else {
			response.Error = ComErrorString(routerRegister(rpl["route"], rpl["dest"], comKey.Permissions))
		}
	case "router:deregister":
		if rpl["route"] == "" || rpl["dest"] == "" {
			response.Error = ComErrorString(TypeError)
		} else {
			response.Error = ComErrorString(routerDeregister(rpl["route"], comKey.Permissions))
		}
	case "config:read":
		pl, err := configRead(comKey.Permissions)
		response.Error = ComErrorString(err)
		response.Payload = pl
	case "config:readraw":
		pl, err := configReadRaw(comKey.Permissions)
		response.Error = ComErrorString(err)
		response.Payload = pl
	case "config:write":
		if rpl["config"] == "" {
			response.Error = ComErrorString(TypeError)
		} else {
			conf, err := base64.StdEncoding.DecodeString(rpl["config"])
			if err != nil {
				response.Error = err.Error()
				break
			}
			response.Error = ComErrorString(configWrite(comKey.Permissions, conf))
		}
	case "channel:renroll":
		if rpl["project"] == "" {
			response.Error = ComErrorString(TypeError)
		} else {
			response.Error = ComErrorString(channelForceRenrollment(comKey.Permissions, rpl["project"]))
		}
	case "channel:auth":
		pl, err := channelAuth(comKey.Permissions)
		response.Error, response.Payload = ComErrorString(err), pl
	case "ray:reload":
		response.Error = ComErrorString(rayReload(comKey.Permissions))
	case "ray:systemctl:restart":
		response.Error = ComErrorString(raySystemctlRestart(comKey.Permissions))
	case "ray:update":
		response.Error = ComErrorString(rayUpdate(comKey.Permissions))
	case "ray:shutdown":
		if permOk(comKey.Permissions, "ray:shutdown", "ray:all", "special:all", "special:ext") {
			cleanUpAndExit()
		} else {
			response.Error = ComErrorString(NotPermitted)
		}
	case "extensions:read":
		pl, err := extensionsRead(comKey.Permissions)
		response.Error, response.Payload = ComErrorString(err), pl
	}

	return comResponse{
		Key: &comKey,
		Data: response,
	}
}