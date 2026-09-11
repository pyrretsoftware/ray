package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/urfave/cli/v3"
)

type Authentication struct {
	Type    string
	Hardkey string
}

type Remote struct {
	Name           string
	URL            string
	Authentication Authentication
}

func AddRemote(cc context.Context, cmd *cli.Command) error {
	var rName string
	var rUrl string
	var rAuthType string
	var rAuthValue string

	remotes, err := GetRemotes()
	if err != nil {
		return err
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&rName).Validate(func(s string) error {
				for _, rem := range remotes {
					if rem.Name == s {
						return errors.New("Remote with that name already exists")
					}
				}
				return nil
			}),
			huh.NewInput().Title("URL").Value(&rUrl).Validate(func(s string) error {
				cUrl, err := url.Parse(s)
				if err != nil {
					return errors.New("Invalid url: " + err.Error())
				}

				if cUrl.Scheme != "http" && cUrl.Scheme != "https" {
					return errors.New("Invalid url scheme, must be http or https.")
				}

				return nil
			}),
			huh.NewSelect[string]().Title("Authentication type").Value(&rAuthType).Options(
				huh.NewOption("Hardkey (static key)", "hardkey"),
			),
			huh.NewInput().Title("Key").Value(&rAuthValue),
		),
	).WithAccessible(UseAccesible).Run()
	if err != nil {
		return err
	}

	auth := Authentication{
		Type: rAuthType,
	}

	switch rAuthType {
	case "hardkey":
		auth.Hardkey = rAuthValue
	}

	remotes = append(remotes, Remote{
		Name: rName,
		URL: rUrl,
		Authentication: auth,
	})
	return WriteRemotes(remotes)
}

func WriteRemotes(remotes []Remote) error {
	ePath, err := os.Executable()
	if err != nil {
		return err
	}

	ba, err := json.MarshalIndent(remotes, "", "    ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(ePath)
	err = os.WriteFile(filepath.Join(dir, "remotes.json"), ba, 0666)
	return err
}

func GetRemotes() ([]Remote, error) {
	ePath, err := os.Executable()
	if err != nil {
		return []Remote{}, err
	}

	dir := filepath.Dir(ePath)
	ba, err := os.ReadFile(filepath.Join(dir, "remotes.json"))
	if errors.Is(err, os.ErrNotExist) {
		return []Remote{}, nil
	} else if err != nil {
		return []Remote{}, err
	}

	remotes := []Remote{}
	err = json.Unmarshal(ba, &remotes)
	if err != nil {
		return []Remote{}, err
	}

	return remotes, nil
}