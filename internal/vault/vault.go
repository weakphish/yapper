package vault

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/weakphish/yapper/internal/model"
)

// Vault defines the contract for loading Markdown notes from the source
// repository. Callers use this interface exclusively so the backing storage
// (filesystem, remote service, etc.) remains abstracted away.
type Vault interface {
	// Root returns the root directory of the vault.
	Root() string
	// ListNotePaths returns every Markdown file path within the vault. The
	// returned paths are relative to Root and sorted deterministically.
	ListNotePaths(ctx context.Context) ([]string, error)
	// LoadNote reads the Markdown file at the provided path (relative to Root)
	// into a Note model. Paths outside the vault are rejected.
	LoadNote(ctx context.Context, path string) (*model.Note, error)
	// LoadNotes loads and returns every note in the vault. This is a convenience
	// helper for callers that need the entire vault materialized at once and is
	// implemented in terms of ListNotePaths + LoadNote.
	LoadNotes(ctx context.Context) ([]*model.Note, error)
}

// FileSystemVault implements Vault by reading directly from the local
// filesystem. It expects Markdown notes to live beneath a single root directory.
type FileSystemVault struct {
	root string
}

// NewFileSystemVault constructs a FileSystemVault rooted at the provided path.
// The root must exist and be a directory.
func NewFileSystemVault(root string) (*FileSystemVault, error) {
	if root == "" {
		return nil, errors.New("vault root cannot be empty")
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat vault root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault root %q is not a directory", root)
	}

	return &FileSystemVault{root: filepath.Clean(root)}, nil
}

// Root returns the normalized root directory path for the vault.
func (v *FileSystemVault) Root() string {
	return v.root
}

// ListNotePaths walks the vault root and collects every Markdown file. Only
// files ending in ".md" are considered notes.
func (v *FileSystemVault) ListNotePaths(ctx context.Context) ([]string, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}

	var paths []string
	err := filepath.WalkDir(v.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		if err := ensureContext(ctx); err != nil {
			return err
		}

		rel, err := filepath.Rel(v.root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk vault: %w", err)
	}

	sort.Strings(paths)
	return paths, nil
}

// LoadNote loads a single Markdown file and converts it into a Note model.
func (v *FileSystemVault) LoadNote(ctx context.Context, path string) (*model.Note, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}

	cleanPath, err := v.normalizePath(path)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(v.root, cleanPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path %q is a directory", path)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	notePath := filepath.ToSlash(cleanPath)
	title := deriveTitle(string(data), notePath)
	modTime := info.ModTime().UTC()

	note := &model.Note{
		ID:      model.NoteID(notePath),
		Path:    notePath,
		Title:   title,
		Date:    modTime,
		Content: string(data),
	}

	return note, nil
}

// LoadNotes loads every Markdown note in the vault. It is primarily intended for
// simple tests and prototyping code that benefit from eagerly materializing the
// vault.
func (v *FileSystemVault) LoadNotes(ctx context.Context) ([]*model.Note, error) {
	paths, err := v.ListNotePaths(ctx)
	if err != nil {
		return nil, err
	}

	notes := make([]*model.Note, 0, len(paths))
	for _, p := range paths {
		if err := ensureContext(ctx); err != nil {
			return nil, err
		}
		note, err := v.LoadNote(ctx, p)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

// normalizePath ensures that the provided path points within the vault root and
// returns it in a cleaned, slash-separated format.
func (v *FileSystemVault) normalizePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) {
		rel, err := filepath.Rel(v.root, clean)
		if err != nil {
			return "", fmt.Errorf("path %q not within vault", path)
		}
		clean = rel
	}

	full := filepath.Join(v.root, clean)
	rel, err := filepath.Rel(v.root, full)
	if err != nil {
		return "", fmt.Errorf("path %q escapes vault root", path)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %q escapes vault root", path)
	}

	return clean, nil
}

// deriveTitle attempts to find a Markdown heading to use as the note title and
// falls back to the filename when no heading is present.
func deriveTitle(content string, fallback string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			trimmed := strings.TrimLeft(line, "#")
			trimmed = strings.TrimSpace(trimmed)
			if trimmed != "" {
				return trimmed
			}
		}
		if line != "" {
			break
		}
	}
	return filepath.Base(strings.TrimSuffix(fallback, filepath.Ext(fallback)))
}

// ensureContext translates nil contexts into context.Background and verifies
// cancellation before expensive work.
func ensureContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= 0 {
		return ctx.Err()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
