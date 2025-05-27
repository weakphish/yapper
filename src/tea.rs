use crate::model::{Message, Model, Update};
use ratatui::crossterm::event;
use ratatui::crossterm::event::{Event, KeyCode};
use ratatui::prelude::{Color, Line, Text};
use ratatui::style::{Modifier, Style};
use ratatui::text::Span;
use ratatui::widgets::{Block, Borders, Paragraph};
use ratatui::Frame;
use std::time::Duration;

/// "A key feature of TEA is immutability.
/// Hence, the update function should avoid direct mutation of the model.
/// Instead, it should produce a new instance of the model reflecting the desired changes."
pub fn update(model: &Model, msg: Message) -> Update {
    let mut new_model = model.clone();
    match msg {
        Message::Quit => todo!(),
        Message::MoveDown => {
            // handle wrapping around the end of the list
            if model.block_cursor() + 1 > model.blocks().len() {
                new_model.set_block_cursor(0);
            } else {
                new_model.set_block_cursor(model.block_cursor() + 1);
            }
        }
        Message::MoveUp => {
            // handle wrapping from the beginning of the list to the end of the list
            if model.block_cursor() == 0 {
                new_model.set_block_cursor(model.blocks().len() - 1);
            } else {
                new_model.set_block_cursor(model.block_cursor() - 1);
            }
        }
        Message::AddBlock => todo!(),
    }
    Update::new(new_model, None)
}

pub fn view(model: &Model, frame: &mut Frame) {
    // Edit block
    // draw a ratatui Block that will contain the popup with a title
    if let Some(editing) = &model.editing_block() {
        let popup_block = Block::default()
            .title("Editing block")
            .borders(Borders::NONE)
            .style(Style::default().bg(Color::DarkGray));

        let area = centered_rect(60, 25, frame.area());
        frame.render_widget(popup_block, area);

    // Render blocks
    model.blocks().iter().enumerate().for_each(|(i, block)| {
        let spans: Vec<Span> = block
            .content()
            .iter()
            .map(|line| {
                // bold style if the span is on the current block
                if model.block_cursor() == i {
                    Span::styled(line, Style::default().add_modifier(Modifier::BOLD))
                } else {
                    Span::from(line)
                }
            })
            .collect();
        let line = Line::from(spans);
        let text = Text::from(line);
        frame.render_widget(Paragraph::new(text).block(Block::bordered()), frame.area());
    })
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
