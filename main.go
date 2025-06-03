package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/weakphish/yapper/model"

	tea "github.com/charmbracelet/bubbletea"
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
		// Add a new block
		case "a":
			m.textArea.Focus()
		// Add a new task block
		case "t":
			m.textArea.SetValue("- [ ] ")
			m.textArea.Focus()
		// If in textArea, save input to block
		case "esc":
			if m.textArea.Focused() {
				m.textArea.Blur()
				// Create a new block from the text area content
				content := m.textArea.Value()
				var newBlock model.Block
				
				// Check if the content starts with a task marker
				if len(content) >= 5 && (content[:5] == "- [ ]" || content[:5] == "- [x]") {
					newBlock = model.NewTaskBlock(content)
				} else {
					newBlock = model.NewBlock(content)
				}
				m.blocks = append(m.blocks, newBlock)
				m.textArea.Reset()
			}
		// Add navigation keys
		case "j", "down":
			if uint(len(m.blocks)) > 0 {
				m.cursor = (m.cursor + 1) % uint(len(m.blocks))
			}
		case "k", "up":
			if uint(len(m.blocks)) > 0 {
				if m.cursor == 0 {
					m.cursor = uint(len(m.blocks) - 1)
				} else {
					m.cursor--
				}
			}
		case "e":
			// Edit the current block
			if len(m.blocks) > 0 && int(m.cursor) < len(m.blocks) {
				m.textArea.SetValue(m.blocks[m.cursor].GetContent())
				m.textArea.Focus()
			}
		case "space":
			// Toggle task completion if the current block is a task
			if len(m.blocks) > 0 && int(m.cursor) < len(m.blocks) {
				block := &m.blocks[m.cursor]
				if block.IsTask() {
					content := block.GetContent()
					if strings.HasPrefix(content, "- [ ]") {
						block.SetContent("- [x]" + content[5:])
					} else if strings.HasPrefix(content, "- [x]") {
						block.SetContent("- [ ]" + content[5:])
					}
				}
			}
		}

		return m, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	if m.textArea.Focused() {
		return m.textArea.View()
	}

	var output strings.Builder
	
	// Add a header
	output.WriteString(headerStyle.Render("Yapper Notes") + "\n\n")
	
	if len(m.blocks) == 0 {
		output.WriteString(normalBlockStyle.Render("No blocks yet. Press 'a' to add a note or 't' to add a task."))
		output.WriteString("\n\n")
	} else {
		for i, block := range m.blocks {
			// Get content and prepare styles
			content := block.GetContent()
			var blockStyle lipgloss.Style
			
			// Determine the style based on selection
			if uint(i) == m.cursor {
				blockStyle = selectedBlockStyle
			} else {
				blockStyle = normalBlockStyle
			}
			
			// Format tasks with appropriate styling
			if block.IsTask() {
				if !strings.HasPrefix(content, "- [") {
					content = "- [ ] " + content
				}
				
				// Apply completed task styling if needed
				if strings.HasPrefix(content, "- [x]") {
					content = completedTaskStyle.Render(content)
				} else {
					content = taskStyle.Render(content)
				}
			}
			
			// Render the block with the appropriate style
			output.WriteString(blockStyle.Render(content))
			output.WriteString("\n")
		}
	}
	
	// Add a footer with commands
	output.WriteString("\n")
	output.WriteString(footerStyle.Render("Commands:"))
	output.WriteString("\n")
	output.WriteString(footerStyle.Render("a: Add note   t: Add task   e: Edit   space: Toggle task"))
	output.WriteString("\n")
	output.WriteString(footerStyle.Render("j/k: Navigate   q: Quit"))
	
	return output.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
