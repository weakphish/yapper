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
use crate::logging::{LogLevel, Logger};
use crate::server::{AppDomain, run_server};

fn main() -> Result<()> {
    let config = DaemonConfig::from_env_and_args(env::args())?;
    let logger = Logger::new(config.log_level);

    logger.log(
        LogLevel::Info,
        format_args!(
            "starting note-daemon (vault: {})",
            config.vault_path.display()
        ),
    );

    let vault = FileSystemVault::new(config.vault_path.clone());
    let index = InMemoryIndexStore::new();
    let parser = Box::new(RegexMarkdownParser::new());
    let index_mgr = VaultIndexManager::new(vault, index, parser);
    let mut domain: AppDomain = Domain::new(index_mgr);

    if let Err(err) = domain.reindex_all() {
        logger.log(
            LogLevel::Error,
            format_args!("initial vault reindex failed: {}", err),
        );
    } else {
        logger.log(LogLevel::Info, format_args!("vault reindex completed"));
    }

    run_server(&mut domain, &logger)
}
