package parser

import (
	"context"

	"github.com/weakphish/yapper/internal/model"
)

// ParsedNote captures the structured metadata emitted by a NoteParser. It
// includes the original note contents and the derived entities (tasks, log
// entries, mentions) that feed downstream indexes.
type ParsedNote struct {
	Note       *model.Note         `json:"note"`
	Tasks      []model.Task        `json:"tasks"`
	LogEntries []model.LogEntry    `json:"log_entries"`
	Mentions   []model.TaskMention `json:"mentions"`
}

// NoteParser describes the behavior required by each Markdown parsing strategy.
// Callers should depend only on this interface so they never need to know which
// concrete parser implementation is active.
type NoteParser interface {
	// Parse converts the given note into structured data. Implementations are
	// expected to be deterministic and must not mutate the input note.
	Parse(ctx context.Context, note *model.Note) (*ParsedNote, error)
}
