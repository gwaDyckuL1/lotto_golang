package main

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gwaDyckuL1/lotto-scraper/pkg/analysis"
	"github.com/gwaDyckuL1/lotto-scraper/pkg/scraper"
)

const (
	screenMain screen = iota
	screenScrape
	screenScraping
	screenAnalyze
	screenAnalysisRunning
	screenAnalysisSelect
)

type logMsg string
type logWriter struct {
	ch chan logMsg
}
type scrapeDoneMsg struct{}
type screen int
type model struct {
	choices          []string
	cursor           int
	db               *sql.DB
	logs             []string
	screen           screen
	selectedGame     string
	selectedAnalysis string
	logCh            chan logMsg
}

var analysisFuncs = map[string]func(db *sql.DB, w io.Writer, option string){
	"Mega Millions": analysis.AnalyzeMegaMillions,
	"Powerball":     analysis.AnalyzePowerball,
}
var analysisOptions = []string{"Counts", "Probabilities", "Top5", "Monte Carlo"}
var gameOptions = []string{"Mega Millions", "Powerball"}
var scrapeFuncs = map[string]func(db *sql.DB, w io.Writer){
	"Mega Millions": scraper.ScrapeMegaMillions,
	"Powerball":     scraper.ScrapingPowerBall,
}

func mainTeaTerminal(db *sql.DB) {
	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Println("UI broke with error", err)
	}
}

func initialModel(db *sql.DB) model {
	return model{
		choices: []string{"Scrape Data", "Analyze Data"},
		screen:  screenMain,
		cursor:  0,
		db:      db,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (w logWriter) Write(p []byte) (int, error) {
	w.ch <- logMsg(p)
	return len(p), nil
}

func gettingStatusUpdates(m model, fn func(*sql.DB, io.Writer)) (model, tea.Cmd) {
	logCh := make(chan logMsg)
	m.logCh = logCh

	go func() {
		fn(m.db, logWriter{ch: logCh})
		close(logCh)
	}()

	return m, readLogChannel(logCh)
}

func readLogChannel(ch <-chan logMsg) tea.Cmd {
	return func() tea.Msg {
		if msg, ok := <-ch; ok {
			return msg
		}
		return scrapeDoneMsg{}
	}

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case logMsg:
		m.logs = append(m.logs, string(msg))
		return m, readLogChannel(m.logCh)
	case scrapeDoneMsg:
		m.logs = append(m.logs, "This step done\n")
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return initialModel(m.db), nil
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(m.choices)) % len(m.choices)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(m.choices)
		case "enter", " ":
			switch m.screen {
			case screenAnalyze:
				m.selectedGame = m.choices[m.cursor]
				m.screen = screenAnalysisSelect
				m.choices = analysisOptions
				m.cursor = 0
			case screenAnalysisRunning:
				m.screen = screenAnalysisSelect
				m.choices = analysisOptions
				m.cursor = 0
			case screenAnalysisSelect:
				m.selectedAnalysis = m.choices[m.cursor]
				m.screen = screenAnalysisRunning
				m.cursor = 0
				m.logs = nil
				m.choices = []string{"Return to Analysis Options"}
				if fn, ok := analysisFuncs[m.selectedGame]; ok {
					return gettingStatusUpdates(m, func(db *sql.DB, w io.Writer) {
						fn(db, w, m.selectedAnalysis)
					})
				}
				m.logs = append(m.logs, "This game is not known")
			case screenMain:
				switch m.choices[m.cursor] {
				case "Scrape Data":
					m.screen = screenScrape
					m.choices = gameOptions
					m.cursor = 0
				case "Analyze Data":
					m.screen = screenAnalyze
					m.choices = gameOptions
					m.cursor = 0
				}
			case screenScrape:
				m.selectedGame = m.choices[m.cursor]
				m.screen = screenScraping
				m.choices = []string{"Return to Scraping Options"}
				m.cursor = 0
				m.logs = nil
				if fn, ok := scrapeFuncs[m.selectedGame]; ok {
					return gettingStatusUpdates(m, fn)
				}
				m.logs = append(m.logs, "This game is not known")

			case screenScraping:
				m.screen = screenScrape
				m.cursor = 0
				m.choices = gameOptions
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	switch m.screen {
	case screenAnalyze:
		b.WriteString("What game data do you want to analyze?\n")
	case screenAnalysisSelect:
		b.WriteString(fmt.Sprintf("What information do you want for %s\n", m.selectedGame))
	case screenAnalysisRunning:
		b.WriteString(fmt.Sprintf("Getting %s for %s\n", m.selectedAnalysis, m.selectedGame))
		if m.selectedAnalysis == "Monte Carlo" {
			b.WriteString("Running simulation...Please wait\n")
		}
		for _, line := range m.logs {
			b.WriteString(line + "\n")
		}
	case screenMain:
		b.WriteString("***Main Menu***\nWhat do you want to do?\n")
	case screenScrape:
		b.WriteString("Select the game you want to scrape\n")
	case screenScraping:
		b.WriteString(fmt.Sprintf("Scraping %s\n", m.selectedGame))
		for _, line := range m.logs {
			b.WriteString(line + "\n")
		}
	}

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		b.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}
	b.WriteString("\nPress q to quit or ESC for main menu.\n")
	return b.String()
}
