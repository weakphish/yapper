package render

import (
	"github.com/charmbracelet/huh"
	"github.com/weakphish/yapper/internal/model"
)

func TaskForm(title string) model.Task {
	task := model.Task{
		Title: title,
	} 

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Value(&task.Title).Title("Title"),
			huh.NewInput().Value(&task.Description).Title("Description"),
		),
	)

	form.Run()
	return task
}