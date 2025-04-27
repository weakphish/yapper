use crate::block::Block;
use tree_ds::prelude::{Node, TraversalStrategy, Tree};

pub struct Page {
    page_tree: Tree<u32, Block>,
}

impl Page {
    pub fn new() -> Self {
        Self {
            page_tree: Tree::new(Some("")),
        }
    }

    fn get_root(&self) -> Node<u32, Block> {
        match self.page_tree.get_root_node() {
            Some(node) => node,
            None => panic!("Page has no root node!"),
        }
    }

    /// Add a top-level block to the page
    pub fn add_block(&mut self, block: Block) {
        let root = self.get_root();
        self.page_tree
            .add_node(Node::new(0, Some(block)), Some(&root.get_node_id()))
            .unwrap();
    }

    /// The client is going to need a traversal to go up/down the page. Get a vec of node ID.
    pub fn get_traversal(&self) -> Vec<u32> {
        let root = self.get_root();
        self.page_tree
            .traverse(&root.get_node_id(), TraversalStrategy::PreOrder)
            .unwrap()
    }
}
