use crate::message::Message;
use crate::state::ViewState;
use ratatui::{
    Frame,
    crossterm::event::{self, Event, KeyCode},
};

#[derive(Debug, Default)]
pub struct Model {
    pub view_state: ViewState,
    pub exit: bool,
}

/// Take the model and a frame and render as a widget
pub fn view(model: &Model, frame: &mut Frame) {
    //frame.render_widget(&model, frame.area());
}

/// Take the model and update it based on the message received
/// TODO: Make the model non-mutable, if possible
pub fn update(model: &mut Model, msg: Message) -> Option<Message> {
    match msg {
        Message::Quit => {
            // You can handle cleanup and exit here
            model.view_state = ViewState::Exit;
        }
    };
    None
}
