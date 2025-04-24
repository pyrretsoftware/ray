package main

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func confirm(text string) bool {
	var result bool
	huh.NewConfirm().
		Title(text).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	return result
}

func prompt(text string) string {
	var result string
	huh.NewInput().
		Prompt(text).
		Value(&result).
		Run()
	return result
}

var serror lipgloss.Style = lipgloss.NewStyle().
Bold(true).
Foreground(lipgloss.Color("#ff0033"))

var warning lipgloss.Style = lipgloss.NewStyle().
Bold(true).
Foreground(lipgloss.Color("#FF5733"))