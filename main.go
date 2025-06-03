package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/weakphish/yapper/logger"
	"github.com/weakphish/yapper/model"

	tea "github.com/charmbracelet/bubbletea"
)

// Styles for the UI
// Default to everforest
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
			Foreground(lipgloss.Color("#A7C080")).
			Bold(false)

	noteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D3C6AA"))

	completedTaskStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#575F66")).
				Strikethrough(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A7C080")).
			PaddingTop(1)

	statusBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3D484D")).
				Bold(true).
				Padding(0, 1).
				Width(1).
				Align(lipgloss.Center)

	taskBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#343F44")).
			MarginLeft(1).
			Padding(0, 1)

	// Use different styles for cursor selection based on block type
	selectedTaskStyle = selectedBlockStyle.
				Background(lipgloss.Color("#3D484D"))

	selectedNoteStyle = selectedBlockStyle
)

// Model is the main application model for the BubbleTea app
type Model struct {
	blocks       []model.Block
	cursor       uint
	noteTextArea textarea.Model
	taskTextArea textarea.Model
}

// Initialize model with a new textarea
func initialModel() Model {
	slog.Debug("initializing model")
	ti := textarea.New()
	tiTwo := textarea.New()
	return Model{
		[]model.Block{},
		0,
		ti,
		tiTwo,
	}
}

func (m Model) Init() tea.Cmd {
	slog.Debug("initializing bubble tea model")
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Get updates from the note text area if it's focused and return early
	if m.noteTextArea.Focused() {
		slog.Debug("updating note text area")
		m.noteTextArea, cmd = m.noteTextArea.Update(msg)
		// If in textArea, save input to block
		switch msg := msg.(type) {
		// Is it a key press?
		case tea.KeyMsg:
			// Cool, what was the actual key pressed?
			switch msg.String() {
			case "esc":
				slog.Debug("exiting text entry mode")
				m.noteTextArea.Blur()
				// Create a new block from the text area content
				content := m.noteTextArea.Value()
				var newBlock model.Block

				// Check task creation context
				slog.Info("creating note block", "content", content)
				newBlock = model.NewNoteBlock(content)
				slog.Debug("note block created", "is_task", newBlock.IsTask(), "block_type", newBlock.GetTypeString())
				m.blocks = append(m.blocks, newBlock)
				m.noteTextArea.Reset()
			}
		}
		return m, cmd
	}
	if m.taskTextArea.Focused() {
		slog.Debug("updating task text area")
		m.taskTextArea, cmd = m.taskTextArea.Update(msg)
		// If in textArea, save input to block
		switch msg := msg.(type) {
		// Is it a key press?
		case tea.KeyMsg:
			// Cool, what was the actual key pressed?
			switch msg.String() {
			case "esc":
				slog.Debug("exiting text entry mode")
				m.taskTextArea.Blur()
				// Create a new block from the text area content
				content := m.taskTextArea.Value()
				var newBlock model.Block

				// Check task creation context
				slog.Info("creating task block", "content", content)
				newBlock = model.NewTaskBlock(content)
				slog.Debug("task block created", "is_task", newBlock.IsTask(), "block_type", newBlock.GetTypeString())
				m.blocks = append(m.blocks, newBlock)
				m.taskTextArea.Reset()
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			slog.Info("user initiated exit")
			return m, tea.Quit
		// Add a new block
		case "a":
			slog.Info("adding new note block")
			m.noteTextArea.Focus()
		// Add a new task block
		case "t":
			slog.Info("adding new task block (using 't' shortcut)")
			m.taskTextArea.Focus()

		// Add navigation keys
		case "j", "down":
			if uint(len(m.blocks)) > 0 {
				m.cursor = (m.cursor + 1) % uint(len(m.blocks))
				slog.Debug("moved cursor down", "new_position", m.cursor)
			} else {
				slog.Warn("tried to navigate down, but no blocks exist")
			}
		case "k", "up":
			if uint(len(m.blocks)) > 0 {
				if m.cursor == 0 {
					m.cursor = uint(len(m.blocks) - 1)
				} else {
					m.cursor--
				}
				slog.Debug("moved cursor up", "new_position", m.cursor)
			} else {
				slog.Warn("tried to navigate up, but no blocks exist")
			}
		case "e":
			// TODO: Edit the current block
			slog.Info("TODO: edit current block")
		case "space":
			// Toggle task completion if the current block is a task
			if len(m.blocks) > 0 && int(m.cursor) < len(m.blocks) {
				block := &m.blocks[m.cursor]
				if block.IsTask() {
					// Toggle task completion status
					success := block.ToggleTask()
					if success {
						slog.Info("toggled task completion", "block_id", block.GetID(), "completed", block.IsComplete())
					} else {
						slog.Warn("failed to toggle task", "block_id", block.GetID())
					}
				} else {
					slog.Debug("space pressed on non-task block", "block_type", block.GetTypeString())
				}
			} else {
				slog.Warn("tried to toggle task, but no valid block at cursor position", "cursor", m.cursor, "blocks", len(m.blocks))
			}
		}

		return m, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	if m.noteTextArea.Focused() {
		return m.noteTextArea.View()
	}

	var output strings.Builder

	// Add a header
	output.WriteString(headerStyle.Render("Yapper Notes") + "\n\n")

	for i, block := range m.blocks {
		// Get content and prepare styles
		content := block.GetContent()
		var blockStyle lipgloss.Style

		// Determine the style based on selection and block type
		if uint(i) == m.cursor {
			if block.IsTask() {
				blockStyle = selectedTaskStyle
			} else {
				blockStyle = selectedNoteStyle
			}
		} else {
			blockStyle = normalBlockStyle
		}

		// Format tasks with appropriate styling
		if block.IsTask() {
			// Get raw content
			content = block.GetContent()
			slog.Debug("rendering task block", "id", block.GetID(), "content", content, "is_task", block.IsTask(), "block_type", block.GetTypeString(), "completed", block.IsComplete())

			// Create a task block with status badge
			var statusBadge string
			var taskContent string

			// Format based on completion status
			if block.IsComplete() {
				statusBadge = statusBadgeStyle.Foreground(lipgloss.Color("#A7C080")).Render("✓")
				taskContent = completedTaskStyle.Render(content)
				slog.Debug("rendering completed task", "id", block.GetID(), "completed", block.IsComplete())
			} else {
				statusBadge = statusBadgeStyle.Foreground(lipgloss.Color("#D3C6AA")).Render("○")
				taskContent = taskStyle.Render(content)
				slog.Debug("rendering incomplete task", "id", block.GetID(), "completed", block.IsComplete())
			}

			// Combine badge and content with task styling
			content = lipgloss.NewStyle().
				Background(lipgloss.Color("#2D353B")).
				Padding(0, 1).
				Render(statusBadge + " " + taskContent)
		} else {
			slog.Debug("rendering note block", "content", content, "is_task", block.IsTask(), "block_type", block.GetTypeString())
			// Apply note styling
			content = noteStyle.Render("• " + content)
		}

		// Render the block with the appropriate style
		output.WriteString(blockStyle.Render(content))
		output.WriteString("\n")
	}

	// TODO: add help view

	return output.String()
}

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	slog.Info("starting Yapper application")

	// Check if stdout is a terminal
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		slog.Error("stdout is not a terminal", "stdout", os.Stdout.Name())
		fmt.Println("Error: This application requires a terminal.")
		os.Exit(1)
	}

	// Start the Bubble Tea program
	slog.Info("starting Bubble Tea program")
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		slog.Error("application error", "error", err)
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	slog.Info("exiting Yapper application")
}
