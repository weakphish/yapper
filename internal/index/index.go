package index

import (
	"context"
	"time"

	"github.com/weakphish/yapper/internal/model"
	"github.com/weakphish/yapper/internal/parser"
)

// TaskFilter allows callers to control the set of tasks returned by the index.
// Filters are optional; omitting every field results in all known tasks.
type TaskFilter struct {
	Statuses []model.TaskStatus
	Tags     []string
	NoteIDs  []model.NoteID
}

// NoteFilter constrains note listing queries to a time range. The range is
// inclusive and either boundary may be omitted.
type NoteFilter struct {
	Start *time.Time
	End   *time.Time
}

// TagItems describes every entity associated with a specific tag. Tags may
// appear on tasks, log entries, and mentions simultaneously.
type TagItems struct {
	Tag        string              `json:"tag"`
	Tasks      []model.Task        `json:"tasks"`
	LogEntries []model.LogEntry    `json:"log_entries"`
	Mentions   []model.TaskMention `json:"mentions"`
}

// IndexStore defines the required behavior for all index implementations.
// Implementations are responsible for keeping derived data synchronized with
// ParsedNotes emitted by the parser layer.
type IndexStore interface {
	// UpsertParsedNote replaces the indexed data for a note with the provided
	// parsed representation.
	UpsertParsedNote(ctx context.Context, parsed *parser.ParsedNote) error
	// RemoveNote deletes all indexed data produced from the note with the given
	// ID. Missing notes are ignored.
	RemoveNote(ctx context.Context, noteID model.NoteID) error
	// GetTask returns the task for the provided ID, if present.
	GetTask(ctx context.Context, id model.TaskID) (model.Task, bool, error)
	// ListTasks lists tasks that satisfy the provided filter.
	ListTasks(ctx context.Context, filter TaskFilter) ([]model.Task, error)
	// GetLogEntriesForTask returns the log entries referencing a given task.
	GetLogEntriesForTask(ctx context.Context, id model.TaskID) ([]model.LogEntry, error)
	// GetMentionsForTask returns the mentions pointing to a given task.
	GetMentionsForTask(ctx context.Context, id model.TaskID) ([]model.TaskMention, error)
	// ListNotes returns every note in the index ordered by descending date.
	ListNotes(ctx context.Context, filter NoteFilter) ([]model.Note, error)
	// ListTags returns the unique set of tags known to the index sorted
	// lexicographically.
	ListTags(ctx context.Context) ([]string, error)
	// ItemsForTag returns every indexed entity linked to the provided tag.
	ItemsForTag(ctx context.Context, tag string) (TagItems, bool, error)
}
