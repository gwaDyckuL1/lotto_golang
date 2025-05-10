package ui

import (
	"database/sql"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gwaDyckuL1/lotto-scraper/pkg/analysis"
	"github.com/gwaDyckuL1/lotto-scraper/pkg/scraper"
)

const (
	screenIntro screen = iota
	screenMain
	screenProcessing
	screenResults
	screenSelectAnalysis
)

type analysisResultsMsg string
type scrapingResultMsg string

type Model struct {
	choices []string
	cursor  int
	db      *sql.DB
	game    string
	results string
	screen  screen
}

type screen int

var analysisChoices = []string{"Get / Update Game Data", "Count Balls", "Get Probabilities", "Monte Carlo", "Monty Looking For A Win", "Back to Game Select"}
var gameChoices = []string{"Powerball", "Mega Millions", "WA Lotto"}
var analysisFuncs = map[string]func(string, *sql.DB) string{
	"Count Balls":             analysis.CountBalls,
	"Get Probabilities":       analysis.Probabilities,
	"Monte Carlo":             analysis.MonteCarlo,
	"Monty Looking For A Win": analysis.MontysCostToWin,
}
var scrapingFuncs = map[string]func(*sql.DB) string{
	"Powerball":     scraper.ScrapingPowerBall,
	"Mega Millions": scraper.ScrapeMegaMillions,
	"WA Lotto":      scraper.ScrapingWALotto,
}

func runAnalysisCmd(game string, option string, db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		funcToRun := analysisFuncs[option]
		result := funcToRun(game, db)
		return analysisResultsMsg(result)
	}
}

func runScrapingCmd(game string, db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		scrapeToRun := scrapingFuncs[game]
		result := scrapeToRun(db)
		return scrapingResultMsg(result)
	}
}

func initialModel(db *sql.DB) Model {
	return Model{
		screen: screenIntro,
		db:     db,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case analysisResultsMsg:
		m.results = string(msg)
		m.screen = screenResults
	case scrapingResultMsg:
		m.results = string(msg)
		m.screen = screenResults
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(m.choices)) % len(m.choices)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(m.choices)
		case "enter", " ":
			switch m.screen {
			case screenIntro:
				m.screen = screenMain
				m.choices = gameChoices
			case screenMain:
				m.game = m.choices[m.cursor]
				m.cursor = 0
				m.choices = analysisChoices
				m.screen = screenSelectAnalysis
			case screenResults:
				m.screen = screenSelectAnalysis
			case screenSelectAnalysis:
				selected := m.choices[m.cursor]
				switch selected {
				case "Get / Update Game Data":
					m.screen = screenProcessing
					m.cursor = 0
					return m, runScrapingCmd(m.game, m.db)
				case "Back to Game Select":
					m.screen = screenMain
					m.cursor = 0
					m.choices = gameChoices
				default:
					m.screen = screenProcessing
					m.cursor = 0
					return m, runAnalysisCmd(m.game, selected, m.db)
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenIntro:
		return m.viewIntro()
	case screenMain:
		return m.viewMain()
	case screenProcessing:
		return m.viewProcessing()
	case screenResults:
		return m.viewResults()
	case screenSelectAnalysis:
		return m.viewSelectAnalysis()
	}
	return "No view implemented"
}

func TeaTerminal(db *sql.DB) {
	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error booting up UI. Received error %v", err)
		os.Exit(1)
	}
}
