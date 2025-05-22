package main

import (
	"fmt"
	"os"

	"github.com/weakphish/yapper/model"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the main application model for the BubbleTea app
type Model struct {
	pages        []model.Page
	currentPage  *model.Page
	currentBlock *model.Block
}

// TODO load from database
func initialModel() Model {
	return Model{
		pages:        []model.Page{},
		currentPage:  nil,
		currentBlock: nil,
	}
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	// TODO: view
	return ""
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
