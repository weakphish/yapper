package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/weakphish/yapper/model"
)

type FindPageModel struct {
	pages []model.PageTree
}

func FindPage() FindPageModel {
	return FindPageModel{
		pages: []model.PageTree{},
	}
}

func (m FindPageModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m FindPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m FindPageModel) View() string {
	return "TODO find a page"
}
