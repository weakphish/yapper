use std::{io, time::Duration};

use ratatui::crossterm::event::{self, Event, KeyCode};

use crate::{message::Message, model::Model};

/// Convert Event (crossterm) to Message (our TEA architecture)
pub fn handle_event(_: &Model) -> io::Result<Option<Message>> {
    if event::poll(Duration::from_millis(250))? {
        if let Event::Key(key) = event::read()? {
            if key.kind == event::KeyEventKind::Press {
                return Ok(handle_key(key));
            }
        }
    }
    Ok(None)
}

/// Helper - convert keycode to Message
pub fn handle_key(key: event::KeyEvent) -> Option<Message> {
    match key.code {
        KeyCode::Char('q') => Some(Message::Quit),
        _ => None,
    }
}
