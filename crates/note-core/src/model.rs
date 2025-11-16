use std::collections::HashMap;
use std::path::PathBuf;

use chrono::{NaiveDate, NaiveDateTime};
use serde::{Deserialize, Serialize};

/// Unique identifier for a note inside the vault (usually the full path string).
#[derive(Clone, Debug, Eq, PartialEq, Hash, Serialize, Deserialize)]
#[serde(transparent)]
pub struct NoteId(pub String);

/// Stable identifier for a task (e.g. human-facing `T-2025-001`).
#[derive(Clone, Debug, Eq, PartialEq, Hash, Serialize, Deserialize)]
#[serde(transparent)]
pub struct TaskId(pub String);

/// Identifier tying a log entry to its originating note and line.
#[derive(Clone, Debug, Eq, PartialEq, Hash, Serialize, Deserialize)]
#[serde(transparent)]
pub struct LogEntryId(pub String);

/// Full fidelity representation of a Markdown note loaded from the vault.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Note {
    pub id: NoteId,
    #[serde(with = "path_buf_serde")]
    pub path: PathBuf,
    pub title: String,
    pub date: Option<NaiveDate>,
    pub content: String,
}

/// Canonical lifecycle state for a task as interpreted by the core.
#[derive(Clone, Debug, Eq, PartialEq, Serialize, Deserialize)]
pub enum TaskStatus {
    Open,
    InProgress,
    Done,
    Blocked,
}

impl Default for TaskStatus {
    /// Defaults tasks to the `Open` state when no explicit status is supplied.
    fn default() -> Self {
        TaskStatus::Open
    }
}

/// Rich task object that augments Markdown tasks with metadata and timestamps.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Task {
    pub id: TaskId,
    pub title: String,
    pub status: TaskStatus,
    pub created_at: NaiveDateTime,
    pub updated_at: NaiveDateTime,
    pub closed_at: Option<NaiveDateTime>,
    pub tags: Vec<String>,
    pub description_md: Option<String>,
    pub source_note_id: Option<NoteId>,
}

/// Parsed representation of a single log entry bullet in the `## Log` section.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct LogEntry {
    pub id: LogEntryId,
    pub note_id: NoteId,
    pub line_number: usize,
    pub timestamp: Option<String>,
    pub content_md: String,
    pub tags: Vec<String>,
    pub task_ids: Vec<TaskId>,
}

/// Backlink created whenever a log entry references a task ID.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct TaskMention {
    pub task_id: TaskId,
    pub note_id: NoteId,
    pub log_entry_id: Option<LogEntryId>,
    pub excerpt: String,
}

/// Container returned by the parser that includes the note and derived entities.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct ParsedNote {
    pub note: Note,
    pub tasks: Vec<Task>,
    pub log_entries: Vec<LogEntry>,
    pub mentions: Vec<TaskMention>,
}

/// Lightweight metadata for list views when full note contents are unnecessary.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct NoteMeta {
    pub id: NoteId,
    #[serde(with = "path_buf_serde")]
    pub path: PathBuf,
    pub title: String,
    pub date: Option<NaiveDate>,
}

/// Aggregate state table used by the in-memory index implementation.
#[derive(Default, Debug)]
pub struct VaultIndex {
    pub notes: HashMap<NoteId, NoteMeta>,
    pub note_content: HashMap<NoteId, Note>,
    pub tasks: HashMap<TaskId, Task>,
    pub log_entries: HashMap<LogEntryId, LogEntry>,
    pub mentions_by_task: HashMap<TaskId, Vec<TaskMention>>,
    pub tags_to_tasks: HashMap<String, Vec<TaskId>>,
    pub tags_to_log_entries: HashMap<String, Vec<LogEntryId>>,
    pub task_refs_by_task: HashMap<TaskId, Vec<LogEntryId>>,
    pub note_to_task_ids: HashMap<NoteId, Vec<TaskId>>,
    pub note_to_log_entry_ids: HashMap<NoteId, Vec<LogEntryId>>,
}

/// Filtering options when querying for tasks across the vault.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct TaskFilter {
    pub status: Option<TaskStatus>,
    pub tags: Vec<String>,
    pub text_search: Option<String>,
    pub touched_since: Option<NaiveDate>,
}

impl Default for TaskFilter {
    /// Constructs an all-inclusive filter (no constraints applied).
    fn default() -> Self {
        Self {
            status: None,
            tags: Vec::new(),
            text_search: None,
            touched_since: None,
        }
    }
}

/// Inclusive range struct used for date bounded queries.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct DateRange {
    pub start: NaiveDate,
    pub end: NaiveDate,
}

/// Result payload for `items_for_tag` that bundles all matching entities.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct TagResult {
    pub tag: String,
    pub tasks: Vec<Task>,
    pub log_entries: Vec<LogEntry>,
}

/// Summary statistics covering a time window for weekly review flows.
#[derive(Clone, Debug, Default, Serialize, Deserialize)]
pub struct WeeklySummary {
    pub new_tasks: Vec<Task>,
    pub completed_tasks: Vec<Task>,
    pub notes: Vec<NoteMeta>,
    pub top_tags: Vec<(String, usize)>,
}

mod path_buf_serde {
    use std::path::PathBuf;

    use serde::{Deserialize, Deserializer, Serializer};

    /// Serializes a `PathBuf` to a string for serde-friendly models.
    pub fn serialize<S>(path: &PathBuf, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_str(&path.to_string_lossy())
    }

    /// Deserializes a `PathBuf` from a stored string representation.
    pub fn deserialize<'de, D>(deserializer: D) -> Result<PathBuf, D::Error>
    where
        D: Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        Ok(PathBuf::from(s))
    }
}
