use std::fmt;

use anyhow::Result;
use chrono::Local;
use log::{LevelFilter, Log, Metadata, Record};

/// Severity levels for the daemon's stderr logging.
#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd)]
pub(crate) enum LogLevel {
    Error,
    Warn,
    Info,
    Debug,
}

impl LogLevel {
    pub(crate) fn parse(input: &str) -> Option<Self> {
        match input.to_ascii_lowercase().as_str() {
            "error" => Some(LogLevel::Error),
            "warn" | "warning" => Some(LogLevel::Warn),
            "info" => Some(LogLevel::Info),
            "debug" => Some(LogLevel::Debug),
            _ => None,
        }
    }
}

impl From<LogLevel> for LevelFilter {
    fn from(value: LogLevel) -> Self {
        match value {
            LogLevel::Error => LevelFilter::Error,
            LogLevel::Warn => LevelFilter::Warn,
            LogLevel::Info => LevelFilter::Info,
            LogLevel::Debug => LevelFilter::Debug,
        }
    }
}

impl fmt::Display for LogLevel {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            LogLevel::Error => f.write_str("ERROR"),
            LogLevel::Warn => f.write_str("WARN"),
            LogLevel::Info => f.write_str("INFO"),
            LogLevel::Debug => f.write_str("DEBUG"),
        }
    }
}

/// Initialize the global logger that bridges to the `log` crate.
pub(crate) fn init_logger(level: LogLevel) -> Result<()> {
    let level_filter = LevelFilter::from(level);
    log::set_boxed_logger(Box::new(StderrLogger::new(level_filter)))?;
    log::set_max_level(level_filter);
    Ok(())
}

#[derive(Debug)]
struct StderrLogger {
    level: LevelFilter,
}

impl StderrLogger {
    fn new(level: LevelFilter) -> Self {
        Self { level }
    }
}

impl Log for StderrLogger {
    fn enabled(&self, metadata: &Metadata<'_>) -> bool {
        metadata.level().to_level_filter() <= self.level
    }

    fn log(&self, record: &Record<'_>) {
        if !self.enabled(record.metadata()) {
            return;
        }

        let timestamp = Local::now().format("%Y-%m-%dT%H:%M:%S");
        eprintln!("[{}][{}] {}", timestamp, record.level(), record.args());
    }

    fn flush(&self) {}
}
