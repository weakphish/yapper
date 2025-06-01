package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/weakphish/yapper/model"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the main application model for the BubbleTea app
type Model struct {
	blocks   []model.Block
	cursor   uint
	textArea textarea.Model
}

// TODO load from database
func initialModel() Model {
	ti := textarea.New()
	return Model{
		[]model.Block{},
		0,
		ti,
	}
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.textArea.Focused() { // Get updates from the text area
		m.textArea, cmd = m.textArea.Update(msg)
	}

	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.textArea.Focus()
		case "esc":
			if m.textArea.Focused() {
				m.textArea.Blur()
				// TODO: Get content of text area and store in a block
				newBlock := model.NewBlock(m.textArea.Value())
				m.blocks = append(m.blocks, newBlock)
				m.textArea.Reset()
			}
			// TODO: edit block by getting block under cursor
		}

		return m, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	// TODO: view
	if m.textArea.Focused() {
		return m.textArea.View()
	}

	var viewString string
	for _, block := range m.blocks {
		viewString += block.GetContent() + "\n"
	}
	return viewString
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
