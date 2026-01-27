package main

import (
	"encoding/json"
	"io"
	"net/http"
)

//comm.go is for code managing the comlines defined in the config

var comLines []*HTTPComLine
var extensions = map[string]Extension{}

func ComlineHandler(wr http.ResponseWriter, hr *http.Request, l *HTTPComLine) {
	wr.Header().Set("Content-Type", "application/json")
	r, err := io.ReadAll(hr.Body)

	if err != nil {
		rlog.Notify("Could not read from comline: " + err.Error(), "err")
		return
	}

	var req comRequest
	jerr := json.Unmarshal(r, &req)
	if jerr != nil && len(r) > 0 {
		wr.WriteHeader(400)
		RespondToWriter(wr, comResponse{
			Data: comData{
				Error: "could not parse request: json error",
			},
		})
		return
	}

	fresp := HandleRequest(req, l)
	if fresp.Data.Error != "" {
		wr.WriteHeader(500)
	}
	RespondToWriter(wr, fresp)
}

//loadlines loads the comLines defined in the provided configuration and initalizes them
func LoadLines(c rayconfig) {
	for _, cl := range comLines {
		rerr.Notify("could not close comline: ", cl.Close(), true)
	}

	newComLines := []*HTTPComLine{}
	lines := append(c.Com.Lines, HTTPComLine{
		Host: "./comsock.sock",
		Type: "unix",
		ExtensionsEnabled: true,
	})

	for _, cl := range lines {
		clPtr := &cl
		cl.handler = func(w http.ResponseWriter, r *http.Request) {
			ComlineHandler(w, r, clPtr)
		}
		
		initErr := cl.Init()
		if initErr != nil {
			rlog.Notify("Failed initalizing comline: " + initErr.Error(), "err")
			continue
		}

		rlog.Notify("Comline now open on " + cl.Host, "done")
		newComLines = append(newComLines, clPtr)
	}
	comLines = newComLines
}