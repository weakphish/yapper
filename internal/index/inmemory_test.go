package index

import (
	"context"
	"testing"
	"time"

	"github.com/weakphish/yapper/internal/model"
	"github.com/weakphish/yapper/internal/parser"
)

func TestInMemoryIndexStoreUpsertQueryAndRemove(t *testing.T) {
	store := NewInMemoryIndexStore()
	ctx := context.Background()

	note := &model.Note{
		ID:    "note-1",
		Path:  "note-1.md",
		Title: "Note 1",
		Date:  time.Now().UTC(),
	}
	task := model.Task{
		ID:        "task-1",
		NoteID:    note.ID,
		Title:     "Finish phase 3",
		Status:    model.TaskStatusInProgress,
		Tags:      []string{"Work", "PROJECT/Yapper"},
		CreatedAt: time.Now().UTC(),
		Line:      10,
	}
	logEntry := model.LogEntry{
		ID:        "log-1",
		NoteID:    note.ID,
		Line:      20,
		Timestamp: time.Now().UTC(),
		Content:   "Mentioned task in log",
		Tags:      []string{"Work"},
		TaskRefs:  []model.TaskID{task.ID},
	}
	mention := model.TaskMention{
		TaskID:  task.ID,
		NoteID:  note.ID,
		Line:    30,
		Context: "Follow up on #Work task",
		Tags:    []string{"WORK"},
	}

	parsed := &parser.ParsedNote{
		Note:       note,
		Tasks:      []model.Task{task},
		LogEntries: []model.LogEntry{logEntry},
		Mentions:   []model.TaskMention{mention},
	}

	if err := store.UpsertParsedNote(ctx, parsed); err != nil {
		t.Fatalf("UpsertParsedNote() error = %v", err)
	}

	gotTask, ok, err := store.GetTask(ctx, task.ID)
	if err != nil || !ok {
		t.Fatalf("GetTask() error = %v, ok=%v", err, ok)
	}
	if gotTask.Title != task.Title {
		t.Fatalf("GetTask() title = %s, want %s", gotTask.Title, task.Title)
	}

	tasks, err := store.ListTasks(ctx, TaskFilter{Statuses: []model.TaskStatus{model.TaskStatusInProgress}})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Fatalf("ListTasks() = %+v, want task %s", tasks, task.ID)
	}

	notes, err := store.ListNotes(ctx, NoteFilter{})
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 1 || notes[0].ID != note.ID {
		t.Fatalf("ListNotes() = %+v, want note %s", notes, note.ID)
	}

	entries, err := store.GetLogEntriesForTask(ctx, task.ID)
	if err != nil || len(entries) != 1 {
		t.Fatalf("GetLogEntriesForTask() error = %v len=%d", err, len(entries))
	}
	mentions, err := store.GetMentionsForTask(ctx, task.ID)
	if err != nil || len(mentions) != 1 {
		t.Fatalf("GetMentionsForTask() error = %v len=%d", err, len(mentions))
	}

	tags, err := store.ListTags(ctx)
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}
	wantTags := []string{"project/yapper", "work"}
	if len(tags) != len(wantTags) {
		t.Fatalf("ListTags() = %+v, want %v", tags, wantTags)
	}
	for i, tag := range wantTags {
		if tags[i] != tag {
			t.Fatalf("ListTags() sorted order mismatch: got %v want %v", tags, wantTags)
		}
	}

	items, ok, err := store.ItemsForTag(ctx, "work")
	if err != nil || !ok {
		t.Fatalf("ItemsForTag() error = %v ok=%v", err, ok)
	}
	if len(items.Tasks) != 1 || len(items.LogEntries) != 1 || len(items.Mentions) != 1 {
		t.Fatalf("ItemsForTag() unexpected counts: %+v", items)
	}

	if err := store.RemoveNote(ctx, note.ID); err != nil {
		t.Fatalf("RemoveNote() error = %v", err)
	}
	tasks, err = store.ListTasks(ctx, TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks() after remove error = %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("ListTasks() after remove = %+v, want empty", tasks)
	}

	_, ok, err = store.ItemsForTag(ctx, "work")
	if err != nil {
		t.Fatalf("ItemsForTag() after remove err = %v", err)
	}
	if ok {
		t.Fatalf("ItemsForTag() after remove returned ok=true, want false")
	}
}
