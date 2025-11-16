use std::env;
use std::io::{self, BufRead, Write};
use std::path::PathBuf;

use anyhow::{Result, anyhow};
use chrono::NaiveDate;
use note_core::IndexStore;
use note_core::{
    DateRange, Domain, FileSystemVault, InMemoryIndexStore, MarkdownParser, NoteId, TaskFilter,
    TaskId, TaskStatus, VaultIndexManager,
};
use serde::de::DeserializeOwned;
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};

fn main() -> Result<()> {
    let vault_root = env::var("NOTE_VAULT_PATH").unwrap_or_else(|_| ".".to_string());
    let vault = FileSystemVault::new(PathBuf::from(vault_root));
    let index = InMemoryIndexStore::new();
    let parser = Box::new(MarkdownParser::new());
    let index_mgr = VaultIndexManager::new(vault, index, parser);
    let mut domain = Domain::new(index_mgr);
    let _ = domain.reindex_all();

    let stdin = io::stdin();
    let mut stdout = io::stdout();

    for line in stdin.lock().lines() {
        let line = line?;
        if line.trim().is_empty() {
            continue;
        }
        match serde_json::from_str::<RpcRequest>(&line) {
            Ok(request) => {
                if let Some(response) = handle_request(&mut domain, request) {
                    let serialized = serde_json::to_string(&response)?;
                    writeln!(stdout, "{}", serialized)?;
                    stdout.flush()?;
                }
            }
            Err(err) => {
                let response = RpcResponse::error(Value::Null, -32700, err.to_string());
                let serialized = serde_json::to_string(&response)?;
                writeln!(stdout, "{}", serialized)?;
                stdout.flush()?;
            }
        }
    }

    Ok(())
}

fn handle_request(
    domain: &mut Domain<FileSystemVault, InMemoryIndexStore>,
    request: RpcRequest,
) -> Option<RpcResponse> {
    let id = match request.id {
        Some(id) => id,
        None => {
            // Notification – process but return no response.
            let _ = dispatch(domain, request.method.as_str(), request.params);
            return None;
        }
    };

    match dispatch(domain, request.method.as_str(), request.params) {
        Ok(value) => Some(RpcResponse::result(id, value)),
        Err(err) => Some(RpcResponse::error(id, -32000, err.to_string())),
    }
}

fn dispatch(
    domain: &mut Domain<FileSystemVault, InMemoryIndexStore>,
    method: &str,
    params: Option<Value>,
) -> Result<Value> {
    match method {
        "core.reindex" => {
            domain.reindex_all()?;
            Ok(json!({ "status": "ok" }))
        }
        "core.list_tasks" => {
            let params: ListTasksParams = parse_params(params)?;
            let mut filter = TaskFilter::default();
            filter.status = params.status;
            filter.tags = params.tags.unwrap_or_default();
            filter.text_search = params.text_search;
            filter.touched_since = params
                .touched_since
                .as_deref()
                .and_then(|date| NaiveDate::parse_from_str(date, "%Y-%m-%d").ok());
            let tasks = domain.list_tasks(&filter);
            Ok(serde_json::to_value(tasks)?)
        }
        "core.task_detail" => {
            let params: TaskDetailParams = parse_params(params)?;
            let id = TaskId(params.task_id);
            if let Some((task, mentions)) = domain.task_detail(&id) {
                let log_entries = domain.index_mgr.index.get_log_entries_for_task(&id);
                Ok(json!({
                    "task": task,
                    "mentions": mentions,
                    "log_entries": log_entries,
                }))
            } else {
                Err(anyhow!("task not found"))
            }
        }
        "core.items_for_tag" => {
            let params: TagParams = parse_params(params)?;
            let result = domain.items_for_tag(&params.tag);
            Ok(serde_json::to_value(result)?)
        }
        "core.notes_in_range" => {
            let params: RangeParams = parse_params(params)?;
            let range = parse_range(params.start, params.end)?;
            let notes = domain.notes_in_range(&range);
            Ok(serde_json::to_value(notes)?)
        }
        "core.weekly_summary" => {
            let params: RangeParams = parse_params(params)?;
            let range = parse_range(params.start, params.end)?;
            let summary = domain.weekly_summary(&range);
            Ok(serde_json::to_value(summary)?)
        }
        "core.open_daily" => {
            let params: OpenDailyParams = parse_params(params)?;
            let date = NaiveDate::parse_from_str(&params.date, "%Y-%m-%d")?;
            let note = domain.open_daily(date)?;
            Ok(serde_json::to_value(note)?)
        }
        "core.read_note" => {
            let params: NoteParams = parse_params(params)?;
            let id = NoteId(params.note_id);
            if let Some(note) = domain.read_note(&id) {
                Ok(serde_json::to_value(note)?)
            } else {
                Err(anyhow!("note not found"))
            }
        }
        "core.write_note" => {
            let params: WriteNoteParams = parse_params(params)?;
            let id = NoteId(params.note_id);
            let note = domain.write_note(&id, &params.content)?;
            Ok(serde_json::to_value(note)?)
        }
        _ => Err(anyhow!("unknown method {}", method)),
    }
}

fn parse_params<T: DeserializeOwned>(params: Option<Value>) -> Result<T> {
    let value = params.unwrap_or_else(|| json!({}));
    Ok(serde_json::from_value(value)?)
}

fn parse_range(start: String, end: String) -> Result<DateRange> {
    let start = NaiveDate::parse_from_str(&start, "%Y-%m-%d")?;
    let end = NaiveDate::parse_from_str(&end, "%Y-%m-%d")?;
    Ok(DateRange { start, end })
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
    error: Option<RpcError>,
}

#[derive(Serialize)]
struct RpcError {
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

    fn error(id: Value, code: i32, message: String) -> Self {
        RpcResponse {
            jsonrpc: "2.0",
            id,
            result: None,
            error: Some(RpcError { code, message }),
        }
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

#[derive(Deserialize)]
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
