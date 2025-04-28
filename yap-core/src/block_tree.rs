use crate::block::Block;
use ropey::Rope;
use tree_ds::prelude::{Node, TraversalStrategy, Tree};
use uuid::Uuid;

#[derive(Debug, Default)]
pub struct BlockTree {
    page_tree: Tree<Uuid, Block>,
}

impl BlockTree {
    pub fn new(root_text: &str) -> Self {
        let mut new = Self {
            page_tree: Tree::new(Some("")),
        };
        // FIXME: do we want the top level root to be hidden?
        let root_block = Block::new(Rope::from_str(root_text), Uuid::now_v7());
        new.page_tree
            .add_node(Node::new(Uuid::now_v7(), Some(root_block)), None)
            .unwrap();
        new
    }

    pub fn get_root(&self) -> Node<Uuid, Block> {
        match self.page_tree.get_root_node() {
            Some(node) => node,
            None => panic!("Page has no root node!"),
        }
    }

    /// Add a top-level block to the page
    pub fn add_block(&mut self, block: Block) {
        // Using UUID v4 for node ID and v7 for block ID - from docs:
        // "If you just want to generate unique identifiers then consider version 4 (v4) UUIDs...
        // ... If you want to use UUIDs as database keys or need to sort them then consider
        // version 7 (v7) UUIDs."
        let root = self.get_root();
        self.page_tree
            .add_node(
                Node::new(Uuid::new_v4(), Some(block)),
                Some(&root.get_node_id()),
            )
            .unwrap();
    }

    /// The client is going to need a traversal to go up/down the page. Get a vec of node ID.
    /// NOTE: tree node ID is separate from block ID, since the block may be moved around.
    /// Returns in pre-order.
    pub fn get_tree_pre_order(&self) -> Vec<Uuid> {
        let root = self.get_root();
        self.page_tree
            .traverse(&root.get_node_id(), TraversalStrategy::PreOrder)
            .unwrap()
    }

    pub fn get_block_by_id(&self, id: &Uuid) -> Option<Block> {
        match self.page_tree.get_node_by_id(id) {
            Some(node) => node.get_value(),
            None => None,
        }
    }

    pub fn get_node_by_id(&self, id: &Uuid) -> Option<Node<Uuid, Block>> {
        self.page_tree.get_node_by_id(id)
    }

    pub fn get_node_children(&self, id: &Uuid) -> Vec<Uuid> {
        let node = self.get_node_by_id(id).unwrap(); // FIXME: handle error
        node.get_children_ids()
    }
}
