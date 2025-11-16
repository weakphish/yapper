use std::io::Write;

use anyhow::Result;
use env_logger::Builder;
use log::Level;

/// Initialize a simple stderr logger backed by env_logger.
pub(crate) fn init_logger(level: Level) -> Result<()> {
    Builder::new()
        .filter_level(level.to_level_filter())
        .format(|buf, record| {
            writeln!(buf, "[{}][{}] {}", buf.timestamp(), record.level(), record.args())
        })
        .try_init()?;
    Ok(())
}
