package parser

import (
	"context"

	"github.com/weakphish/yapper/internal/model"
)

// ParsedNote aggregates the structured entities derived from a Markdown note.
// It exists so downstream layers can consume uniform data regardless of the
// parsing strategy used to produce it.
type ParsedNote struct {
	Note       *model.Note         `json:"note"`
	Tasks      []model.Task        `json:"tasks"`
	LogEntries []model.LogEntry    `json:"log_entries"`
	Mentions   []model.TaskMention `json:"mentions"`
}

// NoteParser defines the contract for each pluggable Markdown parser. Future
// implementations may use simple regexes, Tree-sitter, or other strategies.
type NoteParser interface {
	// Parse converts a raw Note into its structured representation. Returned
	// values must be deterministic for a given input note.
	Parse(ctx context.Context, note *model.Note) (*ParsedNote, error)
}
