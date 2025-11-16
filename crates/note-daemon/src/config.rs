use std::env;
use std::path::PathBuf;

use anyhow::{Result, anyhow};

use crate::logging::LogLevel;

/// Runtime configuration derived from env vars and CLI flags.
#[derive(Clone, Debug)]
pub(crate) struct DaemonConfig {
    pub(crate) vault_path: PathBuf,
    pub(crate) log_level: LogLevel,
}

impl DaemonConfig {
    /// Builds config from the process environment plus CLI arguments.
    pub(crate) fn from_env_and_args(args: env::Args) -> Result<Self> {
        Self::from_iterator(
            env::var("NOTE_VAULT_PATH").ok(),
            env::var("NOTE_DAEMON_LOG").ok(),
            args,
        )
    }

    /// Parses configuration sources in priority order (`args` > env vars).
    pub(crate) fn from_iterator<I>(
        vault_env: Option<String>,
        log_env: Option<String>,
        mut args: I,
    ) -> Result<Self>
    where
        I: Iterator<Item = String>,
    {
        let mut vault_path = vault_env.unwrap_or_else(|| ".".to_string());
        let mut log_level = log_env
            .as_deref()
            .and_then(LogLevel::parse)
            .unwrap_or(LogLevel::Info);

        // Drop the program name if present.
        let _ = args.next();

        while let Some(arg) = args.next() {
            match arg.as_str() {
                "--vault" | "-v" => {
                    let path = args
                        .next()
                        .ok_or_else(|| anyhow!("--vault expects a following path"))?;
                    vault_path = path;
                }
                "--log-level" | "-l" => {
                    let value = args
                        .next()
                        .ok_or_else(|| anyhow!("--log-level expects a value"))?;
                    log_level = LogLevel::parse(&value)
                        .ok_or_else(|| anyhow!("invalid log level '{}'", value))?;
                }
                other => {
                    return Err(anyhow!(
                        "unrecognized argument '{}'. Usage: {}",
                        other,
                        DaemonConfig::usage()
                    ));
                }
            }
        }

        Ok(Self {
            vault_path: PathBuf::from(vault_path),
            log_level,
        })
    }

    /// Returns the CLI usage string for help/error messages.
    pub(crate) fn usage() -> &'static str {
        "note-daemon [--vault PATH] [--log-level error|warn|info|debug]"
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Ensures CLI flags override values sourced from env vars.
    #[test]
    fn config_prefers_cli_values() {
        let args = vec![
            "note-daemon".to_string(),
            "--vault".to_string(),
            "/cli/vault".to_string(),
            "--log-level".to_string(),
            "warn".to_string(),
        ]
        .into_iter();

        let config =
            DaemonConfig::from_iterator(Some("/env/vault".into()), Some("debug".into()), args)
                .expect("config should parse");

        assert_eq!(config.vault_path, PathBuf::from("/cli/vault"));
        assert_eq!(config.log_level, LogLevel::Warn);
    }

    /// Verifies defaults apply when neither env vars nor CLI flags override them.
    #[test]
    fn config_defaults_when_cli_missing() {
        let args = vec!["note-daemon".to_string()].into_iter();
        let config = DaemonConfig::from_iterator(None, None, args).expect("defaults should parse");

        assert_eq!(config.vault_path, PathBuf::from("."));
        assert_eq!(config.log_level, LogLevel::Info);
    }
}
