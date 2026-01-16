package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func config(cc context.Context, cmd *cli.Command) error {
	err, resp := makeRequest(cmd.String("remote"), comRequest{
		Action: "config:readraw",
		Key: cmd.String("hardkey"),
	}, cmd.Bool("debug-local-rays"))
	if err != nil {
		return err
	}

	configstr, ok := resp.Data.Payload.(string)
	config, berr := base64.StdEncoding.DecodeString(configstr)
	if !ok || berr != nil {
		return badFormat()
	}

	f, err := os.CreateTemp("", "rayc-config-*.json")
	if err != nil {
		fmt.Println(redBold.Render("Could not create temporary file, are you running rayc with suffient permissons?"))
		return errors.New("file open failed")
	}

	_, werr := f.Write(config)
	if werr != nil {
		fmt.Println(redBold.Render("Could not write to temporary file, are you running rayc with suffient permissons?"))
		return errors.New("file write failed")
	}

	cerr := f.Close()
	if cerr != nil {
		fmt.Println(redBold.Render("Could not close temporary file."))
		return errors.New("file write failed")
	}

	fmt.Println("You may now edit the remote server's configuration using the file located at", blueBold.Render(f.Name()))
	fmt.Println(greyedOut.Render("s - save changes • d - discard changes"))
	for {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {log.Fatal("stdin read error")}
		if b[0] == 'd' {
			os.Exit(0)
		} else if b[0] == 's' {
			break
		}
	}
	fmt.Println()
	fmt.Println("Saving changes.")

	newconfig, err := os.ReadFile(f.Name())
	if err != nil {
		fmt.Println(redBold.Render("Could not read temporary file, try closing your editor after saving."))
		return errors.New("read temporary file error")
	}

	err, _ = makeRequest(cmd.String("remote"), comRequest{
		Action: "config:write",
		Key: cmd.String("hardkey"),
		Payload: map[string]string{
			"config" : base64.StdEncoding.EncodeToString(newconfig),
		},
	}, cmd.Bool("debug-local-rays"))
	if err != nil {return errors.New("req error")}

	fmt.Println("Remember, you'll also need to run " + greenBold.Render("rayc reload") + " for the changes to take effect!")
	return os.Remove(f.Name())
}