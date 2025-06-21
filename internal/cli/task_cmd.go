package cli

import (
	"github.com/spf13/cobra"
	"github.com/weakphish/yapper/internal/db"
	"github.com/weakphish/yapper/internal/model"
	"github.com/weakphish/yapper/internal/render"
	"golang.org/x/exp/slog"
)

// TaskCmd is the command handler for the "task" command
func TaskCmd(cmd *cobra.Command, args []string) {
	slog.Debug("Task command executed", "args", args)

	if len(args) == 0 {
		slog.Error("No task specified", "args", args)
		cmd.Help()
		return
	}
}

// AddTaskCmd adds a new task with the given description
func AddTaskCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		slog.Error("No task description provided", "args", args)
		cmd.Help()
		return
	}

	title := args[0]
	slog.Info("Adding new task", "title", title)

	db, err := db.InitDB()
	if err != nil {
		slog.Error("error getting database connection", "error", err)
		panic(err)
	}

	var allTasksInDb []model.Task
	result := db.Find(&allTasksInDb)
	if result.Error != nil {
		slog.Error("Could not get tasks from database", "error", result.Error)
	}

	t := render.AddTaskForm(title, allTasksInDb)
	slog.Info("Task created", "task", t)
}
