//go:build headless

package main

var SkipInteractions = true

//NOOPs
func purple(str string) string {return str}

func PromptAction(installText string, alreadyInstalled bool) int {
	return OptionFlag
}