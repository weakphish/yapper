package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Vault abstracts how Markdown notes are discovered and loaded.
type Vault interface {
	ListNotePaths() ([]string, error)
	ReadNote(path string) (Note, error)
	RootPath() string
}

// FileSystemVault walks the local filesystem to manage the vault.
type FileSystemVault struct {
	root string
}

// NewFileSystemVault creates a vault rooted at the given directory.
func NewFileSystemVault(root string) *FileSystemVault {
	return &FileSystemVault{root: root}
}

// RootPath implements Vault.
func (v *FileSystemVault) RootPath() string {
	return v.root
}

// ListNotePaths recursively discovers Markdown files, sorted for stability.
func (v *FileSystemVault) ListNotePaths() ([]string, error) {
	var paths []string
	err := filepath.WalkDir(v.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list note paths: %w", err)
	}
	sort.Strings(paths)
	return paths, nil
}

// ReadNote loads the note content plus derived metadata.
func (v *FileSystemVault) ReadNote(path string) (Note, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Note{}, fmt.Errorf("read note %s: %w", path, err)
	}
	title := deriveTitle(path)
	date := deriveDate(path)
	var datePtr *Date
	if !date.Time.IsZero() {
		datePtr = &date
	}
	return Note{
		ID:      NoteID(path),
		Path:    path,
		Title:   title,
		Date:    datePtr,
		Content: string(content),
	}, nil
}

func deriveTitle(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}

func deriveDate(path string) Date {
	stem := deriveTitle(path)
	formats := []string{"2006-01-02", "06-01-02"}
	for _, layout := range formats {
		if t, err := time.Parse(layout, stem); err == nil {
			return NewDate(t)
		}
	}
	return Date{}
}
