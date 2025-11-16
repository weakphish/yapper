use std::fmt;

use chrono::NaiveDate;
use note_core::{DateRange, TaskStatus};
use serde::de::DeserializeOwned;
use serde::{Deserialize, Serialize};
use serde_json::{Value, json};

/// Convenience result alias used across the RPC surface area.
pub(crate) type RpcResult<T> = Result<T, RpcError>;

/// Serializes arbitrary types to `serde_json::Value`, mapping errors to RPC form.
pub(crate) fn to_json<T: Serialize>(value: T) -> RpcResult<Value> {
    serde_json::to_value(value).map_err(|err| RpcError::internal(err.to_string()))
}

/// Parses RPC params into the desired struct, defaulting to `{}` when absent.
pub(crate) fn parse_params<T: DeserializeOwned>(params: Option<Value>) -> RpcResult<T> {
    let value = params.unwrap_or_else(|| json!({}));
    serde_json::from_value(value).map_err(|err| RpcError::invalid_params(err.to_string()))
}

/// Helper that converts string dates into a `DateRange`.
pub(crate) fn parse_range(start: String, end: String) -> RpcResult<DateRange> {
    let start = parse_date(&start)?;
    let end = parse_date(&end)?;
    Ok(DateRange { start, end })
}

/// Parses the canonical `YYYY-MM-DD` date format.
pub(crate) fn parse_date(input: &str) -> RpcResult<NaiveDate> {
    NaiveDate::parse_from_str(input, "%Y-%m-%d")
        .map_err(|_| RpcError::invalid_params(format!("invalid date '{}'", input)))
}

/// Raw JSON-RPC request payload.
#[derive(Deserialize)]
pub(crate) struct RpcRequest {
    pub jsonrpc: Option<String>,
    pub id: Option<Value>,
    pub method: String,
    pub params: Option<Value>,
}

/// Successful or error response envelope according to JSON-RPC 2.0.
#[derive(Serialize)]
pub(crate) struct RpcResponse {
    jsonrpc: &'static str,
    id: Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    result: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<RpcErrorBody>,
}

/// Payload reported when an RPC call fails.
#[derive(Serialize)]
pub(crate) struct RpcErrorBody {
    code: i32,
    message: String,
}

impl RpcResponse {
    /// Creates a success response containing the provided result.
    pub(crate) fn result(id: Value, result: Value) -> Self {
        RpcResponse {
            jsonrpc: "2.0",
            id,
            result: Some(result),
            error: None,
        }
    }

    /// Creates an error response with the provided metadata.
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

/// Enumerates codes defined by the JSON-RPC spec (plus server errors).
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
    /// Returns the integer representation for a code.
    fn as_i32(self) -> i32 {
        self as i32
    }
}

/// Application error wrapper used throughout the RPC layer.
#[derive(Debug)]
pub(crate) struct RpcError {
    code: RpcErrorCode,
    message: String,
}

impl RpcError {
    /// Builds an error for the provided code/message pair.
    fn new(code: RpcErrorCode, message: impl Into<String>) -> Self {
        RpcError {
            code,
            message: message.into(),
        }
    }

    /// Creates a parse error response.
    pub(crate) fn parse_error(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::ParseError, message)
    }

    /// Creates an invalid request response.
    pub(crate) fn invalid_request(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InvalidRequest, message)
    }

    /// Creates an invalid params response.
    pub(crate) fn invalid_params(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InvalidParams, message)
    }

    /// Reports a method-not-found error.
    pub(crate) fn method_not_found(method: &str) -> Self {
        RpcError::new(
            RpcErrorCode::MethodNotFound,
            format!("unknown method '{}'", method),
        )
    }

    /// Creates an internal server error response.
    pub(crate) fn internal(message: impl Into<String>) -> Self {
        RpcError::new(RpcErrorCode::InternalError, message)
    }
}

impl fmt::Display for RpcError {
    /// Formats the error in a human-readable manner for log output.
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} (code {})", self.message, self.code.as_i32())
    }
}

impl From<anyhow::Error> for RpcError {
    /// Converts application errors into JSON-RPC server error responses.
    fn from(err: anyhow::Error) -> Self {
        RpcError::new(RpcErrorCode::ServerError, err.to_string())
    }
}

/// Parameters allowed for `core.list_tasks`.
#[derive(Default, Deserialize)]
#[serde(default)]
pub(crate) struct ListTasksParams {
    pub status: Option<TaskStatus>,
    pub tags: Option<Vec<String>>,
    pub text_search: Option<String>,
    pub touched_since: Option<String>,
}

/// Parameters for `core.task_detail`.
#[derive(Deserialize)]
pub(crate) struct TaskDetailParams {
    pub task_id: String,
}

/// Parameters for `core.items_for_tag`.
#[derive(Deserialize)]
pub(crate) struct TagParams {
    pub tag: String,
}

/// Shared params for range-based queries (notes + weekly summary).
#[derive(Debug, Deserialize)]
pub(crate) struct RangeParams {
    pub start: String,
    pub end: String,
}

/// Parameters for `core.open_daily`.
#[derive(Deserialize)]
pub(crate) struct OpenDailyParams {
    pub date: String,
}

/// Parameters for `core.read_note`.
#[derive(Deserialize)]
pub(crate) struct NoteParams {
    pub note_id: String,
}

/// Parameters for `core.write_note`.
#[derive(Deserialize)]
pub(crate) struct WriteNoteParams {
    pub note_id: String,
    pub content: String,
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    /// Ensures invalid parameter payloads surface the expected JSON-RPC error.
    #[test]
    fn parse_params_reports_invalid_input() {
        let err = parse_params::<RangeParams>(Some(json!({ "start": 123 }))).unwrap_err();
        assert_eq!(err.code.as_i32(), RpcErrorCode::InvalidParams.as_i32());
    }
}
