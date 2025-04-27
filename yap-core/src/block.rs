use ropey::Rope;
use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Block {
    id: Uuid,
    content: Rope,
}

impl Block {
    pub fn new(content: Rope, node_id: Uuid) -> Self {
        Block {
            id: Uuid::now_v7(),
            content,
        }
    }

    pub fn id(&self) -> Uuid {
        self.id
    }

    pub fn content(&self) -> &Rope {
        &self.content
    }
}
