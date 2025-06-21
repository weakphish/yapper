package cli

import (
	"github.com/spf13/cobra"
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

	t := render.TaskForm(title)
	slog.Info("Task created", "task", t)
}