package pages

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/weakphish/yapper/model"
)

type DailyPageModel struct {
	pageTree model.PageTree
	cursor   int
}

func NewDailyPage() DailyPageModel {
	d := DailyPageModel{
		pageTree: model.NewPageTree(),
		cursor:   0,
	}
	slog.Debug("Creating a DailyPageModel", "dailyPageModel", d)
	return d
}

// TODO: Constructor for a daily page with input for existing content

func (m DailyPageModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m DailyPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
			// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			m.cursor++
		}
	}
	slog.Debug("Returning message from DailyPageModel", "msg", msg)
	return m, nil
}

// For now, render top level blocks
// Render the top level blocks on the page, and the cursor on the block index
func (m DailyPageModel) View() string {
	view := ""
	traversal := m.pageTree.GetTopLevelBlocks()
	for i, b := range traversal {
		if i == m.cursor {
			view += "> "
		}
		view += b.RenderBlock()
		view += "\n"
	}
	slog.Debug("Created view for DailyPageModel", "traversal", traversal, "view", view)
	return view
}
