use crate::message::Message;
use crate::page::Page;
use crate::state::ViewState;
use ratatui::{
    DefaultTerminal, Frame,
    crossterm::event::{self, Event, KeyCode},
};
use std::io;
use std::time::Duration;

#[derive(Debug, Default)]
pub struct Model {
    view_state: ViewState,
    page: Page,
    exit: bool,
}

/// TEA Architecture reference: https://ratatui.rs/concepts/application-patterns/the-elm-architecture/
/// Runs the application's main loop until the user quits
pub fn run(mut model: &mut Model, terminal: &mut DefaultTerminal) -> io::Result<()> {
    model.page = Page::new();
    while model.view_state != ViewState::Exit {
        // Render current view
        terminal.draw(|frame| view(&model, frame))?;

        // Handle events and map to a Message
        // TODO: can this be modified to be non-mutable?
        let mut current_message = handle_event(&model)?;

        // Process updates as long as they return a non-None message
        while current_message.is_some() {
            current_message = update(&mut model, current_message.unwrap());
        }
    }
    Ok(())
}

fn view(model: &Model, frame: &mut Frame) {
    frame.render_widget(&model.page, frame.area());
}

/// Convert Event (crossterm) to Message (our TEA architecture)
fn handle_event(_: &Model) -> io::Result<Option<Message>> {
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
        KeyCode::Char('a') => Some(Message::AddBlock),
        _ => None,
    }
}

fn update(model: &mut Model, msg: Message) -> Option<Message> {
    match msg {
        Message::Quit => {
            // You can handle cleanup and exit here
            model.view_state = ViewState::Exit;
        }
        Message::AddBlock => todo!(),
        Message::MoveUpBlock => todo!(),
        Message::MoveDownBlock => todo!(),
        Message::SelectBlock => todo!(),
    };
    None
}
