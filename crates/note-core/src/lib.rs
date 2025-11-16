pub mod domain;
pub mod index;
pub mod model;
pub mod parser;
pub mod vault;

pub use domain::Domain;
pub use domain::VaultIndexManager;
pub use index::{InMemoryIndexStore, IndexStore};
pub use model::*;
pub use parser::{NoteParser, RegexMarkdownParser};
pub use vault::{FileSystemVault, Vault};
