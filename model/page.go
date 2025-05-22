package model

import "github.com/google/uuid"

// Page is a tree structure of blocks with a single root and unique ID
// TODO: define TEA arch functions for this guy
type Page struct {
	id   uuid.UUID
	root *Block
}

// Create a new page, optionally with a parent block
func NewPageWithParent(parent *Block) Page {
	root := NewBlockWithParent(parent)
	return Page{
		id:   uuid.New(),
		root: &root,
	}
}

// Create a new page, optionally with a parent block
func NewPage() Page {
	return Page{
		id:   uuid.New(),
		root: nil,
	}
}
