package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
)

var (
	centerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(80).
			Height(20).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 4)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffff00")).
		//Background(lipgloss.Color("63")).
		Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

func (m Model) viewIntro() string {
	s := `
	Welcome to the Lotto WINNER!!
	
	The program that will give you some lotto numbers...that might work.
	
	If you're lucky!
	

	Press Enter to get started or Q to quit`
	return lipgloss.Place(80, 20, lipgloss.Center, lipgloss.Center, centerStyle.Render(s))
}

func (m Model) viewMain() string {
	s := "Select the game you want to win!\n\n"

	for i, choice := range m.choices {
		cursor := " "
		style := normalStyle
		if m.cursor == i {
			cursor = ">"
			style = highlightStyle
		}
		s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
	}

	return lipgloss.Place(80, 20, lipgloss.Center, lipgloss.Center, centerStyle.Render(s))
}

func (m Model) viewProcessing() string {
	s := "Processing request...Please wait."
	return lipgloss.Place(80, 20, lipgloss.Center, lipgloss.Center, centerStyle.Render(s))
}

func (m Model) viewResults() string {
	return lipgloss.Place(80, 20, lipgloss.Center, lipgloss.Center, centerStyle.Render(m.results+"\n\nPress Enter To Go Back"))
}

func (m Model) viewSelectAnalysis() string {
	s := m.game + "\n"
	s += "What do you want to do?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		style := normalStyle
		if m.cursor == i {
			cursor = ">"
			style = highlightStyle
		}
		s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
	}

	return lipgloss.Place(80, 20, lipgloss.Center, lipgloss.Center, centerStyle.Render(s))
}
