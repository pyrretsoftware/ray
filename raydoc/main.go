package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

//go:embed options/introduction.md
var DocsIntroduction string
//go:embed options/dials_indent_prefix.txt
var IndentHeadings string

var Config raydocConfig

func main() {
	fmt.Println("raydoc generates documentation from type files.")
	fmt.Println("version 2.1")
	fmt.Println()
	fmt.Print("arming workspace...")
	err := os.MkdirAll("./out", 0600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(" all ok")

	fmt.Println("reading config...")
	confBa, err := os.ReadFile("./raydoc.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(confBa, &Config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("identified", len(Config.Jobs), "jobs from config.")
	for jobI, job := range Config.Jobs {
		st := time.Now()
		fmt.Println("starting job " + job.Name + " (" + strconv.Itoa(jobI + 1) + "/" + strconv.Itoa(len(Config.Jobs)) + ")")
		ProcessJob(job.Output, job.Path, job.Base)
		fmt.Println("finished job " + job.Name + " (" + strconv.Itoa(jobI + 1) + "/" + strconv.Itoa(len(Config.Jobs)) + ") in " + strconv.FormatInt(time.Since(st).Milliseconds(), 10) + "ms")
	}
	fmt.Println("everything finished!")
}