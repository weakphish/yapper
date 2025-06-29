package render

import (
	"log/slog"

	"github.com/charmbracelet/huh"
	"github.com/weakphish/yapper/internal/model"
	"gorm.io/gorm"
)

// AddTaskForm creates a form for adding a new task. It takes in a list of all tasks in the
// database to populate the "Depends On" dropdown.
func AddTaskForm(title string, db *gorm.DB) model.Task {
	var (
		description    string
		status         model.TaskStatus
		dependsOnTitle string
	)

	// make an option list of all existing tasks for dependency
	var allTasks []model.Task
	result := db.Find(&allTasks)
	if result.Error != nil {
		slog.Error("Could not get tasks from database", "error", result.Error)
	}
	slog.Debug("Got all tasks from database", "tasks", allTasks)

	taskOptions := make([]huh.Option[string], len(allTasks))
	for i, task := range allTasks {
		taskOptions[i].Key = task.Title
		taskOptions[i].Value = task.Title
	}

	slog.Info("Created list of taskOptions for depends", "taskOptions", taskOptions)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Value(&title).Title("Title"),
			huh.NewInput().Value(&description).Title("Description"),
			huh.NewSelect[model.TaskStatus]().
				Title("Status").
				Options(
					huh.NewOption("Todo", model.Todo),
					huh.NewOption("In Progress", model.InProgress),
					huh.NewOption("Completed", model.Completed),
				).
				Value(&status),
			huh.NewSelect[string]().
				Title("Depends On").
				Options(taskOptions...).
				Value(&dependsOnTitle),
		),
	)

	err := form.Run()
	if err != nil {
		slog.Error("Error running task form", "error", err)
	}

	// get the ID of the task that it depends on and put it as the dependent
	var dependsOnTask model.Task
	db.Where(&model.Task{Title: dependsOnTitle}).Find(&dependsOnTask)

	task := model.Task{
		Title:       title,
		Description: description,
		Status:      status,
		DependsOn:   &dependsOnTask,
		DependsOnID: dependsOnTask.ID,
	}

	return task
}
