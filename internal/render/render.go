package render

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the main application model for the BubbleTea app
type Model struct {
}

// Initialize model with a new textarea
func initialModel() Model {
	return Model{} 
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (m Model) View() string {
	return ""
}

