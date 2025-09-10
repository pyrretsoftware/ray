package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

func getClient(rFlag string, unixAddr string) *http.Client {
	if rFlag == "" {
		if _, err := os.Stat(unixAddr); err != nil {
			fmt.Println(redBold.Render("Could not find a comsocket on this machine."), "is ray server installed here or did yoy mean to use a remote comsocket?")
		}
	}
	transport := &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			if rFlag == "" {
				network = "unix"
				addr = unixAddr
			}
			return net.Dial(network, addr)
		},
	}
	
	return &http.Client{
		Transport: transport,
	}
}

type comData struct {
	Payload any `json:"payload,omitempty"`
	Type string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
}
type comRequest struct {
	Action string `json:"action"`
	Payload map[string]string `json:"payload"`
	Key string `json:"key"`
}

type comRayInfo struct {
	RayVer string `json:"version"`
	ProtocolVersion string `json:"protocolVersion"`
}

type comKeyInfo struct {
	Holder string `json:"holder"`
	Permissions []string `json:"permissions"`
}

type comResponse struct {
	Ray comRayInfo `json:"ray"`
	Key *comKeyInfo `json:"key"`
	Data comData `json:"response"`
}

func makeRequest(rFlag string, req comRequest) (error, comResponse) {
	c := getClient(rFlag, "../rays/ray-env/comsock.sock")
	if rFlag == "" {
		rFlag = "http://how-can-you-see-this"
	}

	ba, err := json.Marshal(req)
	if err != nil {
		fmt.Println(redBold.Render("formatting request failed,"), "see the info below:")
		fmt.Println(err)
		return err, comResponse{}
	}
	resp, err := c.Post(rFlag, "application/json", bytes.NewReader(ba))
	if err != nil {
		fmt.Println(redBold.Render("Sending request failed,"), "is the comline online?")
		return err, comResponse{}
	}

	rba, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(redBold.Render("Failed receiving response from comline,"), "do you have a flaky connection?")
		return err, comResponse{}
	}

	var response comResponse
	jerr := json.Unmarshal(rba, &response)
	if jerr != nil {
		fmt.Println(redBold.Render("Invalid response from comline,"), "see the info below:")
		fmt.Println(jerr)
		return jerr, comResponse{}
	}

	if resp.StatusCode != 200 {
		fmt.Println(redBold.Render("Comline reported an error,"), "see the info below:")
		fmt.Println(response.Data.Error)
		return errors.New("comline reported error"), comResponse{}
	}
	return nil, response
}