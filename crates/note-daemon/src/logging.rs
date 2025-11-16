use std::io::Write;

use anyhow::Result;
use env_logger::Builder;
use log::LevelFilter;

/// Initialize a simple stderr logger backed by env_logger.
pub(crate) fn init_logger(level: LevelFilter) -> Result<()> {
    Builder::new()
        .filter_level(level)
        .format(|buf, record| {
            writeln!(buf, "[{}][{}] {}", buf.timestamp(), record.level(), record.args())
        })
        .try_init()?;
    Ok(())
}
