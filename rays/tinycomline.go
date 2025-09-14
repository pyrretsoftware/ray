//tinycomline - single file comline client
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"path/filepath"
)

const (
	LocalComline = ""
)
type RawComData struct {
	Payload any `json:"payload,omitempty"`
	Type string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
}
type RawComRequest struct {
	Action string `json:"action"`
	Payload map[string]string `json:"payload"`
	Key string `json:"key"`
}

type RawComRayInfo struct {
	RayVersion string `json:"version"`
	ProtocolVersion string `json:"protocolVersion"`
}

type RawComKeyInfo struct {
	Holder string `json:"holder"`
	Permissions []string `json:"permissions"`
}

type RawComResponse struct {
	Ray RawComRayInfo `json:"ray"`
	Key *RawComKeyInfo `json:"key"`
	Data RawComData `json:"response"`
}

func getLocalComlineAddress() (string, error) {
	return filepath.Join(dotslash, "comsock.sock"), nil
}

func getComlineClient(address string) (*http.Client, string, error) {
	unix := false
	dummyAddress := address
	if address == LocalComline {
		unix = true
		_addr, err := getLocalComlineAddress()
		if err != nil {return nil, "", err}
		address = _addr
		dummyAddress = "http://comline-dummy-address"
	}
	transport := &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			if unix {
				network = "unix"
				addr = address
			}
			return net.Dial(network, addr)
		},
	}
	
	return &http.Client{
		Transport: transport,
	}, dummyAddress, nil
}

//SendComlineRequest sends a raw request to the comline at the provided address, returning a raw response and any errors the occured. Use the constant LocalComline as the address for local comlines.
func SendComlineRequest(address string, req RawComRequest) (RawComResponse, error) {
	c, addr, err := getComlineClient(address)
	if err != nil {
		return RawComResponse{}, err
	}

	ba, err := json.Marshal(req)
	if err != nil {
		return RawComResponse{}, err
	}
	resp, err := c.Post(addr, "application/json", bytes.NewReader(ba))
	if err != nil {
		return RawComResponse{}, err
	}

	rba, err := io.ReadAll(resp.Body)
	if err != nil {
		return RawComResponse{}, err
	}

	var response RawComResponse
	jerr := json.Unmarshal(rba, &response)
	if jerr != nil {
		return RawComResponse{}, jerr
	}

	if resp.StatusCode != 200 {
		return RawComResponse{}, errors.New("comline reported error: " + response.Data.Error)
	}
	return response, nil
}