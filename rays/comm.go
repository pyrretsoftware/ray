package main

import (
	"encoding/json"
	"io"
)

//com manager

var comLines []ComLine
var extensions map[string]Extension
type Extension struct {
	Description string
	URL string
	ImageBlob string
}

func ReadFromLineLoop(l ComLine) {
	for {
		ReadFromLine(l)
	}
}

func ReadFromLine(l ComLine) {
	recv, resp, setCode := l.Read()
	r, err := io.ReadAll(recv)

	if err != nil {
		rlog.Notify("could not read from comline: " + err.Error(), "err")
		return
	}

	var req comRequest
	jerr := json.Unmarshal(r, &req)
	if jerr != nil && len(r) > 0 {
		setCode(400)
		RespondToWriter(resp, comResponse{
			Data: comData{
				Error: "could not parse request: json error",
			},
		})
		return
	}

	fresp := HandleRequest(req, l)
	if fresp.Data.Error != "" {
		setCode(500)
	}
	RespondToWriter(resp, fresp)
}

//loadlines loads the global comLines according to the provided configuration and initalizes them
func LoadLines(c rayconfig) {
	for _, cl := range comLines {
		rerr.Notify("could not close comline: ", cl.Close(), true)
	}

	newComLines := []ComLine{}
	for _, cl := range c.Com.Lines {		
		if cl.Init() != nil {
			rlog.Notify("could not init comline: ", "err")
			continue
		}
		clptr := &cl
		newComLines = append(newComLines, clptr)
		go ReadFromLineLoop(clptr)
	}
	newComLines = append(newComLines, &HTTPComLine{
		Host: "./comsock.sock",
		Type: "unix",
	})
	comLines = newComLines
}