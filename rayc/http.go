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
	"path/filepath"
	"runtime"

	"charm.land/huh/v2"
	"github.com/urfave/cli/v3"
)

func getClient(rFlag string, unixAddr string) *http.Client {
	if rFlag == "" {
		if _, err := os.Stat(unixAddr); err != nil {
			fmt.Println(redBold.Render("Could not find a comsocket on this machine."), "Is ray server installed here or did you intend to use a remote comsocket?")
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
	Payload any    `json:"payload,omitempty"`
	Type    string `json:"type,omitempty"`
	Error   string `json:"error,omitempty"`
}
type comRequest struct {
	Action  string            `json:"action"`
	Payload map[string]string `json:"payload"`
	Key     string            `json:"key"`
}

type comRayInfo struct {
	RayVer          string `json:"version"`
	ProtocolVersion string `json:"protocolVersion"`
}

type comKeyInfo struct {
	Holder      string   `json:"holder"`
	Permissions []string `json:"permissions"`
}

type comResponse struct {
	Ray  comRayInfo  `json:"ray"`
	Key  *comKeyInfo `json:"key"`
	Data comData     `json:"response"`
}

func getLocalComlineAddress() (string, error) {
	address := ""
	switch runtime.GOOS {
	case "windows":
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		address = filepath.Join(dir, "rays", "ray-env", "comsock.sock")
	case "linux":
		address = "/usr/bin/ray-env/comsock.sock"
	default:
		return "", errors.New("unsupported platform, linux and windows are supported")
	}
	return address, nil
}

func GetKey(auth Authentication) (string, error) {
	if auth.Type == "hardkey" {
		return auth.Hardkey, nil
	}
	return "", errors.New("unsupported authentication type")
}

func makeRequest(cmd *cli.Command, req comRequest) (error, comResponse) {
	localPath, err := getLocalComlineAddress()
	if err != nil {
		return err, comResponse{}
	}

	if cmd.Bool("debug-local-rays") {
		localPath = "../rays/ray-env/comsock.sock"
	}

	target := cmd.String("remote")
	if target == "" {
		remotes, err := GetRemotes()
		if err != nil {
			return err, comResponse{}
		}

		remotes = append([]Remote{
			{
				Name: "Local server",
				URL: "http://how-can-you-see-this",
				Authentication: Authentication{
					Type: "hardkey",
					Hardkey: "ext:Rayc;This extension is used by rayc for local communications;https://ray.pyrret.com",
				},
			},
		}, remotes...)

		options := []huh.Option[Remote]{}

		for _, remote := range remotes {
			options = append(options, huh.Option[Remote]{
				Key:   remote.Name,
				Value: remote,
			})
		}

		var selected Remote
		if len(remotes) != 0 {
			err := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[Remote]().
						Title("Choose a comline:").
						Options(options...).
						Value(&selected),
				),
			).WithAccessible(UseAccesible).Run()

			if err != nil {
				return err, comResponse{}
			}
		}
		target = selected.URL
		key, err := GetKey(selected.Authentication)
		if err != nil {
			return err, comResponse{}
		}
		req.Key = key


	}
	c := getClient(target, localPath)

	ba, err := json.Marshal(req)
	if err != nil {
		return err, comResponse{}
	}
	resp, err := c.Post(target, "application/json", bytes.NewReader(ba))
	if err != nil {
		return err, comResponse{}
	}

	rba, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, comResponse{}
	}

	var response comResponse
	jerr := json.Unmarshal(rba, &response)
	if jerr != nil {
		return jerr, comResponse{}
	}

	if resp.StatusCode != 200 {
		return errors.New(response.Data.Error), comResponse{}
	}
	return nil, response
}
