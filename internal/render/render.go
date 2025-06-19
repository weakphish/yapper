package render

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/weakphish/yapper/internal/logger"
)

// Styles for the UI
var (
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#D3C6AA")).
		Background(lipgloss.Color("#2D353B")).
		PaddingLeft(2).
		Width(60).
		Align(lipgloss.Center)

	normalBlockStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D3C6AA")).
		PaddingLeft(2)

	selectedBlockStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4D5A63")).
		PaddingLeft(2)

	taskStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A7C080"))

	completedTaskStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#575F66")).
		Strikethrough(true)

	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A7C080")).
		PaddingTop(1)
)

// Model is the main application model for the BubbleTea app
type Model struct {
}

// Initialize model with a new textarea
func initialModel() Model {
	return Model{} 
}

func (m Model) Init() tea.Cmd {
	logger.Debug("initializing bubble tea model")
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (m Model) View() string {
	return ""
}

func render() {
}
