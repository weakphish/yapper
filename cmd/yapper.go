package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/weakphish/yapper/internal/cli"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "yapper",
		Short: "Yapper is a terminal-based task management application",
	}

	// Define task command and its subcommands
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
		Run:   cli.TaskCmd,
	}

	addTaskCmd := &cobra.Command{
		Use:   "add [task title]",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		Run:   cli.AddTaskCmd,
	}
	taskCmd.AddCommand(addTaskCmd)
	rootCmd.AddCommand(taskCmd)

	// Define note command and its subcommands
	noteCmd := &cobra.Command{
		Use:   "note",
		Short: "Manage notes",
		Run:   cli.NoteCmd,
	}
	addNoteCmd := &cobra.Command{
		Use:   "add [note title]",
		Short: "Add a new note",
		Args:  cobra.ExactArgs(1),
		Run:   cli.AddNoteCmd,
	}
	noteCmd.AddCommand(addNoteCmd)
	rootCmd.AddCommand(noteCmd)

	if err := fang.Execute(context.TODO(), rootCmd); err != nil {
		slog.Error("failed to execute command", "error", err)
		os.Exit(1)
	}
}
