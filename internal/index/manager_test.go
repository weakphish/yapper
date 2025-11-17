package index

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/weakphish/yapper/internal/model"
	"github.com/weakphish/yapper/internal/parser"
)

func TestVaultIndexManagerFullReindex(t *testing.T) {
	v := &fakeVault{
		paths: []string{"note-a.md", "note-b.md"},
		notes: map[string]*model.Note{
			"note-a.md": {ID: "a", Path: "note-a.md", Title: "A", Date: time.Now().UTC()},
			"note-b.md": {ID: "b", Path: "note-b.md", Title: "B", Date: time.Now().UTC()},
		},
	}
	store := &fakeStore{}
	manager, err := NewVaultIndexManager(v, &fakeParser{}, store)
	if err != nil {
		t.Fatalf("NewVaultIndexManager() error = %v", err)
	}
	if err := manager.FullReindex(context.Background()); err != nil {
		t.Fatalf("FullReindex() error = %v", err)
	}
	want := []string{"a", "b"}
	if !reflect.DeepEqual(store.upserted, want) {
		t.Fatalf("upserted = %v, want %v", store.upserted, want)
	}
}

func TestVaultIndexManagerReindexNoteError(t *testing.T) {
	v := &fakeVault{
		paths: []string{"note.md"},
		notes: map[string]*model.Note{
			"note.md": {ID: "n", Path: "note.md", Title: "N", Date: time.Now().UTC()},
		},
	}
	manager, err := NewVaultIndexManager(v, &fakeParser{err: errors.New("parse boom")}, &fakeStore{})
	if err != nil {
		t.Fatalf("NewVaultIndexManager() error = %v", err)
	}
	if err := manager.ReindexNote(context.Background(), "note.md"); err == nil {
		t.Fatalf("ReindexNote() expected error, got nil")
	}
}

type fakeVault struct {
	paths []string
	notes map[string]*model.Note
}

func (f *fakeVault) Root() string { return "/fake" }

func (f *fakeVault) ListNotePaths(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append([]string(nil), f.paths...), nil
}

func (f *fakeVault) LoadNote(ctx context.Context, path string) (*model.Note, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	note, ok := f.notes[path]
	if !ok {
		return nil, errors.New("missing note")
	}
	return note, nil
}

func (f *fakeVault) LoadNotes(ctx context.Context) ([]*model.Note, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var result []*model.Note
	for _, p := range f.paths {
		note, err := f.LoadNote(ctx, p)
		if err != nil {
			return nil, err
		}
		result = append(result, note)
	}
	return result, nil
}

type fakeParser struct {
	err error
}

func (f *fakeParser) Parse(ctx context.Context, note *model.Note) (*parser.ParsedNote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.err != nil {
		return nil, f.err
	}
	return &parser.ParsedNote{Note: note}, nil
}

type fakeStore struct {
	upserted []string
}

func (f *fakeStore) UpsertParsedNote(ctx context.Context, parsed *parser.ParsedNote) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.upserted = append(f.upserted, string(parsed.Note.ID))
	return nil
}

func (f *fakeStore) RemoveNote(ctx context.Context, noteID model.NoteID) error { return nil }

func (f *fakeStore) GetTask(ctx context.Context, id model.TaskID) (model.Task, bool, error) {
	return model.Task{}, false, nil
}

func (f *fakeStore) ListTasks(ctx context.Context, filter TaskFilter) ([]model.Task, error) {
	return nil, nil
}

func (f *fakeStore) GetLogEntriesForTask(ctx context.Context, id model.TaskID) ([]model.LogEntry, error) {
	return nil, nil
}

func (f *fakeStore) GetMentionsForTask(ctx context.Context, id model.TaskID) ([]model.TaskMention, error) {
	return nil, nil
}

func (f *fakeStore) ListNotes(ctx context.Context, filter NoteFilter) ([]model.Note, error) {
	return nil, nil
}

func (f *fakeStore) ListTags(ctx context.Context) ([]string, error) { return nil, nil }

func (f *fakeStore) ItemsForTag(ctx context.Context, tag string) (TagItems, bool, error) {
	return TagItems{}, false, nil
}
