use std::fmt;

use chrono::NaiveDate;
use note_core::{DateRange, TaskStatus};
use serde::de::DeserializeOwned;
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};

pub(crate) type RpcResult<T> = Result<T, RpcError>;

pub(crate) fn to_json<T: Serialize>(value: T) -> RpcResult<Value> {
    serde_json::to_value(value).map_err(|err| RpcError::internal(err.to_string()))
}

pub(crate) fn parse_params<T: DeserializeOwned>(params: Option<Value>) -> RpcResult<T> {
    let value = params.unwrap_or_else(|| json!({}));
    serde_json::from_value(value).map_err(|err| RpcError::invalid_params(err.to_string()))
}

pub(crate) fn parse_range(start: String, end: String) -> RpcResult<DateRange> {
    let start = parse_date(&start)?;
    let end = parse_date(&end)?;
    Ok(DateRange { start, end })
}

pub(crate) fn parse_date(input: &str) -> RpcResult<NaiveDate> {
    NaiveDate::parse_from_str(input, "%Y-%m-%d")
        .map_err(|_| RpcError::invalid_params(format!("invalid date '{}'", input)))
}

#[derive(Deserialize)]
pub(crate) struct RpcRequest {
    pub jsonrpc: Option<String>,
    pub id: Option<Value>,
    pub method: String,
    pub params: Option<Value>,
}

#[derive(Serialize)]
pub(crate) struct RpcResponse {
    jsonrpc: &'static str,
    id: Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    result: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<RpcErrorBody>,
}

#[derive(Serialize)]
pub(crate) struct RpcErrorBody {
    code: i32,
    message: String,
}

impl RpcResponse {
    pub(crate) fn result(id: Value, result: Value) -> Self {
        RpcResponse {
            jsonrpc: "2.0",
            id,
            result: Some(result),
            error: None,
        }
    }

    pub(crate) fn error(id: Value, error: RpcError) -> Self {
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
pub(crate) struct RpcError {
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

    pub(crate) fn parse_error(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::ParseError, message)
    }

    pub(crate) fn invalid_request(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InvalidRequest, message)
    }

    pub(crate) fn invalid_params(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InvalidParams, message)
    }

    pub(crate) fn method_not_found(method: &str) -> Self {
        RpcError::new(
            RpcErrorCode::MethodNotFound,
            format!("unknown method '{}'", method),
        )
    }

    pub(crate) fn internal(message: impl Into<String>) -> Self {
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
pub(crate) struct ListTasksParams {
    pub status: Option<TaskStatus>,
    pub tags: Option<Vec<String>>,
    pub text_search: Option<String>,
    pub touched_since: Option<String>,
}

#[derive(Deserialize)]
pub(crate) struct TaskDetailParams {
    pub task_id: String,
}

#[derive(Deserialize)]
pub(crate) struct TagParams {
    pub tag: String,
}

#[derive(Debug, Deserialize)]
pub(crate) struct RangeParams {
    pub start: String,
    pub end: String,
}

#[derive(Deserialize)]
pub(crate) struct OpenDailyParams {
    pub date: String,
}

#[derive(Deserialize)]
pub(crate) struct NoteParams {
    pub note_id: String,
}

#[derive(Deserialize)]
pub(crate) struct WriteNoteParams {
    pub note_id: String,
    pub content: String,
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn parse_params_reports_invalid_input() {
        let err = parse_params::<RangeParams>(Some(json!({ "start": 123 }))).unwrap_err();
        assert_eq!(err.code.as_i32(), RpcErrorCode::InvalidParams.as_i32());
    }
}
