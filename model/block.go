package model

import (
	"log/slog"
	"time"

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
	date          time.Time
	dependentIds  []uuid.UUID
	dependencyIds []uuid.UUID
	parent        *Block
	children      []*Block // in order of render
	content       string
	blockType     BlockType
	completed     bool // For tracking task completion status
}

func NewBlockWithParent(content string, blockType BlockType, parent *Block) Block {
	id := uuid.New()
	if parent == nil {
		slog.Warn("creating block with nil parent", "id", id.String())
		return NewNoteBlock(content)
	}
	if content == "" {
		slog.Warn("creating block with empty content", "id", id.String())
	}
	slog.Debug("creating new block with parent",
		"id", id.String(),
		"type", blockType,
		"content", content,
		"parent_id", parent.id.String(),
	)
	return Block{
		id:            id,
		date:          time.Now(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        parent,
		children:      []*Block{},
		content:       content,
		blockType:     blockType,
		completed:     false,
	}
}

func NewNoteBlock(content string) Block {
	id := uuid.New()
	slog.Debug("creating new note block",
		"id", id.String(),
		"content", content)
	// Always create a note block with Note type
	return Block{
		id:            id,
		date:          time.Now(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        nil,
		children:      []*Block{},
		content:       content,
		blockType:     Note, // Default to Note type
		completed:     false,
	}
}

// NewTaskBlock creates a new task block
func NewTaskBlock(content string) Block {
	id := uuid.New()

	slog.Debug("creating new task block",
		"id", id.String(),
	)

	// Always explicitly set the block type to Task
	taskBlock := Block{
		id:            id,
		date:          time.Now(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        nil,
		children:      []*Block{},
		content:       content,
		blockType:     Task, // Must be Task type
		completed:     false,
	}

	return taskBlock
}

func (b *Block) GetContent() string {
	slog.Debug("getting block content", "id", b.id.String())
	return b.content
}

// SetContent sets the block's content
func (b *Block) SetContent(content string) {
	if content == "" {
		slog.Warn("setting empty content for block", "id", b.id.String())
	}
	slog.Debug("setting block content",
		"id", b.id.String(),
		"old_content", b.content,
		"new_content", content,
	)
	b.content = content
}

// GetType returns the block's type
func (b *Block) GetType() BlockType {
	return b.blockType
}

// GetTypeString returns a string representation of the block's type
func (b *Block) GetTypeString() string {
	switch b.blockType {
	case Note:
		return "Note"
	case Task:
		return "Task"
	default:
		return "Unknown"
	}
}

// SetType sets the block's type
func (b *Block) SetType(blockType BlockType) {
	if b.blockType == blockType {
		slog.Warn("attempting to set block type to same value",
			"id", b.id.String(),
			"type", blockType)
		return
	}
	slog.Debug("changing block type",
		"id", b.id.String(),
		"old_type", b.blockType,
		"new_type", blockType,
	)
	b.blockType = blockType
}

// IsTask returns true if the block is a task
func (b *Block) IsTask() bool {
	result := b.blockType == Task
	slog.Debug("checking if block is task", "id", b.id.String(), "is_task", result, "block_type", b.GetTypeString())
	return result
}

// IsNote returns true if the block is a note
func (b *Block) IsNote() bool {
	result := b.blockType == Note
	slog.Debug("checking if block is note", "id", b.id.String(), "is_note", result)
	return result
}

// GetID returns the block's ID
func (b *Block) GetID() string {
	return b.id.String()
}
