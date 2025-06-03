package model

import (
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
}

func NewBlockWithParent(content string, blockType BlockType, parent *Block) Block {
	return Block{
		id:            uuid.New(),
		date:          time.Now(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        parent,
		children:      []*Block{},
		content:       content,
		blockType:     blockType,
	}
}

func NewBlock(content string) Block {
	return Block{
		id:            uuid.New(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        nil,
		children:      []*Block{},
		content:       content,
		blockType:     Note, // Default to Note type
	}
}

// NewTaskBlock creates a new task block
func NewTaskBlock(content string) Block {
	return Block{
		id:            uuid.New(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        nil,
		children:      []*Block{},
		content:       content,
		blockType:     Task,
	}
}

func (b *Block) GetContent() string {
	return b.content
}

// SetContent sets the block's content
func (b *Block) SetContent(content string) {
	b.content = content
}

// GetType returns the block's type
func (b *Block) GetType() BlockType {
	return b.blockType
}

// SetType sets the block's type
func (b *Block) SetType(blockType BlockType) {
	b.blockType = blockType
}

// IsTask returns true if the block is a task
func (b *Block) IsTask() bool {
	return b.blockType == Task
}

// IsNote returns true if the block is a note
func (b *Block) IsNote() bool {
	return b.blockType == Note
}
