package render

import (
	"github.com/charmbracelet/huh"
	"github.com/weakphish/yapper/internal/model"
)

func AddTaskForm(title string, allTasks []model.Task) model.Task {
	var (
		description string
		status      int
		dependsOnId int
	)

	taskOptions := make([]huh.Option[int], len(allTasks))
	for i, option := range taskOptions {
		option.Key = allTasks[i].Title
		option.Value = allTasks[i].ID
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Value(&title).Title("Title"),
			huh.NewInput().Value(&description).Title("Description"),
			huh.NewSelect[int]().
				Title("Status").
				Options(
					huh.NewOption("Todo", 0),
					huh.NewOption("Doing", 1),
					huh.NewOption("Done", 2),
				).
				Value(&status),
			huh.NewSelect[int]().
				Title("Depends On").
				Options(taskOptions...).
				Value(&dependsOnId),
		),
	)

	task := model.Task{
		Title:       title,
		Description: description,
	}

	form.Run()
	return task
}
