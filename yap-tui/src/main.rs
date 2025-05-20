use std::io;

use model::{Model, update, view};
use state::ViewState;
use util::{handle_event, handle_key};
use yap_core::block_tree::BlockTree;

mod message;
mod model;
mod state;
mod util;

/// TEA Architecture reference: https://ratatui.rs/concepts/application-patterns/the-elm-architecture/
/// Runs the application's main loop until the user quits
fn main() -> color_eyre::Result<()>{
    // let mut terminal = ratatui::init();
 
    let block_tree = BlockTree::new("foobar");
    dbg!(&block_tree);

    let mut model = Model::new(block_tree);
    dbg!(&model);
    // while model.view_state != ViewState::Exit {
    //     // Render current view
    //     terminal.draw(|frame| view(&model, frame))?;
    //
    //     // Handle events and map to a Message
    //     // TODO: can this be modified to be non-mutable?
    //     let mut current_message = handle_event(&model)?;
    //
    //     // Process updates as long as they return a non-None message
    //     while current_message.is_some() {
    //         current_message = update(&mut model, current_message.unwrap());
    //     }
    // }
    // ratatui::restore();
    Ok(())
}
