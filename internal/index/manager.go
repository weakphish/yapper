package index

import (
	"context"
	"errors"
	"fmt"

	"github.com/weakphish/yapper/internal/parser"
	"github.com/weakphish/yapper/internal/vault"
)

// VaultIndexManager coordinates the vault, parser, and index store. It owns the
// logic for indexing a vault in bulk or refreshing a single note.
type VaultIndexManager struct {
	vault  vault.Vault
	parser parser.NoteParser
	store  IndexStore
}

// NewVaultIndexManager wires together the collaborating components. Each
// dependency must be non-nil.
func NewVaultIndexManager(v vault.Vault, p parser.NoteParser, store IndexStore) (*VaultIndexManager, error) {
	if v == nil || p == nil || store == nil {
		return nil, errors.New("vault, parser, and store are required")
	}
	return &VaultIndexManager{
		vault:  v,
		parser: p,
		store:  store,
	}, nil
}

// FullReindex scans the entire vault, parses each note, and upserts the
// resulting structured data into the index.
func (m *VaultIndexManager) FullReindex(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	paths, err := m.vault.ListNotePaths(ctx)
	if err != nil {
		return fmt.Errorf("list notes: %w", err)
	}
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := m.indexSingle(ctx, path); err != nil {
			return err
		}
	}
	return nil
}

// ReindexNote refreshes the indexed representation for a specific note path.
func (m *VaultIndexManager) ReindexNote(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.indexSingle(ctx, path)
}

func (m *VaultIndexManager) indexSingle(ctx context.Context, path string) error {
	note, err := m.vault.LoadNote(ctx, path)
	if err != nil {
		return fmt.Errorf("load note %q: %w", path, err)
	}
	parsed, err := m.parser.Parse(ctx, note)
	if err != nil {
		return fmt.Errorf("parse note %q: %w", path, err)
	}
	if err := m.store.UpsertParsedNote(ctx, parsed); err != nil {
		return fmt.Errorf("index note %q: %w", path, err)
	}
	return nil
}
