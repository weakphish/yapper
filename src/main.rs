 use ratatui::{
        backend::{Backend, CrosstermBackend},
        crossterm::{
            terminal::{
                disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen,
            },
            ExecutableCommand,
        },
        Terminal,
    };
    use std::{io::stdout, panic};

#[derive(Debug, Default, PartialEq, Eq)]
struct Model {
    runnning_state: RunningState,
    blocks: Vec<Block>,
}

#[derive(Debug, Default, PartialEq, Eq)]
struct Block {
    tags: Vec<String>,
    content: Vec<String>,
}

#[derive(Debug, Default, PartialEq, Eq)]
enum RunningState {
    #[default]
    Running,
    Done,
}

enum Message {
    // TODO: define messages
}

/// Hence, the update function should avoid direct mutation of the model.
/// "A key feature of TEA is immutability.
/// Instead, it should produce a new instance of the model reflecting the desired changes."
fn update(model: &Model, msg: Message) -> Model {
    match msg {
        // Match each possible message and decide how the model should change
        // Return a new model reflecting those changes
    }
}

fn view(model: &Model) {
    //... use `ratatui` functions to draw your UI based on the model's state
}

fn main() {
    todo!()
}
