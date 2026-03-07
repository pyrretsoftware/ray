//go:build headless

package main

import (
	"flag"
)

var SkipInteractions = true

//NOOPs
func purple(str string) string {return str}

func PromptAction(installText string, alreadyInstalled bool) int {
	iptr := flag.Int("option", 0, "option int 0-3")
	flag.Parse()
	return *iptr
}