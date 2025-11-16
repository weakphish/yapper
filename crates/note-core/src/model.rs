use std::collections::HashMap;
use std::path::PathBuf;

use chrono::{NaiveDate, NaiveDateTime};
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Eq, PartialEq, Hash, Serialize, Deserialize)]
#[serde(transparent)]
pub struct NoteId(pub String);

#[derive(Clone, Debug, Eq, PartialEq, Hash, Serialize, Deserialize)]
#[serde(transparent)]
pub struct TaskId(pub String);

#[derive(Clone, Debug, Eq, PartialEq, Hash, Serialize, Deserialize)]
#[serde(transparent)]
pub struct LogEntryId(pub String);

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Note {
    pub id: NoteId,
    #[serde(with = "path_buf_serde")]
    pub path: PathBuf,
    pub title: String,
    pub date: Option<NaiveDate>,
    pub content: String,
}

#[derive(Clone, Debug, Eq, PartialEq, Serialize, Deserialize)]
pub enum TaskStatus {
    Open,
    InProgress,
    Done,
    Blocked,
}

impl Default for TaskStatus {
    fn default() -> Self {
        TaskStatus::Open
    }
}

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

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct TaskMention {
    pub task_id: TaskId,
    pub note_id: NoteId,
    pub log_entry_id: Option<LogEntryId>,
    pub excerpt: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct ParsedNote {
    pub note: Note,
    pub tasks: Vec<Task>,
    pub log_entries: Vec<LogEntry>,
    pub mentions: Vec<TaskMention>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct NoteMeta {
    pub id: NoteId,
    #[serde(with = "path_buf_serde")]
    pub path: PathBuf,
    pub title: String,
    pub date: Option<NaiveDate>,
}

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

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct TaskFilter {
    pub status: Option<TaskStatus>,
    pub tags: Vec<String>,
    pub text_search: Option<String>,
    pub touched_since: Option<NaiveDate>,
}

impl Default for TaskFilter {
    fn default() -> Self {
        Self {
            status: None,
            tags: Vec::new(),
            text_search: None,
            touched_since: None,
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct DateRange {
    pub start: NaiveDate,
    pub end: NaiveDate,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct TagResult {
    pub tag: String,
    pub tasks: Vec<Task>,
    pub log_entries: Vec<LogEntry>,
}

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

    pub fn serialize<S>(path: &PathBuf, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_str(&path.to_string_lossy())
    }

    pub fn deserialize<'de, D>(deserializer: D) -> Result<PathBuf, D::Error>
    where
        D: Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        Ok(PathBuf::from(s))
    }
}
