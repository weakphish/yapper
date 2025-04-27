use ropey::Rope;
use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Block {
    id: Uuid,
    content: Rope,
}

impl Block {
    pub fn new(content: Rope) -> Self {
        Block {
            id: Uuid::now_v7(),
            content,
        }
    }
}
