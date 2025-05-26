package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/weakphish/yapper/pages"
)

// applicationModel is the main application model for the BubbleTea app
type applicationModel struct {
	currentPage tea.Model
	pageStack   []tea.Model
}

func initialModel() applicationModel {
	// TODO: get today's page
	today := pages.NewDailyPage()

	return applicationModel{
		currentPage: &today,
		pageStack:   []tea.Model{},
	}
}

func (m applicationModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	// TODO: Load application from database, load today's daily page as currentPage
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

	// Subordinate to the current page type
	subMsg, subCmd := m.currentPage.Update(m)
	slog.Debug("Got message from the model's current page", "subMsg", subMsg, "subCmd", subCmd)

	return subMsg, subCmd
}

func (m applicationModel) View() string {
	view := m.currentPage.View()
	slog.Debug("Deferring to current page's view", "view", view)
	return view
}

func main() {
	// Initialize logging
	logFile, _ := os.OpenFile("yap.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	logHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	slog.Info("Logging initialized")

	// Initialize application model
	m := initialModel()

	// Run!
	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
