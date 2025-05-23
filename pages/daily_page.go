package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/weakphish/yapper/model"
)

type DailyPageModel struct {
	page model.Page
}

func DailyPage() DailyPageModel {
	return DailyPageModel{
		page: model.NewPage(),
	}
}

func (m DailyPageModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m DailyPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m DailyPageModel) View() string {
	return "hello world"
}
