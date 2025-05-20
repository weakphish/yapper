use crate::message::Message;

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub enum ViewState {
    #[default]
    DailyPage,
    TagPage,
    EditBlock,
    Exit,
}

#[derive(Debug)]
pub struct StateMachine {
    active_state: ViewState,
    previous_state: ViewState,
}

impl StateMachine {
    pub fn new() -> Self {
        Self {
            active_state: ViewState::DailyPage,
            previous_state: ViewState::DailyPage,
        }
    }

    pub fn get_active_state(&self) -> &ViewState {
        &self.active_state
    }

    pub fn handle_message(&mut self, msg: Message) {
        match msg {
            Message::Quit => self.active_state = ViewState::Exit,
            Message::EditBlock => {
                // if we are in DailyPage, edit block
                if self.active_state == ViewState::DailyPage {
                    self.previous_state = self.active_state.clone();
                    self.active_state = ViewState::EditBlock;
                }
            }
        }
    }
}
