package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"github.com/urfave/cli/v2"
)
type metadata struct {
	RayVersion string
	Platform string
}

type binary struct {
	Name string
	Blob string
}
type inPack struct {
	Metadata metadata
	Binaries []binary
}

var fileEnding string
func main() {
	fileEnding = ""
	if (runtime.GOOS == "windows") {
		fileEnding = ".exe"
	}

	app := &cli.App{
  		Name: "rayinstall",
  		Usage: "create and use ray installation packages",
		Description: "utility for creating, unpacking and installing ray installation packages",
		Version: "1.0.0",
		Authors: []*cli.Author{
			{
				Name: "axell",
				Email: "mail@axell.me",
			},
		},
		Suggest: true,
		Commands: []*cli.Command{
			{
				Name: "package",
				Aliases: []string{"p", "pack"},
				Usage: "create an installation package",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "leave-files",
						Aliases: []string{"l"},
						Usage: "whether or not to delete the files thats being packaged",
					},
				},
				Action: func(ctx *cli.Context) error {
					fmt.Println("Packaging...")

					rays, err := os.ReadFile("./rays" + fileEnding)
					if (err != nil) {
						fmt.Println(err)
						return err
					}
					_metadata, err := os.ReadFile("./metadata.json")
					var met metadata
					if (err != nil) {
						fmt.Println(err)
						return err
					}
					err = json.Unmarshal(_metadata, &met)
					if (err != nil) {
						fmt.Println(err)
						return err
					}

					pack := inPack{
						Metadata: met,
						Binaries: []binary{
							{
								Name: "rays",
								Blob: base64.StdEncoding.EncodeToString(rays),
							},
						},
					}

					packFile, err := json.Marshal(pack)
					if (err != nil) {
						fmt.Println(err)
						return err
					}
					err = os.WriteFile("ray-" + met.RayVersion + "-" + met.Platform + ".rpack", packFile, 0777)
					if (err != nil) {
						fmt.Println(err)
						return err
					}
					
					if !ctx.Bool("leave-files") {
						for _, file := range []string{"rays" + fileEnding, "metadata.json"} {
							err := os.Remove("./" + file)
							if (err != nil) {
								fmt.Println(err)
								return err
							}
						}
					}

					fmt.Println("Done!")
					return nil
				},
			},
			{
				Name: "unpack",
				Aliases: []string{"u", "unpackage", "up"},
				Usage: "unpack an installation package",
				Args: true,
				Action: func(ctx *cli.Context) error {
					packFile, err := os.ReadFile(ctx.Args().First())
					if (err != nil) {
						fmt.Println(err)
						return err
					}
					var pack inPack
					err = json.Unmarshal(packFile, &pack)
					if (err != nil) {
						fmt.Println(err)
						return err
					}

					for _, file := range pack.Binaries {
						blob, err := base64.StdEncoding.DecodeString(file.Blob)
						if (err != nil) {
							fmt.Println(err)
							return err
						}
						os.WriteFile(file.Name, blob, 0777)
					}
					return nil
				},
			},
			{
				Name: "install",
				Aliases: []string{"i"},
				Usage: "install an installation package",
				Action: func(ctx *cli.Context) error {
					packFile, err := os.ReadFile(ctx.Args().First())
					if (err != nil) {
						fmt.Println(err)
						return err
					}
					var pack inPack
					err = json.Unmarshal(packFile, &pack)
					if (err != nil) {
						fmt.Println(err)
						return err
					}

					installPack(pack)
					return nil
				},
			},
			{
				Name: "uninstall",
				Aliases: []string{"i"},
				Usage: "install rays",
				Action: func(ctx *cli.Context) error {
					uninstall()
					return nil
				},
			},
		},
		
	}
	

	app.Run(os.Args)
}