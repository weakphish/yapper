use std::fmt;

use chrono::Local;

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

    fn as_str(self) -> &'static str {
        match self {
            LogLevel::Error => "ERROR",
            LogLevel::Warn => "WARN",
            LogLevel::Info => "INFO",
            LogLevel::Debug => "DEBUG",
        }
    }
}

impl fmt::Display for LogLevel {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(self.as_str())
    }
}

/// Minimal stderr logger that avoids extra dependencies.
#[derive(Clone, Debug)]
pub(crate) struct Logger {
    level: LogLevel,
}

impl Logger {
    pub(crate) fn new(level: LogLevel) -> Self {
        Logger { level }
    }

    fn enabled(&self, level: LogLevel) -> bool {
        level <= self.level
    }

    pub(crate) fn log(&self, level: LogLevel, args: fmt::Arguments<'_>) {
        if !self.enabled(level) {
            return;
        }
        let timestamp = Local::now().format("%Y-%m-%dT%H:%M:%S");
        eprintln!("[{}][{}] {}", timestamp, level, args);
    }
}
