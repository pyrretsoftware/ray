//go:build !headless && (!compatibility)

package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var SkipInteractions = false

var purpleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
func purple(str string) string {
	return purpleStyle.Render(str)
}

var greyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
func grey(str string) string {
	return greyStyle.Render(str)
}

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

func PromptAction(installText string, alreadyInstalled bool) int {
    fmt.Println()
	fmt.Println("What would you like to do?")
	fmt.Println(grey("←/→ - move • ↵ - select"))
    
    boxes := tea.NewProgram(box{
		items: []string{
			"📦\n" + installText,
			"🔧\nRepair",
			"🧹\nUninstall",
			"💾\nExport",
		},
		itemsAvailable: []bool{
			true,
			alreadyInstalled,
			alreadyInstalled,
			true,
		},
	})
	var boxResultRaw tea.Model
    var err error
    if boxResultRaw, err = boxes.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
    }
	boxResult := boxResultRaw.(box).active
	if boxResult == -1 {
		os.Exit(0)
	}
    return boxResult
}