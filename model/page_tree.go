package model

import (
	"log/slog"

	"github.com/google/uuid"
)

// PageTree is a tree structure of blocks with a single root and unique ID
// TODO: define TEA arch functions for this guy
type PageTree struct {
	id   uuid.UUID
	Root *Block
}

// Create a new page, optionally with a parent block
func NewPageTreeWithParent(parent *Block) PageTree {
	root := NewBlockWithParent(parent)
	return PageTree{
		id:   uuid.New(),
		Root: &root,
	}
}

// Create a new page, optionally with a parent block
func NewPageTree() PageTree {
	rootBlock := NewBlock()
	p := PageTree{
		id:   uuid.New(),
		Root: rootBlock,
	}
	slog.Debug("Creating a new PageTree", "rootBlock", rootBlock, "pageTree", p)
	return p
}

// AddTopLevelBlock adds a new block as a child of the PageTree's root and
// returns a pointer to it
func (p *PageTree) AddTopLevelBlock() *Block {
	b := NewBlockWithParent(p.Root)
	return &b
}

func (p *PageTree) GetBlocksPreOrder() []Block {
	var result []Block

	// If root is nil, return empty slice
	if p.Root == nil {
		return result
	}

	// Helper function for recursive pre-order traversal
	var preOrderTraversal func(*Block)
	preOrderTraversal = func(block *Block) {
		if block == nil {
			return
		}

		// Visit current node (add to result)
		result = append(result, *block)

		// Recursively visit all children
		for _, child := range block.children {
			preOrderTraversal(child)
		}
	}

	// Start traversal from root
	preOrderTraversal(p.Root)

	return result
}

func (p *PageTree) GetTopLevelBlocks() []*Block {
	var result []*Block
	// If root is nil, return empty slice
	if p.Root == nil {
		return result
	}
	result = append(result, p.Root.children...)
	return result
}
