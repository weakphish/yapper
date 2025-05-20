use std::rc::Rc;

use crate::message::Message;
use crate::state::ViewState;
use ratatui::layout::{Constraint, Direction, Layout, Rect};
use ratatui::style::{Color, Style};
use ratatui::widgets::{BorderType, Borders, Paragraph};
use ratatui::Frame;
use yap_core::block::{self, Block};
use yap_core::block_tree::BlockTree;

#[derive(Debug)]
pub struct Model<'a> {
    pub view_state: ViewState,
    pub block_tree: BlockTree,
    pub current_block: Option<&'a Block>,
    pub exit: bool,
}

impl <'a> Model<'a> {
    /// Create a new model instance defaulting to the daily page view and no current block
    /// selection.
    /// TODO: Model should own the block tree, right? what about when we switch pages?
    /// thought: the block tree is app-global, each root is a page/tag?
    pub fn new(block_tree: BlockTree) -> Self {
        Model{
            view_state: ViewState::DailyPage,
            block_tree, 
            current_block: None,
            exit: false
        }
    }
}

/// Take the model and a frame and render the view with the given state
pub fn view(model: &Model, frame: &mut Frame) {
    let paragraph_blocks = model.block_tree.get_tree_pre_order().iter().map(|id| {
        let block = model.block_tree.get_block_by_id(id).unwrap();
        let content = block.content();
        let p = Paragraph::new(content.to_string().clone())
        .style(Style::default().fg(Color::Yellow))
        .block(
            ratatui::widgets::Block::default()
                .borders(Borders::ALL)
                .title("Title")
                .border_type(BorderType::Rounded)
        );
        frame.render_widget(p, frame.area());
    });
}

/// Get layout for a page based on how many blocks there are
fn get_layout_based_on_blocks(block_count: usize, f: &Frame) -> Rc<[Rect]> {
    let block_percentage = if block_count > 50 { 80 } else { 50 };

    Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage(block_percentage),
            Constraint::Percentage(100 - block_percentage),
        ])
        .split(f.area())
}

/// Take the model and update it based on the message received
/// TODO: Make the model non-mutable, if possible
pub fn update(model: &mut Model, msg: Message) -> Option<Message> {
    match msg {
        Message::Quit => {
            // You can handle cleanup and exit here
            model.view_state = ViewState::Exit;
        }
        Message::EditBlock => todo!(),
    };
    None
}
