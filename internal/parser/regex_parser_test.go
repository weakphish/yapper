package parser

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/weakphish/yapper/internal/model"
)

// TestRegexNoteParser_Parse validates that tasks, logs, tags, and mentions are
// extracted from a representative Markdown note.
func TestRegexNoteParser_Parse(t *testing.T) {
	content := strings.TrimSpace(`
## Tasks
- [ ] Draft summary #work [T-1234]
- [x] Finish report #work
## Log
- 2024-05-01 Completed milestone #wins [T-1234]
- 2024-05-02 Followed up with #work #team [T-5678]
`)

	note := &model.Note{
		ID:      "daily/2024-05-02.md",
		Path:    "daily/2024-05-02.md",
		Title:   "Daily",
		Date:    time.Date(2024, 5, 2, 8, 30, 0, 0, time.UTC),
		Content: content,
	}

	parser := NewRegexNoteParser()
	result, err := parser.Parse(context.Background(), note)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}

	firstTask := result.Tasks[0]
	if firstTask.ID != "T-1234" {
		t.Fatalf("unexpected task ID: %+v", firstTask)
	}
	if firstTask.Status != model.TaskStatusTodo {
		t.Fatalf("expected todo status, got %s", firstTask.Status)
	}
	if firstTask.Title != "Draft summary" {
		t.Fatalf("title mismatch: %q", firstTask.Title)
	}
	if len(firstTask.Tags) != 1 || firstTask.Tags[0] != "work" {
		t.Fatalf("unexpected tags: %+v", firstTask.Tags)
	}
	if firstTask.Line != 2 {
		t.Fatalf("expected line 2, got %d", firstTask.Line)
	}

	secondTask := result.Tasks[1]
	if secondTask.ID != model.TaskID("daily/2024-05-02.md#3") {
		t.Fatalf("fallback ID mismatch: %s", secondTask.ID)
	}
	if secondTask.Status != model.TaskStatusDone {
		t.Fatalf("expected done status, got %s", secondTask.Status)
	}
	if secondTask.Title != "Finish report" {
		t.Fatalf("title mismatch: %q", secondTask.Title)
	}

	if len(result.LogEntries) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(result.LogEntries))
	}

	firstLog := result.LogEntries[0]
	if firstLog.Line != 5 {
		t.Fatalf("unexpected log line: %d", firstLog.Line)
	}
	if firstLog.Timestamp != time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("timestamp mismatch: %s", firstLog.Timestamp)
	}
	if firstLog.Content != "Completed milestone" {
		t.Fatalf("content mismatch: %q", firstLog.Content)
	}
	if len(firstLog.Tags) != 1 || firstLog.Tags[0] != "wins" {
		t.Fatalf("log tags mismatch: %+v", firstLog.Tags)
	}
	if len(firstLog.TaskRefs) != 1 || firstLog.TaskRefs[0] != "T-1234" {
		t.Fatalf("log refs mismatch: %+v", firstLog.TaskRefs)
	}

	if len(result.Mentions) != 2 {
		t.Fatalf("expected 2 mentions, got %d", len(result.Mentions))
	}
	if result.Mentions[0].TaskID != "T-1234" || result.Mentions[0].Line != 5 {
		t.Fatalf("unexpected mention: %+v", result.Mentions[0])
	}
	if result.Mentions[1].TaskID != "T-5678" || result.Mentions[1].Line != 6 {
		t.Fatalf("unexpected mention: %+v", result.Mentions[1])
	}
}
