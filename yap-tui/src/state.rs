use crate::message::Message;

#[derive(Debug, Default, PartialEq, Eq)]
pub enum ViewState {
    #[default]
    DailyPage,
    TagPage,
    EditBlock,
    Exit,
}

pub struct StateMachine {
    active_state: ViewState,
    previous_states: Vec<ViewState>,
}

impl StateMachine {
    pub fn new() -> Self {
        Self {
            active_state: ViewState::DailyPage,
            previous_states: vec![],
        }
    }

    pub fn get_active_state(&self) -> &ViewState {
        &self.active_state
    }

    pub fn handle_message(&mut self, msg: Message) {
        match msg {
            Message::Quit => self.active_state = ViewState::Exit,
        }
    }
}
