package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/jack/yapper/go-note/internal/core"
)

// Request models the JSON-RPC 2.0 request envelope.
type Request struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params"`
}

// Response represents either a successful or failed JSON-RPC call.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *ErrorBody      `json:"error,omitempty"`
}

// ErrorBody matches the JSON-RPC error object.
type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorCode enumerates JSON-RPC error codes.
type ErrorCode int

const (
	CodeParseError     ErrorCode = -32700
	CodeInvalidRequest ErrorCode = -32600
	CodeMethodNotFound ErrorCode = -32601
	CodeInvalidParams  ErrorCode = -32602
	CodeInternalError  ErrorCode = -32603
	CodeServerError    ErrorCode = -32000
)

// Error is the application-level error propagated to the RPC layer.
type Error struct {
	Code    ErrorCode
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (code %d)", e.Message, e.Code)
}

// ResponseResult builds a success response.
func ResponseResult(id json.RawMessage, result interface{}) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// ResponseError builds an error response.
func ResponseError(id json.RawMessage, err Error) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrorBody{
			Code:    int(err.Code),
			Message: err.Message,
		},
	}
}

// NullID returns a JSON null identifier placeholder.
func NullID() json.RawMessage {
	return json.RawMessage("null")
}

// ParseParams decodes params into the supplied struct, defaulting to {}.
func ParseParams[T any](params json.RawMessage) (T, Error) {
	var out T
	payload := params
	if len(payload) == 0 || string(payload) == "null" {
		payload = []byte("{}")
	}
	if err := json.Unmarshal(payload, &out); err != nil {
		return out, Error{Code: CodeInvalidParams, Message: err.Error()}
	}
	return out, Error{}
}

// ParseDate parses YYYY-MM-DD inputs into a core.Date.
func ParseDate(value string) (core.Date, Error) {
	date, err := core.ParseDate(value)
	if err != nil {
		return core.Date{}, Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid date %q", value),
		}
	}
	return date, Error{}
}

// ParseRange builds a DateRange from ISO date strings.
func ParseRange(start, end string) (core.DateRange, Error) {
	startDate, err := core.ParseDate(start)
	if err != nil {
		return core.DateRange{}, Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid start date %q", start),
		}
	}
	endDate, err := core.ParseDate(end)
	if err != nil {
		return core.DateRange{}, Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid end date %q", end),
		}
	}
	return core.DateRange{Start: startDate, End: endDate}, Error{}
}

// Convenience constructors for common error responses.
func ParseError(message string) Error {
	return Error{Code: CodeParseError, Message: message}
}

func InvalidRequest(message string) Error {
	return Error{Code: CodeInvalidRequest, Message: message}
}

func InvalidParams(message string) Error {
	return Error{Code: CodeInvalidParams, Message: message}
}

func MethodNotFound(method string) Error {
	return Error{Code: CodeMethodNotFound, Message: fmt.Sprintf("unknown method %q", method)}
}

func InternalError(message string) Error {
	return Error{Code: CodeInternalError, Message: message}
}

func ServerError(message string) Error {
	return Error{Code: CodeServerError, Message: message}
}

// Parameter payloads ---------------------------------------------------------

type ListTasksParams struct {
	Status       *core.TaskStatus `json:"status"`
	Tags         []string         `json:"tags"`
	TextSearch   *string          `json:"text_search"`
	TouchedSince *string          `json:"touched_since"`
}

type TaskDetailParams struct {
	TaskID string `json:"task_id"`
}

type TagParams struct {
	Tag string `json:"tag"`
}

type RangeParams struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type OpenDailyParams struct {
	Date string `json:"date"`
}

type NoteParams struct {
	NoteID string `json:"note_id"`
}

type WriteNoteParams struct {
	NoteID  string `json:"note_id"`
	Content string `json:"content"`
}
