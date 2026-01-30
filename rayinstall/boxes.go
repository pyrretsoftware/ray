package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbletea"
)

var BoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(11).Height(4).Align(lipgloss.Center, lipgloss.Center)
var BoxStyleGrey = BoxStyle.Foreground(lipgloss.Color("#808080")).BorderForeground(lipgloss.Color("#808080"))
var BoxStyleSelected = BoxStyle.Foreground(lipgloss.Color("5")).BorderForeground(lipgloss.Color("5"))

type box struct {
    active int
    items []string
    itemsAvailable []bool
}

func (m box) Init() tea.Cmd {
    return nil
}

func (m box) View() string {
    result := []string{}
    availableIndex := 0
    for i, item := range m.items {
        style := BoxStyle
        if m.active == availableIndex && m.itemsAvailable[i] {
            style = BoxStyleSelected
            availableIndex++
        } else if !m.itemsAvailable[i] {
            style = BoxStyleGrey
        } else {
            availableIndex++
        }
        result = append(result, style.Render(item))
    }
    return lipgloss.JoinHorizontal(lipgloss.Top, result...) + "\n"
}

func (m box) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        switch msg.String() {

        case "ctrl+c":
            m.active = -1
            return m, tea.Quit

        case "left":
            if m.active > 0 {
                m.active--
            }

        case "right":
            if m.active < len(m.items)-1 {
                m.active++
            }

        case "enter", " ":
            return m, tea.Quit
        }
    case tea.MouseMsg:
        
    }

    return m, nil
}