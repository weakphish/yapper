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
    /// Parses a textual log level value into the enum variant.
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
    /// Converts the custom log level into the `log` crate filter.
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
    /// Formats the level in uppercase for log line prefixes.
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

/// Simple stderr logger that honors a max level filter.
#[derive(Debug)]
struct StderrLogger {
    level: LevelFilter,
}

impl StderrLogger {
    /// Constructs the logger with the provided filter threshold.
    fn new(level: LevelFilter) -> Self {
        Self { level }
    }
}

impl Log for StderrLogger {
    /// Indicates whether a record should be emitted.
    fn enabled(&self, metadata: &Metadata<'_>) -> bool {
        metadata.level().to_level_filter() <= self.level
    }

    /// Writes formatted log records to stderr with a timestamp prefix.
    fn log(&self, record: &Record<'_>) {
        if !self.enabled(record.metadata()) {
            return;
        }

        let timestamp = Local::now().format("%Y-%m-%dT%H:%M:%S");
        eprintln!("[{}][{}] {}", timestamp, record.level(), record.args());
    }

    /// Flush is a no-op because writes go directly to stderr.
    fn flush(&self) {}
}
