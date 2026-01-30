package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type PostHogRequest struct {
	APIKey     string `json:"api_key"`
	Event      string `json:"event"`
	DistinctID string `json:"distinct_id"`
	Properties struct {
		Version     string `json:"version"`
		Platform    string `json:"platform"`
		Os          string `json:"$os"`
		SystemdUsed bool `json:"systemdUsed"`
	} `json:"properties"`
}

func Collect(action string) {
	ba, err := os.ReadFile(filepath.Join(os.TempDir(), "rayinstall-tid.bin"))
	if os.IsNotExist(err) {
		ba = make([]byte, 32)
		rand.Read(ba)
		os.WriteFile(filepath.Join(os.TempDir(), "rayinstall-tid.bin"), ba, 0600)
	}
	
	req := PostHogRequest{
		APIKey:     "phc_aIvTbRjWXiCtTAEkrWbpvr3SnWZipUmDmaTy01p3yrk", //not a secret
		Event:      action,
		DistinctID: hex.EncodeToString(ba),
		Properties: struct{Version string "json:\"version\""; Platform string "json:\"platform\""; Os string "json:\"$os\""; SystemdUsed bool "json:\"systemdUsed\""}{
			Version: Version,
			Platform: runtime.GOARCH,
			Os: runtime.GOOS,
			SystemdUsed: systemdDaemonExists(),
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {return}
	http.Post("https://eu.i.posthog.com/capture/", "application/json", bytes.NewReader(reqBody))
}