package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/kevm/bubbleo/menu"
	"github.com/kevm/bubbleo/navstack"
	shell "github.com/kevm/bubbleo/shell"
	"github.com/weakphish/yapper/pages"
)

// applicationModel is the main application model for the BubbleTea app
type applicationModel struct {
	menu menu.Model
}

// TODO: load from database
func initialModel() applicationModel {
	daily := menu.Choice{
		Title:       "Daily Page",
		Description: "Today's daily note page",
		Model:       pages.DailyPage(),
	}
	find := menu.Choice{
		Title:       "Find Page",
		Description: "Find a page by name",
		Model:       pages.FindPage(),
	}

	choices := []menu.Choice{daily, find}

	return applicationModel{
		menu: menu.New("Pages", choices, nil),
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
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	log.Debug("Menu prior to update: ", "menu", m.menu)
	updatedMenu, cmd := m.menu.Update(msg)
	m.menu = updatedMenu.(menu.Model)
	log.Debug("Menu after update: ", "menu", m.menu)
	return m, cmd

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m applicationModel) View() string {
	// TODO: view
	return m.menu.View()
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
