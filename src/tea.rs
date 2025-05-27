use crate::model::{Message, Model, Update};
use ratatui::crossterm::event;
use ratatui::crossterm::event::{Event, KeyCode};
use ratatui::Frame;
use std::time::Duration;

/// "A key feature of TEA is immutability.
/// Hence, the update function should avoid direct mutation of the model.
/// Instead, it should produce a new instance of the model reflecting the desired changes."
pub fn update(model: &Model, msg: Message) -> Update {
    match msg {
        Message::Quit => todo!(),
        Message::MoveDown | Message::MoveUp | Message::AddBlock => todo!(),
    }
}

pub fn view(model: &Model, frame: &mut Frame) {
    //... use `ratatui` functions to draw your UI based on the model's state
}

/// Convert Event to Message
///
/// We don't need to pass in a `model` to this function in this example
/// but you might need it as your project evolves
pub fn handle_event(_: &Model) -> color_eyre::Result<Option<Message>> {
    if event::poll(Duration::from_millis(250))? {
        if let Event::Key(key) = event::read()? {
            if key.kind == event::KeyEventKind::Press {
                return Ok(handle_key(key));
            }
        }
    }
    Ok(None)
}

fn handle_key(key: event::KeyEvent) -> Option<Message> {
    match key.code {
        KeyCode::Char('q') => Some(Message::Quit),
        _ => None,
    }
}
