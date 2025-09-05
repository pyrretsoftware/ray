package main

import (
	"encoding/json"
	"slices"
	"strings"
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
	if permOk(permissons, "config:read", "config:all", "special:all") {
		return readConfigRaw(), Success
	}
	return nil, NotPermitted
}
func configRead(permissons []string) (rayconfig, ComError) {
	if permOk(permissons, "config:read", "config:all", "special:all") {
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
		if len(sections) != 4 {
			return comResponse{
				Data: comData{
					Error: "inproperly formatted extension declaration: should have 4 sections!",
				},
			}
		}

		keyFound = true
		comKeyConf.DisplayName = sections[0]
		comKeyConf.Permissons = []string{"special:ext"}
		extensions[sections[0]] = Extension{
			Description: sections[1],
			URL: sections[2],
			ImageBlob: sections[3],
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
	pl := r.Payload

	switch r.Action{
	case "process:read":
		pl, err := processesGet(comKey.Permissions)
		response.Error = ComErrorString(err)
		response.Payload = pl
	case "router:register":
		if pl["route"] == "" || pl["dest"] == "" {
			response.Error = ComErrorString(TypeError)
		} else {
			response.Error = ComErrorString(routerRegister(pl["route"], pl["dest"], comKey.Permissions))
		}
	case "router:deregister":
		if pl["route"] == "" || pl["dest"] == "" {
			response.Error = ComErrorString(TypeError)
		} else {
			response.Error = ComErrorString(routerDeregister(pl["route"], comKey.Permissions))
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
		if pl["config"] == "" {
			response.Error = ComErrorString(TypeError)
		} else {
			response.Error = ComErrorString(configWrite(comKey.Permissions, []byte(pl["config"])))
		}
	}

	return comResponse{
		Key: &comKey,
		Data: response,
	}
}