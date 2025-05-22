package model

import "github.com/google/uuid"

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

func NewBlockWithParent(parent *Block) Block {
	return Block{
		id:            uuid.New(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        parent,
		children:      []*Block{},
		content:       "",
	}
}

func NewBlock() Block {
	return Block{
		id:            uuid.New(),
		dependentIds:  []uuid.UUID{},
		dependencyIds: []uuid.UUID{},
		parent:        nil,
		children:      []*Block{},
		content:       "",
	}
}
