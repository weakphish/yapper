use std::io;

use crate::model::run;
use model::Model;
use ratatui::{style::Stylize, widgets::Widget};

mod message;
mod model;
mod page;
mod state;

fn main() -> io::Result<()> {
    let mut terminal = ratatui::init();
    let mut model = Model::default();
    let app_result = run(&mut model, &mut terminal);
    ratatui::restore();
    app_result
}
