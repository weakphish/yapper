package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// NoteID uniquely identifies a note on disk.
type NoteID string

// TaskID is the stable identifier for a task (e.g. T-2025-001).
type TaskID string

// LogEntryID ties a log entry back to its originating note/line.
type LogEntryID string

// TaskStatus captures the canonical lifecycle states for tasks.
type TaskStatus string

const (
	TaskStatusOpen       TaskStatus = "Open"
	TaskStatusInProgress TaskStatus = "InProgress"
	TaskStatusDone       TaskStatus = "Done"
	TaskStatusBlocked    TaskStatus = "Blocked"
)

// Date wraps time.Time to ensure YYYY-MM-DD serialization.
type Date struct {
	time.Time
}

// NewDate returns a Date truncated to midnight UTC for stability.
func NewDate(t time.Time) Date {
	if t.IsZero() {
		return Date{}
	}
	y, m, d := t.Date()
	return Date{Time: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

// ParseDate creates a Date from a YYYY-MM-DD string.
func ParseDate(value string) (Date, error) {
	if value == "" {
		return Date{}, fmt.Errorf("empty date")
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return Date{}, err
	}
	return NewDate(t), nil
}

// MarshalJSON renders the date in YYYY-MM-DD form or null when empty.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format("2006-01-02"))
}

// UnmarshalJSON parses the canonical YYYY-MM-DD representation.
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*d = Date{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// String returns the formatted YYYY-MM-DD date, empty when unset.
func (d Date) String() string {
	if d.Time.IsZero() {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

// Note contains the raw markdown plus metadata derived from the filesystem.
type Note struct {
	ID      NoteID `json:"id"`
	Path    string `json:"path"`
	Title   string `json:"title"`
	Date    *Date  `json:"date,omitempty"`
	Content string `json:"content"`
}

// Task represents a Markdown checkbox plus metadata tracked by the index.
type Task struct {
	ID            TaskID     `json:"id"`
	Title         string     `json:"title"`
	Status        TaskStatus `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ClosedAt      *time.Time `json:"closed_at,omitempty"`
	Tags          []string   `json:"tags"`
	DescriptionMD *string    `json:"description_md,omitempty"`
	SourceNoteID  *NoteID    `json:"source_note_id,omitempty"`
}

// LogEntry models a single bullet inside the ## Log section.
type LogEntry struct {
	ID         LogEntryID `json:"id"`
	NoteID     NoteID     `json:"note_id"`
	LineNumber int        `json:"line_number"`
	Timestamp  *string    `json:"timestamp,omitempty"`
	ContentMD  string     `json:"content_md"`
	Tags       []string   `json:"tags"`
	TaskIDs    []TaskID   `json:"task_ids"`
}

// TaskMention ties a snippet of Markdown back to a referenced task.
type TaskMention struct {
	TaskID     TaskID      `json:"task_id"`
	NoteID     NoteID      `json:"note_id"`
	LogEntryID *LogEntryID `json:"log_entry_id,omitempty"`
	Excerpt    string      `json:"excerpt"`
}

// ParsedNote captures the structured data extracted from a note.
type ParsedNote struct {
	Note       Note          `json:"note"`
	Tasks      []Task        `json:"tasks"`
	LogEntries []LogEntry    `json:"log_entries"`
	Mentions   []TaskMention `json:"mentions"`
}

// NoteMeta contains lightweight note info when full content is unnecessary.
type NoteMeta struct {
	ID    NoteID `json:"id"`
	Path  string `json:"path"`
	Title string `json:"title"`
	Date  *Date  `json:"date,omitempty"`
}

// VaultIndex mirrors the Rust implementation for the in-memory backend.
type VaultIndex struct {
	Notes             map[NoteID]NoteMeta
	NoteContent       map[NoteID]Note
	Tasks             map[TaskID]Task
	LogEntries        map[LogEntryID]LogEntry
	MentionsByTask    map[TaskID][]TaskMention
	TagsToTasks       map[string][]TaskID
	TagsToLogEntries  map[string][]LogEntryID
	TaskRefsByTask    map[TaskID][]LogEntryID
	NoteToTaskIDs     map[NoteID][]TaskID
	NoteToLogEntryIDs map[NoteID][]LogEntryID
}

// NewVaultIndex allocates the nested maps for the in-memory store.
func NewVaultIndex() VaultIndex {
	return VaultIndex{
		Notes:             make(map[NoteID]NoteMeta),
		NoteContent:       make(map[NoteID]Note),
		Tasks:             make(map[TaskID]Task),
		LogEntries:        make(map[LogEntryID]LogEntry),
		MentionsByTask:    make(map[TaskID][]TaskMention),
		TagsToTasks:       make(map[string][]TaskID),
		TagsToLogEntries:  make(map[string][]LogEntryID),
		TaskRefsByTask:    make(map[TaskID][]LogEntryID),
		NoteToTaskIDs:     make(map[NoteID][]TaskID),
		NoteToLogEntryIDs: make(map[NoteID][]LogEntryID),
	}
}

// TaskFilter narrows down the task listing.
type TaskFilter struct {
	Status       *TaskStatus `json:"status,omitempty"`
	Tags         []string    `json:"tags,omitempty"`
	TextSearch   *string     `json:"text_search,omitempty"`
	TouchedSince *Date       `json:"touched_since,omitempty"`
}

// DateRange bounds date-scoped queries.
type DateRange struct {
	Start Date `json:"start"`
	End   Date `json:"end"`
}

// TagResult bundles tasks and log entries for a tag lookup.
type TagResult struct {
	Tag        string     `json:"tag"`
	Tasks      []Task     `json:"tasks"`
	LogEntries []LogEntry `json:"log_entries"`
}

// TagCount represents aggregated tag usage.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// WeeklySummary mirrors the spec for review workflows.
type WeeklySummary struct {
	NewTasks       []Task     `json:"new_tasks"`
	CompletedTasks []Task     `json:"completed_tasks"`
	Notes          []NoteMeta `json:"notes"`
	TopTags        []TagCount `json:"top_tags"`
}
