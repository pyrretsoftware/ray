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

var listProp = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true)
var linkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
var greyedOut = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

var listStyle = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1).
	Border(lipgloss.RoundedBorder())