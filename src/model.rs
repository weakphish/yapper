#[derive(Debug, Default, PartialEq, Eq, Clone)]
pub struct Model {
    running_state: RunningState,
    blocks: Vec<Block>,
}

impl Model {
    pub fn running_state(&self) -> &RunningState {
        &self.running_state
    }

    pub fn blocks(&self) -> &Vec<Block> {
        &self.blocks
    }
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
struct Block {
    tags: Vec<String>,
    content: Vec<String>,
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub enum RunningState {
    #[default]
    Running,
    Done,
}

#[derive(Clone)]
pub enum Message {
    Quit,
    MoveDown,
    MoveUp,
    AddBlock,
}

#[derive(Clone)]
pub struct Update {
    model: Model,
    message: Option<Message>,
}

impl Update {
    pub fn new(model: Model, message: Option<Message>) -> Self {
        Self { model, message }
    }

    pub fn into_parts(self) -> (Model, Option<Message>) {
        (self.model, self.message)
    }

    // Keep existing reference methods if needed for backward compatibility
    pub fn model(&self) -> &Model {
        &self.model
    }

    pub fn message(&self) -> &Option<Message> {
        &self.message
    }
}
