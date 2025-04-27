use std::io;

use model::Model;
use ratatui::{style::Stylize, widgets::Widget};

mod model;
mod page;
mod state;

fn main() -> io::Result<()> {
    let mut terminal = ratatui::init();
    let app_result = Model::default().run(&mut terminal);
    ratatui::restore();
    app_result
}
