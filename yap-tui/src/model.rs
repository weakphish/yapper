use crate::message::Message;
use crate::state::ViewState;
use ratatui::style::{Color, Style};
use ratatui::widgets::{BorderType, Borders, Paragraph};
use ratatui::Frame;
use yap_core::block::Block;

#[derive(Debug)]
pub struct Model<'a> {
    pub view_state: ViewState,
    pub current_block: Option<&'a Block>,
    pub exit: bool,
}

impl <'a> Model<'a> {
    /// Create a new model instance defaulting to the daily page view and no current block
    /// selection
    pub fn new() -> Self {
        Model{
            view_state: ViewState::DailyPage,
            current_block: None,
            exit: false
        }
    }
}

/// Take the model and a frame and render the view with the given state
pub fn view(model: &Model, frame: &mut Frame) {
    // TODO: for now, just render a single block
    let p = Paragraph::new("Hello, World!")
    .style(Style::default().fg(Color::Yellow))
    .block(
        ratatui::widgets::Block::default()
            .borders(Borders::ALL)
            .title("Title")
            .border_type(BorderType::Rounded)
    );
    frame.render_widget(p, frame.area());
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
