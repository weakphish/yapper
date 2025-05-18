use ratatui::buffer::Buffer;
use ratatui::layout::Rect;
use ratatui::prelude::Widget;
use tree_ds::prelude::Node;
use tui_tree_widget::{Tree, TreeItem, TreeState};
use uuid::Uuid;
use yap_core::block::Block;
use yap_core::block_tree::BlockTree;

#[derive(Debug, Default)]
pub struct Page {
    block_tree: BlockTree,
    tree_state: TreeState<Uuid>,
}

impl Page {
    pub fn new() -> Self {
        Page {
            block_tree: BlockTree::new(""), // start with a blank root block
            tree_state: TreeState::default(),
        }
    }
    /// Get a representation of the internal tree, which uses BlockTree from the core library,
    /// as a TreeItem pointed at the root.
    pub fn blocks_as_tree_item(&self) -> TreeItem<Uuid> {
        /// Recursive function to create tree widget items from a core library block tree.
        /// Essentially an adapter to convert tree representation.
        fn make_tree_item(page: &Page, parent: Node<Uuid, Block>) -> TreeItem<Uuid> {
            let root_node = page.block_tree.get_root();
            let root_node_content: String = root_node.get_value().unwrap().content().into(); // TODO: better way to get content?
            let children = root_node.get_children_ids();
            // base case
            if children.is_empty() {
                return TreeItem::new_leaf(root_node.get_node_id(), root_node_content);
            }
            let children_items = children
                .iter()
                .map(|child_id| {
                    let child_node = page.block_tree.get_node_by_id(child_id).unwrap();
                    make_tree_item(page, child_node)
                })
                .collect();

            TreeItem::new(root_node.get_node_id(), root_node_content, children_items).unwrap()
        }

        make_tree_item(self, self.block_tree.get_root())
    }

    pub fn state(&self) -> &TreeState<Uuid> {
        &self.tree_state
    }

    pub fn add_block(&mut self, block: Block) {
        self.block_tree.add_block(block);
    }
}

impl Widget for &Page {
    fn render(self, area: Rect, buf: &mut Buffer) {
        let page_as_tree_items = [self.blocks_as_tree_item()];
        let tree_widg = Tree::new(&page_as_tree_items).unwrap();
        // defer to tree widget
        tree_widg.render(area, buf)
    }
}
