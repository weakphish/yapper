use model::{Model, RunningState};
use ratatui::{
    backend::{Backend, CrosstermBackend},
    crossterm::{
        terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
        ExecutableCommand,
    },
    Terminal,
};
use std::cmp::PartialEq;
use std::{io::stdout, panic};
use tea::handle_event;

mod model;
mod tea;

pub fn init_terminal() -> color_eyre::Result<Terminal<impl Backend>> {
    enable_raw_mode()?;
    stdout().execute(EnterAlternateScreen)?;
    let terminal = Terminal::new(CrosstermBackend::new(stdout()))?;
    Ok(terminal)
}

pub fn restore_terminal() -> color_eyre::Result<()> {
    stdout().execute(LeaveAlternateScreen)?;
    disable_raw_mode()?;
    Ok(())
}

pub fn install_panic_hook() {
    let original_hook = panic::take_hook();
    panic::set_hook(Box::new(move |panic_info| {
        stdout().execute(LeaveAlternateScreen).unwrap();
        disable_raw_mode().unwrap();
        original_hook(panic_info);
    }));
}

fn main() -> color_eyre::Result<()> {
    install_panic_hook();
    let mut terminal = init_terminal()?;
    let mut model = Model::default();

    while *model.running_state() != RunningState::Done {
        // Render the current view
        terminal.draw(|f| tea::view(&model, f))?;

        // Handle events and map to a Message
        if let Some(msg) = handle_event(&model)? {
            // Process messages until there are no more
            let mut current_model = model;
            let mut current_msg = Some(msg);

            while let Some(msg) = current_msg {
                let update = tea::update(&current_model, msg);
                (current_model, current_msg) = update.into_parts();
            }

            model = current_model;
        }
    }

    restore_terminal()?;
    Ok(())
}
