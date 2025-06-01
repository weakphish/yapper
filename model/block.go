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
}

func NewBlockWithParent(content string, parent *Block) Block {
	return Block{
		id:            uuid.New(),
		date:          time.Now(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        parent,
		children:      []*Block{},
		content:       content,
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
	}
}

func (b *Block) GetContent() string {
	return b.content
}
