mod config;
mod logging;
mod rpc;
mod server;

use std::env;

use anyhow::Result;
use note_core::{
    Domain, FileSystemVault, InMemoryIndexStore, RegexMarkdownParser, VaultIndexManager,
};

use crate::config::DaemonConfig;
use crate::logging::init_logger;
use crate::server::{AppDomain, run_server};

fn main() -> Result<()> {
    let config = DaemonConfig::from_env_and_args(env::args())?;
    init_logger(config.log_level)?;

    log::info!(
        "starting note-daemon (vault: {})",
        config.vault_path.display()
    );

    let vault = FileSystemVault::new(config.vault_path.clone());
    let index = InMemoryIndexStore::new();
    let parser = Box::new(RegexMarkdownParser::new());
    let index_mgr = VaultIndexManager::new(vault, index, parser);
    let mut domain: AppDomain = Domain::new(index_mgr);

    if let Err(err) = domain.reindex_all() {
        log::error!("initial vault reindex failed: {}", err);
    } else {
        log::info!("vault reindex completed");
    }

    run_server(&mut domain)
}
