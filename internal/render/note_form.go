package render

import (
	"log/slog"

	"github.com/charmbracelet/huh"
	"github.com/weakphish/yapper/internal/model"
	"gorm.io/gorm"
)

// AddNoteForm renders a form for adding a new note
func AddNoteForm(title string, db *gorm.DB) model.Note {
	var (
		content         string
		relatedToTitles []string
	)

	// make an option list of all existing tasks for dependency
	// TODO: refactor out to common function for getting all tasks from DB and making option list
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

	slog.Debug("Created list of taskOptions for depends", "taskOptions", taskOptions)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Value(&title).Title("Title"),
			huh.NewInput().Value(&content).Title("Content"),
			huh.NewMultiSelect[string]().
				Title("Depends On").
				Options(taskOptions...).
				Value(&relatedToTitles),
		),
	)

	err := form.Run()
	if err != nil {
		slog.Error("Error running task form", "error", err)
	}

	// get the ID of the task that it depends on and put it as the dependent
	var relatedTasks []model.Task
	db.Where("title IN ?", relatedToTitles).Find(&relatedTasks)

	note := model.Note{
		Title:        title,
		Content:      content,
		RelatedTasks: relatedTasks,
	}

	return note
}
