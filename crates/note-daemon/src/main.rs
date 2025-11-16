use std::env;
use std::fmt;
use std::io::{self, BufRead, Write};
use std::path::PathBuf;

use anyhow::{Result, anyhow};
use chrono::{Local, NaiveDate};
use note_core::IndexStore;
use note_core::{
    DateRange, Domain, FileSystemVault, InMemoryIndexStore, NoteId, RegexMarkdownParser,
    TaskFilter, TaskId, TaskStatus, VaultIndexManager,
};
use serde::de::DeserializeOwned;
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};

/// Severity levels for the daemon's stderr logging.
#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd)]
enum LogLevel {
    Error,
    Warn,
    Info,
    Debug,
}

impl LogLevel {
    fn parse(input: &str) -> Option<Self> {
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
struct Logger {
    level: LogLevel,
}

impl Logger {
    fn new(level: LogLevel) -> Self {
        Logger { level }
    }

    fn enabled(&self, level: LogLevel) -> bool {
        level <= self.level
    }

    fn log(&self, level: LogLevel, args: fmt::Arguments<'_>) {
        if !self.enabled(level) {
            return;
        }
        let timestamp = Local::now().format("%Y-%m-%dT%H:%M:%S");
        eprintln!("[{}][{}] {}", timestamp, level, args);
    }
}

macro_rules! log_line {
    ($logger:expr, $level:ident, $($arg:tt)*) => {{
        $logger.log(LogLevel::$level, format_args!($($arg)*));
    }};
}

/// Runtime configuration derived from env vars and CLI flags.
#[derive(Clone, Debug)]
struct DaemonConfig {
    vault_path: PathBuf,
    log_level: LogLevel,
}

impl DaemonConfig {
    fn from_env_and_args(args: env::Args) -> Result<Self> {
        Self::from_iterator(
            env::var("NOTE_VAULT_PATH").ok(),
            env::var("NOTE_DAEMON_LOG").ok(),
            args,
        )
    }

    fn from_iterator<I>(
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

    fn usage() -> &'static str {
        "note-daemon [--vault PATH] [--log-level error|warn|info|debug]"
    }
}

fn main() -> Result<()> {
    let config = DaemonConfig::from_env_and_args(env::args())?;
    let logger = Logger::new(config.log_level);

    log_line!(
        logger,
        Info,
        "starting note-daemon (vault: {})",
        config.vault_path.display()
    );

    let vault = FileSystemVault::new(config.vault_path.clone());
    let index = InMemoryIndexStore::new();
    let parser = Box::new(RegexMarkdownParser::new());
    let index_mgr = VaultIndexManager::new(vault, index, parser);
    let mut domain = Domain::new(index_mgr);

    if let Err(err) = domain.reindex_all() {
        log_line!(logger, Error, "initial vault reindex failed: {}", err);
    } else {
        log_line!(logger, Info, "vault reindex completed");
    }

    run_server(&mut domain, &logger)
}

fn run_server(
    domain: &mut Domain<FileSystemVault, InMemoryIndexStore>,
    logger: &Logger,
) -> Result<()> {
    let stdin = io::stdin();
    let mut stdout = io::stdout();

    for raw_line in stdin.lock().lines() {
        let raw_line = match raw_line {
            Ok(line) => line,
            Err(err) => {
                log_line!(logger, Error, "stdin read error: {}", err);
                break;
            }
        };

        if raw_line.trim().is_empty() {
            continue;
        }

        match serde_json::from_str::<RpcRequest>(&raw_line) {
            Ok(request) => {
                if let Some(response) = handle_request(domain, request, logger) {
                    let serialized = serde_json::to_string(&response)?;
                    writeln!(stdout, "{}", serialized)?;
                    stdout.flush()?;
                    log_line!(logger, Debug, "responded with {}", serialized);
                }
            }
            Err(err) => {
                log_line!(logger, Warn, "malformed JSON: {}", err);
                let response =
                    RpcResponse::error(Value::Null, RpcError::parse_error(err.to_string()));
                let serialized = serde_json::to_string(&response)?;
                writeln!(stdout, "{}", serialized)?;
                stdout.flush()?;
            }
        }
    }

    log_line!(logger, Info, "stdin closed, shutting down");
    Ok(())
}

fn handle_request(
    domain: &mut Domain<FileSystemVault, InMemoryIndexStore>,
    request: RpcRequest,
    logger: &Logger,
) -> Option<RpcResponse> {
    let RpcRequest {
        jsonrpc,
        id,
        method,
        params,
    } = request;

    if jsonrpc.as_deref() != Some("2.0") {
        let response_id = id.clone().unwrap_or(Value::Null);
        return Some(RpcResponse::error(
            response_id,
            RpcError::invalid_request("jsonrpc field must be \"2.0\""),
        ));
    }

    match dispatch(domain, &method, params) {
        Ok(value) => id.map(|id| RpcResponse::result(id, value)),
        Err(err) => {
            if let Some(id) = id {
                Some(RpcResponse::error(id, err))
            } else {
                log_line!(
                    logger,
                    Warn,
                    "notification for method '{}' failed: {}",
                    method,
                    err
                );
                None
            }
        }
    }
}

fn dispatch(
    domain: &mut Domain<FileSystemVault, InMemoryIndexStore>,
    method: &str,
    params: Option<Value>,
) -> RpcResult<Value> {
    match method {
        "core.reindex" => {
            domain.reindex_all()?;
            to_json(json!({ "status": "ok" }))
        }
        "core.list_tasks" => {
            let params: ListTasksParams = parse_params(params)?;
            let mut filter = TaskFilter::default();
            filter.status = params.status;
            filter.tags = params.tags.unwrap_or_default();
            filter.text_search = params.text_search;
            filter.touched_since = match params.touched_since.as_deref() {
                Some(date) => Some(parse_date(date)?),
                None => None,
            };
            let tasks = domain.list_tasks(&filter);
            to_json(tasks)
        }
        "core.task_detail" => {
            let params: TaskDetailParams = parse_params(params)?;
            let id = TaskId(params.task_id);
            if let Some((task, mentions)) = domain.task_detail(&id) {
                let log_entries = domain.index_mgr.index.get_log_entries_for_task(&id);
                to_json(json!({
                    "task": task,
                    "mentions": mentions,
                    "log_entries": log_entries,
                }))
            } else {
                Err(RpcError::invalid_request("task not found"))
            }
        }
        "core.items_for_tag" => {
            let params: TagParams = parse_params(params)?;
            let result = domain.items_for_tag(&params.tag);
            to_json(result)
        }
        "core.notes_in_range" => {
            let params: RangeParams = parse_params(params)?;
            let range = parse_range(params.start, params.end)?;
            let notes = domain.notes_in_range(&range);
            to_json(notes)
        }
        "core.weekly_summary" => {
            let params: RangeParams = parse_params(params)?;
            let range = parse_range(params.start, params.end)?;
            let summary = domain.weekly_summary(&range);
            to_json(summary)
        }
        "core.open_daily" => {
            let params: OpenDailyParams = parse_params(params)?;
            let date = parse_date(&params.date)?;
            let note = domain.open_daily(date)?;
            to_json(note)
        }
        "core.read_note" => {
            let params: NoteParams = parse_params(params)?;
            let id = NoteId(params.note_id);
            if let Some(note) = domain.read_note(&id) {
                to_json(note)
            } else {
                Err(RpcError::invalid_request("note not found"))
            }
        }
        "core.write_note" => {
            let params: WriteNoteParams = parse_params(params)?;
            let id = NoteId(params.note_id);
            let note = domain.write_note(&id, &params.content)?;
            to_json(note)
        }
        _ => Err(RpcError::method_not_found(method)),
    }
}

type RpcResult<T> = Result<T, RpcError>;

fn to_json<T: Serialize>(value: T) -> RpcResult<Value> {
    serde_json::to_value(value).map_err(|err| RpcError::internal(err.to_string()))
}

fn parse_params<T: DeserializeOwned>(params: Option<Value>) -> RpcResult<T> {
    let value = params.unwrap_or_else(|| json!({}));
    serde_json::from_value(value).map_err(|err| RpcError::invalid_params(err.to_string()))
}

fn parse_range(start: String, end: String) -> RpcResult<DateRange> {
    let start = parse_date(&start)?;
    let end = parse_date(&end)?;
    Ok(DateRange { start, end })
}

fn parse_date(input: &str) -> RpcResult<NaiveDate> {
    NaiveDate::parse_from_str(input, "%Y-%m-%d")
        .map_err(|_| RpcError::invalid_params(format!("invalid date '{}'", input)))
}

#[derive(Deserialize)]
struct RpcRequest {
    pub jsonrpc: Option<String>,
    pub id: Option<Value>,
    pub method: String,
    pub params: Option<Value>,
}

#[derive(Serialize)]
struct RpcResponse {
    jsonrpc: &'static str,
    id: Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    result: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<RpcErrorBody>,
}

#[derive(Serialize)]
struct RpcErrorBody {
    code: i32,
    message: String,
}

impl RpcResponse {
    fn result(id: Value, result: Value) -> Self {
        RpcResponse {
            jsonrpc: "2.0",
            id,
            result: Some(result),
            error: None,
        }
    }

    fn error(id: Value, error: RpcError) -> Self {
        RpcResponse {
            jsonrpc: "2.0",
            id,
            result: None,
            error: Some(RpcErrorBody {
                code: error.code.as_i32(),
                message: error.message,
            }),
        }
    }
}

#[derive(Clone, Copy, Debug)]
enum RpcErrorCode {
    ParseError = -32700,
    InvalidRequest = -32600,
    MethodNotFound = -32601,
    InvalidParams = -32602,
    InternalError = -32603,
    ServerError = -32000,
}

impl RpcErrorCode {
    fn as_i32(self) -> i32 {
        self as i32
    }
}

#[derive(Debug)]
struct RpcError {
    code: RpcErrorCode,
    message: String,
}

impl RpcError {
    fn new(code: RpcErrorCode, message: impl Into<String>) -> Self {
        RpcError {
            code,
            message: message.into(),
        }
    }

    fn parse_error(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::ParseError, message)
    }

    fn invalid_request(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InvalidRequest, message)
    }

    fn invalid_params(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InvalidParams, message)
    }

    fn method_not_found(method: &str) -> Self {
        RpcError::new(
            RpcErrorCode::MethodNotFound,
            format!("unknown method '{}'", method),
        )
    }

    fn internal(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InternalError, message)
    }
}

impl fmt::Display for RpcError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} (code {})", self.message, self.code.as_i32())
    }
}

impl From<anyhow::Error> for RpcError {
    fn from(err: anyhow::Error) -> Self {
        RpcError::new(RpcErrorCode::ServerError, err.to_string())
    }
}

#[derive(Default, Deserialize)]
#[serde(default)]
struct ListTasksParams {
    status: Option<TaskStatus>,
    tags: Option<Vec<String>>,
    text_search: Option<String>,
    touched_since: Option<String>,
}

#[derive(Deserialize)]
struct TaskDetailParams {
    task_id: String,
}

#[derive(Deserialize)]
struct TagParams {
    tag: String,
}

#[derive(Debug, Deserialize)]
struct RangeParams {
    start: String,
    end: String,
}

#[derive(Deserialize)]
struct OpenDailyParams {
    date: String,
}

#[derive(Deserialize)]
struct NoteParams {
    note_id: String,
}

#[derive(Deserialize)]
struct WriteNoteParams {
    note_id: String,
    content: String,
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;
    use std::path::PathBuf;

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

    #[test]
    fn config_defaults_when_cli_missing() {
        let args = vec!["note-daemon".to_string()].into_iter();
        let config = DaemonConfig::from_iterator(None, None, args).expect("defaults should parse");

        assert_eq!(config.vault_path, PathBuf::from("."));
        assert_eq!(config.log_level, LogLevel::Info);
    }

    #[test]
    fn parse_params_reports_invalid_input() {
        let err = parse_params::<RangeParams>(Some(json!({ "start": 123 }))).unwrap_err();
        assert_eq!(err.code.as_i32(), RpcErrorCode::InvalidParams.as_i32());
    }
}
