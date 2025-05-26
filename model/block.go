package model

import (
	"log/slog"

	"github.com/google/uuid"
)

type BlockType int

const (
	Note BlockType = iota
	Task
)

// Block is a element in the page tree structure that has both
// dependendies and dependents. It also has a single parent and
// N many children.
type Block struct {
	id            uuid.UUID
	dependentIds  []uuid.UUID
	dependencyIds []uuid.UUID
	parent        *Block
	children      []*Block // in order of render
	content       string
}

// NewBlockWithParent creates a block with a given block as the parent,
// and modifies the parent, adding the new block as a child
func NewBlockWithParent(parent *Block) Block {
	b := Block{
		id:            uuid.New(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        parent,
		children:      []*Block{},
		content:       "",
	}
	parent.children = append(parent.children, &b)
	return b
}

// NewBlock creates a basic block with no parent or children
func NewBlock() *Block {
	b := Block{
		id:            uuid.New(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        nil,
		children:      []*Block{},
		content:       "",
	}
	slog.Debug("Created a new block", "block", b)
	return &b
}

func (b *Block) GetContent() string {
	return b.content
}

// UpdateContent replaces a block's content with a new string
func (b *Block) UpdateContent(content string) {
	b.content = content
}

// RenderBlock returns a charm-compatible string that renders the
// block content
func (b *Block) RenderBlock() string {
	// TODO: render a block
	return b.content
}
