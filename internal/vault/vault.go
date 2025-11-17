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

	"github.com/weakphish/yapper/internal/model"
)

// Vault abstracts the storage location that holds Markdown notes. Keeping the
// interface narrow makes it straightforward to swap local filesystem access for
// future remote implementations.
type Vault interface {
	// Root returns the normalized path representing the vault root.
	Root() string
	// ListNotePaths enumerates every Markdown file beneath the vault. Paths are
	// returned relative to the root and sorted deterministically.
	ListNotePaths(ctx context.Context) ([]string, error)
	// LoadNote reads a specific Markdown file (relative to Root) into a Note.
	LoadNote(ctx context.Context, path string) (*model.Note, error)
	// LoadNotes eagerly materializes every note in the vault.
	LoadNotes(ctx context.Context) ([]*model.Note, error)
}

// FileSystemVault implements Vault using a directory tree on the local disk.
type FileSystemVault struct {
	root string
}

// NewFileSystemVault ensures the provided root is a directory and returns a
// Vault implementation that reads notes from that path.
func NewFileSystemVault(root string) (*FileSystemVault, error) {
	if root == "" {
		return nil, errors.New("vault root cannot be empty")
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("vault root %q is not a directory", root)
	}
	return &FileSystemVault{root: filepath.Clean(root)}, nil
}

// Root returns the normalized vault root path.
func (v *FileSystemVault) Root() string {
	return v.root
}

// ListNotePaths walks the vault and records every Markdown file.
func (v *FileSystemVault) ListNotePaths(ctx context.Context) ([]string, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}

	var paths []string
	err := filepath.WalkDir(v.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(entry.Name()) != ".md" {
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

// LoadNote reads a Markdown file into a Note model.
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

	content := string(data)
	note := &model.Note{
		ID:      model.NoteID(filepath.ToSlash(cleanPath)),
		Path:    filepath.ToSlash(cleanPath),
		Title:   deriveTitle(content, cleanPath),
		Date:    info.ModTime().UTC(),
		Content: content,
	}

	return note, nil
}

// LoadNotes eagerly loads every note in the vault. This helper keeps simple
// prototypes clean at the cost of more I/O.
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

func (v *FileSystemVault) normalizePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path cannot be empty")
	}
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("path %q escapes vault root", path)
	}
	return clean, nil
}

func ensureContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	return ctx.Err()
}

func deriveTitle(contents, path string) string {
	for _, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" {
		return "Untitled"
	}
	return base
}
