package cli

import (
	"github.com/spf13/cobra"
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

	description := args[0]
	slog.Info("Adding new task", "description", description)

	// Here you would typically create a new task in your data store
	// For now, we just log the action
	slog.Debug("Task added successfully", "description", description)
}