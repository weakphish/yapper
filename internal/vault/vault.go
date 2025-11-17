package vault

import (
	"context"
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
