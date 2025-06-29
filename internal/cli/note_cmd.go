package cli

import (
	"github.com/spf13/cobra"
	"github.com/weakphish/yapper/internal/db"
	"github.com/weakphish/yapper/internal/render"
	"golang.org/x/exp/slog"
)

// NoteCmd is the command handler for the "note" command
func NoteCmd(cmd *cobra.Command, args []string) {
	slog.Debug("Note command executed", "args", args)

	if len(args) == 0 {
		slog.Error("No note specified", "args", args)
		cmd.Help()
		return
	}
}

// AddNoteCmd adds a new note with the given description
func AddNoteCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		slog.Error("No note title provided", "args", args)
		cmd.Help()
		return
	}

	title := args[0]
	slog.Info("Adding new note", "title", title)

	db, err := db.InitDB()
	if err != nil {
		slog.Error("error getting database connection", "error", err)
		panic(err)
	}

	n := render.AddNoteForm(title, db)
	slog.Info("Note created", "note", n)
	db.Create(&n)
	if db.Error != nil {
		slog.Error("Error inserting note into database", "error", db.Error)
		return
	}
	slog.Info("Note inserted into database", "note", n)
}
