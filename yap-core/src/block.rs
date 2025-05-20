use ropey::Rope;
use uuid::Uuid;

enum BlockType {
    Note,
    Task
}

#[derive(Debug, Clone, Default,PartialEq, Eq)]
pub struct Block {
    id: Uuid,
    dependent_ids: Vec<Uuid>,
    dependency_ids: Vec<Uuid>,
    content: Rope,
}

impl Block {
    pub fn new(content: Rope, node_id: Uuid) -> Self {
        Block {
            id: Uuid::now_v7(),
            dependent_ids: vec!(),
            dependency_ids: vec!(),
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
