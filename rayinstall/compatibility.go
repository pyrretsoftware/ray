//go:build compatibility

package main

import (
	"fmt"
	"os"
)

var SkipInteractions = false

//NOOPs
func purple(str string) string {return str}

func PromptAction(installText string, alreadyInstalled bool) int {
	nav := ""
	if !alreadyInstalled {
		nav = " (not available)"
	}
	fmt.Println()
	fmt.Println("Installler options:")
	fmt.Println("1. " + installText)
	fmt.Println("2. Repair" + nav)
	fmt.Println("3. Uninstall" + nav)
	fmt.Println("4. Export")
	fmt.Println()
	fmt.Print("What would you like to do? (1-4): ")
	input := 0
	fmt.Scan(&input)

	if input < 1 || input > 4 {
		fmt.Println("Invalid input")
		os.Exit(0)
	}

	if (input == 2 || input == 3) && !alreadyInstalled {
		fmt.Println("Option not available")
		os.Exit(0)
	}
	return input -1
}