package main

import "github.com/charmbracelet/lipgloss"

var greenBold = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2fb170"))
var redBold = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d83838ff"))
var blueBold = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#188be9ff"))
var yellowBold = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f5e62f"))
var greyedOut = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
var empty = lipgloss.NewStyle()
var seperatedContent = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("#808080"))
var link = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#71bfffff"))