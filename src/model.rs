#[derive(Debug, Default, PartialEq, Eq, Clone)]
pub struct Model {
    running_state: ApplicationView,
    blocks: Vec<Block>,
    block_cursor: usize,
}

impl Model {
    pub fn running_state(&self) -> &ApplicationView {
        &self.running_state
    }

    pub fn blocks(&self) -> &Vec<Block> {
        &self.blocks
    }

    pub fn set_block_cursor(&mut self, block_cursor: usize) {
        self.block_cursor = block_cursor;
    }

    pub fn block_cursor(&self) -> usize {
        self.block_cursor
    }
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub enum ApplicationView {
    #[default]
    Main,
    Editing,
    Done,
}

#[derive(Clone)]
pub enum Message {
    Quit,
    MoveDown,
    MoveUp,
    AddBlock,
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Block {
    tags: Vec<String>,
    content: Vec<String>,
    editing_block: bool,
}

impl Block {
    pub fn tags(&self) -> &Vec<String> {
        &self.tags
    }

    pub fn content(&self) -> &Vec<String> {
        &self.content
    }

    pub fn set_tags(&mut self, tags: Vec<String>) {
        self.tags = tags;
    }

    pub fn set_content(&mut self, content: Vec<String>) {
        self.content = content;
    }

    pub fn editing_block(&self) -> bool {
        self.editing_block
    }

    pub fn set_editing_block(&mut self, editing_block: bool) {
        self.editing_block = editing_block;
    }
}

/// Represents an update returned from the `update` function, derived from the TEA architecture.
/// Wraps an optional `Message` and the model.
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
