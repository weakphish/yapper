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

fn view(model: &Model, frame: &mut Frame) {
    //frame.render_widget(&model, frame.area());
}

fn update(model: &mut Model, msg: Message) -> Option<Message> {
    match msg {
        Message::Quit => {
            // You can handle cleanup and exit here
            model.view_state = ViewState::Exit;
        }
    };
    None
}
