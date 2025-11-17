package vault

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestFileSystemVault_ListAndLoad verifies note discovery and loading semantics.
func TestFileSystemVault_ListAndLoad(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	write := func(rel, contents string) {
		path := filepath.Join(tmp, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	write("alpha.md", "# Alpha\nhello")
	write("nested/bravo.md", "## Heading\nbody")
	write("ignore.txt", "nope")

	v, err := NewFileSystemVault(tmp)
	if err != nil {
		t.Fatalf("NewFileSystemVault: %v", err)
	}

	paths, err := v.ListNotePaths(ctx)
	if err != nil {
		t.Fatalf("ListNotePaths: %v", err)
	}

	expected := []string{"alpha.md", "nested/bravo.md"}
	if len(paths) != len(expected) {
		t.Fatalf("unexpected path count: %v", paths)
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Fatalf("paths differ: got %v want %v", paths, expected)
		}
	}

	note, err := v.LoadNote(ctx, "alpha.md")
	if err != nil {
		t.Fatalf("LoadNote: %v", err)
	}
	if got := note.Title; got != "Alpha" {
		t.Fatalf("title mismatch: %q", got)
	}
	if note.ID != "alpha.md" || note.Path != "alpha.md" {
		t.Fatalf("unexpected identifiers: %+v", note)
	}
	if got := note.Content; got != "# Alpha\nhello" {
		t.Fatalf("content mismatch: %q", got)
	}
}

// TestFileSystemVault_LoadNotes ensures the helper loads all notes without
// requiring repeated calls to LoadNote.
func TestFileSystemVault_LoadNotes(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	noteFiles := map[string]string{
		"one.md":           "# One",
		"nested/two.md":    "## Two",
		"nested/three.txt": "ignore",
	}

	for path, contents := range noteFiles {
		full := filepath.Join(tmp, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	v, err := NewFileSystemVault(tmp)
	if err != nil {
		t.Fatalf("NewFileSystemVault: %v", err)
	}

	notes, err := v.LoadNotes(ctx)
	if err != nil {
		t.Fatalf("LoadNotes: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}

	gotPaths := []string{notes[0].Path, notes[1].Path}
	sort.Strings(gotPaths)
	if gotPaths[0] != "nested/two.md" || gotPaths[1] != "one.md" {
		t.Fatalf("unexpected paths: %v", gotPaths)
	}
}
