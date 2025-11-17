package model

import (
	"time"
)

// NoteID uniquely identifies a note.
type NoteID string

// TaskID uniquely identifies a task.
type TaskID string

// LogEntryID uniquely identifies a log entry extracted from a note.
type LogEntryID string

// TaskStatus captures the lifecycle state for a task.
type TaskStatus string

const (
	// TaskStatusTodo indicates a task that has not been started.
	TaskStatusTodo TaskStatus = "todo"
	// TaskStatusInProgress indicates a task with work underway.
	TaskStatusInProgress TaskStatus = "in_progress"
	// TaskStatusBlocked indicates a task that is currently blocked.
	TaskStatusBlocked TaskStatus = "blocked"
	// TaskStatusDone indicates a completed task.
	TaskStatusDone TaskStatus = "done"
)

// Note is the source-of-truth representation for a Markdown file in the vault.
type Note struct {
	ID      NoteID    `json:"id"`
	Path    string    `json:"path"`
	Title   string    `json:"title"`
	Date    time.Time `json:"date"`
	Content string    `json:"content"`
}

// Task models a first-class task extracted from a note.
type Task struct {
	ID          TaskID     `json:"id"`
	NoteID      NoteID     `json:"note_id"`
	Title       string     `json:"title"`
	Status      TaskStatus `json:"status"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Line        int        `json:"line"`
}

// LogEntry captures structured log lines parsed from a note.
type LogEntry struct {
	ID        LogEntryID `json:"id"`
	NoteID    NoteID     `json:"note_id"`
	Line      int        `json:"line"`
	Timestamp time.Time  `json:"timestamp"`
	Content   string     `json:"content"`
	Tags      []string   `json:"tags"`
	TaskRefs  []TaskID   `json:"task_refs"`
}

// TaskMention links a task back to the source note that references it.
type TaskMention struct {
	TaskID  TaskID   `json:"task_id"`
	NoteID  NoteID   `json:"note_id"`
	Line    int      `json:"line"`
	Context string   `json:"context"`
	Tags    []string `json:"tags"`
}
