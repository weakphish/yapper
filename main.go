package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/weakphish/yapper/logger"
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

// Initialize model with a new textarea
func initialModel() Model {
	logger.Debug("initializing model")
	ti := textarea.New()
	return Model{
		[]model.Block{},
		0,
		ti,
	}
}

func (m Model) Init() tea.Cmd {
	logger.Debug("initializing bubble tea model")
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.textArea.Focused() { // Get updates from the text area
		logger.Debug("updating text area")
		var err error
		m.textArea, cmd = m.textArea.Update(msg)
		if err != nil {
			logger.Error("failed to update text area", "error", err)
		}
	}

	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			logger.Info("user initiated exit")
			return m, tea.Quit
		// Add a new block
		case "a":
			logger.Info("adding new note block")
			m.textArea.Focus()
		// Add a new task block
		case "t":
			logger.Info("adding new task block")
			m.textArea.SetValue("- [ ] ")
			m.textArea.Focus()
		// If in textArea, save input to block
		case "esc":
			if m.textArea.Focused() {
				logger.Debug("exiting text entry mode")
				m.textArea.Blur()
				// Create a new block from the text area content
				content := m.textArea.Value()
				var newBlock model.Block
				
				// Check if the content starts with a task marker
				if len(content) >= 5 && (content[:5] == "- [ ]" || content[:5] == "- [x]") {
					logger.Info("creating new task block", "content", content)
					newBlock = model.NewTaskBlock(content)
				} else {
					logger.Info("creating new note block", "content", content)
					newBlock = model.NewBlock(content)
				}
				m.blocks = append(m.blocks, newBlock)
				m.textArea.Reset()
			}
		// Add navigation keys
		case "j", "down":
			if uint(len(m.blocks)) > 0 {
				m.cursor = (m.cursor + 1) % uint(len(m.blocks))
				logger.Debug("moved cursor down", "new_position", m.cursor)
			} else {
				logger.Warn("tried to navigate down, but no blocks exist")
			}
		case "k", "up":
			if uint(len(m.blocks)) > 0 {
				if m.cursor == 0 {
					m.cursor = uint(len(m.blocks) - 1)
				} else {
					m.cursor--
				}
				logger.Debug("moved cursor up", "new_position", m.cursor)
			} else {
				logger.Warn("tried to navigate up, but no blocks exist")
			}
		case "e":
			// Edit the current block
			if len(m.blocks) > 0 && int(m.cursor) < len(m.blocks) {
				logger.Info("editing block", "block_id", m.cursor)
				m.textArea.SetValue(m.blocks[m.cursor].GetContent())
				m.textArea.Focus()
			} else {
				logger.Warn("tried to edit, but no valid block at cursor position", "cursor", m.cursor, "blocks", len(m.blocks))
			}
		case "space":
			// Toggle task completion if the current block is a task
			if len(m.blocks) > 0 && int(m.cursor) < len(m.blocks) {
				block := &m.blocks[m.cursor]
				if block.IsTask() {
					content := block.GetContent()
					if strings.HasPrefix(content, "- [ ]") {
						logger.Info("marking task as complete", "task", content)
						block.SetContent("- [x]" + content[5:])
					} else if strings.HasPrefix(content, "- [x]") {
						logger.Info("marking task as incomplete", "task", content)
						block.SetContent("- [ ]" + content[5:])
					} else {
						logger.Warn("task has invalid format", "content", content)
					}
				} else {
					logger.Debug("space pressed on non-task block", "block_type", block.GetType())
				}
			} else {
				logger.Warn("tried to toggle task, but no valid block at cursor position", "cursor", m.cursor, "blocks", len(m.blocks))
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
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("starting Yapper application")
	
	// Check if stdout is a terminal
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		logger.Error("stdout is not a terminal", "stdout", os.Stdout.Name())
		fmt.Println("Error: This application requires a terminal.")
		os.Exit(1)
	}

	// Start the Bubble Tea program
	logger.Info("starting Bubble Tea program")
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		logger.Error("application error", "error", err)
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	
	logger.Info("exiting Yapper application")
}
