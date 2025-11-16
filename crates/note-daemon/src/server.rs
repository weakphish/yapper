use std::io::{self, BufRead, Write};

use anyhow::Result;
use note_core::{Domain, FileSystemVault, InMemoryIndexStore, IndexStore, NoteId, TaskFilter, TaskId};
use serde_json::{Value, json};

use crate::logging::{LogLevel, Logger};
use crate::rpc::{
    ListTasksParams, NoteParams, OpenDailyParams, RangeParams, RpcError, RpcRequest, RpcResponse,
    RpcResult, TagParams, TaskDetailParams, WriteNoteParams, parse_date, parse_params, parse_range,
    to_json,
};

pub(crate) type AppDomain = Domain<FileSystemVault, InMemoryIndexStore>;

pub(crate) fn run_server(domain: &mut AppDomain, logger: &Logger) -> Result<()> {
    let stdin = io::stdin();
    let mut stdout = io::stdout();

    for raw_line in stdin.lock().lines() {
        let raw_line = match raw_line {
            Ok(line) => line,
            Err(err) => {
                logger.log(LogLevel::Error, format_args!("stdin read error: {}", err));
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
                    logger.log(
                        LogLevel::Debug,
                        format_args!("responded with {}", serialized),
                    );
                }
            }
            Err(err) => {
                logger.log(LogLevel::Warn, format_args!("malformed JSON: {}", err));
                let response =
                    RpcResponse::error(Value::Null, RpcError::parse_error(err.to_string()));
                let serialized = serde_json::to_string(&response)?;
                writeln!(stdout, "{}", serialized)?;
                stdout.flush()?;
            }
        }
    }

    logger.log(LogLevel::Info, format_args!("stdin closed, shutting down"));
    Ok(())
}

fn handle_request(
    domain: &mut AppDomain,
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
                logger.log(
                    LogLevel::Warn,
                    format_args!("notification for method '{}' failed: {}", method, err),
                );
                None
            }
        }
    }
}

fn dispatch(domain: &mut AppDomain, method: &str, params: Option<Value>) -> RpcResult<Value> {
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
