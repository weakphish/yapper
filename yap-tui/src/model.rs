use crate::message::Message;
use crate::page::Page;
use crate::state::ViewState;
use ratatui::{
    crossterm::event::{self, Event}, DefaultTerminal,
    Frame,
};
use std::io;
use std::time::Duration;

#[derive(Debug, Default)]
pub struct Model {
    view_state: ViewState,
    block_tree: Page,
    exit: bool,
}

/// TEA Architecture reference: https://ratatui.rs/concepts/application-patterns/the-elm-architecture/
/// Runs the application's main loop until the user quits
pub fn run(model: &mut Model, terminal: &mut DefaultTerminal) -> io::Result<()> {
    while model.view_state != ViewState::Exit {
        // Render current view
        terminal.draw(|frame| view(&model, frame))?;

        // Handle events and map to a Message
        // TODO: can this be modified to be non-mutable?
        let mut current_message = handle_event(&model)?;

        // Process updates as long as they return a non-None message
        while current_message.is_some() {
            current_message = update(&mut model, current_message);
        }
    }
    Ok(())
}

fn view(model: &Model, frame: &mut Frame) {
    todo!()
    // frame.render_widget(
    //     Paragraph::new(format!("Counter: {}", model.counter)),
    //     frame.area(),
    // );
}

/// Convert Event to Message
fn handle_event(_: &Model) -> io::Result<()> {
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
        // KeyCode::Char('q') => Some(Message::Quit),
        _ => None,
    }
}

fn update(model: &mut Model, msg: Message) -> Option<Message> {
    todo!()
}
