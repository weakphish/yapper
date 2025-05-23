package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/kevm/bubbleo/navstack"
	shell "github.com/kevm/bubbleo/shell"
	"github.com/weakphish/yapper/pages"
)

type PageType int

const (
	DailyPage PageType = iota
	FindPage
	TagPage
)

// applicationModel is the main application model for the BubbleTea app
type applicationModel struct {
	currentPage tea.Model // TODO: pointer to a model
	pageStack   []tea.Model
}

// TODO: load from database
func initialModel() applicationModel {
	// TODO: get today's page
	today := pages.DailyPage()
	return applicationModel{
		currentPage: &today,
		pageStack:   []tea.Model{},
	}
}

func (m applicationModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	// TODO: Load application from database, load today's daily page
	return nil
}

func (m applicationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Subordinate to the page type
	subMsg, subCmd := m.currentPage.Update(m)

	return subMsg, subCmd
}

func (m applicationModel) View() string {
	return m.currentPage.View()
}

func main() {
	// Initialize logging
	f, _ := os.OpenFile("log.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	log.SetOutput(f)
	log.SetFormatter(log.JSONFormatter) // Use JSON format
	log.SetLevel(log.DebugLevel)
	log.Info("Initializing Yapper...")

	// Initialize application model
	m := initialModel()
	s := shell.New()

	// Start the page stack on the home page, which uses the base applicationModel defined here
	s.Navstack.Push(navstack.NavigationItem{Model: m, Title: "Yapper"})

	// Run!
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
