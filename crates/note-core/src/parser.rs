use std::iter::Peekable;
use std::str::Lines;

use chrono::Utc;
use once_cell::sync::Lazy;
use regex::{Captures, Regex};

use crate::model::{LogEntry, LogEntryId, Note, ParsedNote, Task, TaskId, TaskMention, TaskStatus};

pub trait NoteParser: Send + Sync {
    fn parse(&self, note: Note) -> ParsedNote;
}

#[derive(Default)]
pub struct MarkdownParser;

impl MarkdownParser {
    pub fn new() -> Self {
        Self
    }
}

impl NoteParser for MarkdownParser {
    fn parse(&self, note: Note) -> ParsedNote {
        parse_note_internal(note)
    }
}

static TASK_LINE_RE: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r#"^\s*-\s*\[(?P<mark> |x)\]\s+\[(?P<id>T-[0-9A-Za-z_-]+)\]\s*(?P<rest>.*)$"#)
        .expect("valid task regex")
});

static TIME_PREFIX_RE: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r#"^\s*-\s*(?P<time>[0-9]{1,2}:[0-9]{2}(?:\s?(?:am|pm|AM|PM))?)\s+(?P<rest>.*)$"#)
        .expect("valid time regex")
});

static LOG_TASK_REF_RE: Lazy<Regex> =
    Lazy::new(|| Regex::new(r#"\[(T-[0-9A-Za-z_-]+)\]"#).expect("valid task reference regex"));

#[derive(Copy, Clone, Debug, Eq, PartialEq)]
enum Section {
    Tasks,
    Log,
    Other,
}

fn parse_note_internal(note: Note) -> ParsedNote {
    let mut tasks = Vec::new();
    let mut log_entries = Vec::new();
    let mut mentions = Vec::new();

    let mut current_section = Section::Other;
    let timestamp_now = Utc::now().naive_utc();

    let mut lines = note.content.lines().enumerate().peekable();
    while let Some((idx, raw_line)) = lines.next() {
        let line = raw_line.trim_end();
        let trimmed = line.trim_start();
        if let Some(stripped) = trimmed.strip_prefix("## ") {
            let heading = stripped.trim();
            current_section = if heading.eq_ignore_ascii_case("tasks") {
                Section::Tasks
            } else if heading.eq_ignore_ascii_case("log") {
                Section::Log
            } else {
                Section::Other
            };
            continue;
        }

        match current_section {
            Section::Tasks => {
                if let Some(caps) = TASK_LINE_RE.captures(trimmed) {
                    let continuation = collect_continuation(&mut lines);
                    let rest = caps.name("rest").map(|m| m.as_str()).unwrap_or("").trim();
                    let combined = combine_with_continuation(rest, &continuation);
                    let task = build_task_from_caps(&note, &caps, &combined, timestamp_now);
                    tasks.push(task);
                }
            }
            Section::Log => {
                if let Some((entry, entry_mentions)) =
                    parse_log_line(&note, raw_line, idx + 1, &mut lines)
                {
                    mentions.extend(entry_mentions);
                    log_entries.push(entry);
                }
            }
            Section::Other => {}
        }
    }

    ParsedNote {
        note,
        tasks,
        log_entries,
        mentions,
    }
}

fn build_task_from_caps(
    note: &Note,
    caps: &Captures<'_>,
    rest: &str,
    now: chrono::NaiveDateTime,
) -> Task {
    let status = match caps.name("mark").map(|m| m.as_str()) {
        Some("x") | Some("X") => TaskStatus::Done,
        _ => TaskStatus::Open,
    };
    let task_id = TaskId(
        caps.name("id")
            .map(|m| m.as_str())
            .unwrap_or_default()
            .to_string(),
    );
    let (title, tags) = split_title_and_tags(rest);

    Task {
        id: task_id,
        title,
        status: status.clone(),
        created_at: now,
        updated_at: now,
        closed_at: if matches!(status, TaskStatus::Done) {
            Some(now)
        } else {
            None
        },
        tags,
        description_md: None,
        source_note_id: Some(note.id.clone()),
    }
}

fn parse_log_line(
    note: &Note,
    line: &str,
    line_number: usize,
    lines: &mut LineIter<'_>,
) -> Option<(LogEntry, Vec<TaskMention>)> {
    if !line.trim_start().starts_with("- ") {
        return None;
    }

    let continuation = collect_continuation(lines);
    let (timestamp, remainder_line) = if let Some(caps) = TIME_PREFIX_RE.captures(line) {
        (
            Some(caps.name("time").unwrap().as_str().to_string()),
            caps.name("rest").unwrap().as_str().trim().to_string(),
        )
    } else {
        (None, line.trim_start_matches("- ").trim().to_string())
    };

    let combined = combine_with_continuation(&remainder_line, &continuation);
    let (content_without_tags, tags) = split_title_and_tags(&combined);
    let mut task_ids = Vec::new();
    let mut mentions = Vec::new();
    for caps in LOG_TASK_REF_RE.captures_iter(&combined) {
        let id = TaskId(caps.get(1).unwrap().as_str().to_string());
        task_ids.push(id.clone());
        mentions.push(TaskMention {
            task_id: id,
            note_id: note.id.clone(),
            log_entry_id: None,
            excerpt: build_excerpt(&combined),
        });
    }

    let entry_id = LogEntryId(format!("{}:{}", note.id.0, line_number));
    for mention in &mut mentions {
        mention.log_entry_id = Some(entry_id.clone());
    }

    let entry = LogEntry {
        id: entry_id,
        note_id: note.id.clone(),
        line_number,
        timestamp,
        content_md: content_without_tags,
        tags,
        task_ids,
    };

    Some((entry, mentions))
}

fn combine_with_continuation(base: &str, continuation: &[String]) -> String {
    let mut combined = base.trim().to_string();
    for extra in continuation {
        if !combined.is_empty() {
            combined.push('\n');
        }
        combined.push_str(extra.trim());
    }
    combined
}

type LineIter<'a> = Peekable<std::iter::Enumerate<Lines<'a>>>;

fn collect_continuation<'a>(lines: &mut LineIter<'a>) -> Vec<String> {
    let mut extras = Vec::new();
    while let Some((_, next_line)) = lines.peek() {
        let trimmed = next_line.trim_start();
        if trimmed.starts_with("## ") || trimmed.starts_with("- ") || next_line.trim().is_empty() {
            break;
        }
        if next_line.starts_with(' ') || next_line.starts_with('\t') {
            if let Some((_, line)) = lines.next() {
                extras.push(line.trim().to_string());
            }
        } else {
            break;
        }
    }
    extras
}

fn split_title_and_tags(input: &str) -> (String, Vec<String>) {
    let mut title_parts = Vec::new();
    let mut tags = Vec::new();
    let mut parts = input.split_whitespace();
    while let Some(part) = parts.next() {
        if let Some(tag) = part.strip_prefix('#') {
            if !tag.is_empty() {
                tags.push(tag.to_string());
                continue;
            }
        }
        title_parts.push(part);
    }
    let title = if title_parts.is_empty() {
        input.trim().to_string()
    } else {
        title_parts.join(" ")
    };
    (title, tags)
}

fn build_excerpt(content: &str) -> String {
    const MAX_EXCERPT: usize = 120;
    if content.len() <= MAX_EXCERPT {
        content.to_string()
    } else {
        let mut excerpt = content[..MAX_EXCERPT].to_string();
        excerpt.push_str("…");
        excerpt
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::model::{Note, NoteId};
    use std::path::PathBuf;

    fn make_note(content: &str) -> Note {
        Note {
            id: NoteId("note-1".into()),
            path: PathBuf::from("note.md"),
            title: "note".into(),
            date: None,
            content: content.into(),
        }
    }

    #[test]
    fn parses_multiline_tasks_and_logs() {
        let parser = MarkdownParser::new();
        let content = r#"
# 2025-03-15

## Tasks
- [ ] [T-2025-001] Implement parser #projects/note-app
  capture multiline context
  more detail #people/jack
- [x] [T-2025-002] Finish docs

## Log
- 09:10 Investigated spec [T-2025-001] #projects/note-app
  linked follow-up [T-2025-002]
- Wrote general notes
"#;
        let note = make_note(content);
        let parsed = parser.parse(note);

        assert_eq!(parsed.tasks.len(), 2);
        let first_task = &parsed.tasks[0];
        assert_eq!(first_task.id.0, "T-2025-001");
        assert!(first_task.title.contains("Implement parser"));
        assert!(first_task.title.contains("capture multiline context"));
        assert_eq!(
            first_task.tags,
            vec!["projects/note-app".to_string(), "people/jack".to_string()]
        );

        assert_eq!(parsed.log_entries.len(), 2);
        let first_entry = &parsed.log_entries[0];
        assert_eq!(first_entry.timestamp.as_deref(), Some("09:10"));
        assert_eq!(first_entry.task_ids.len(), 2);
        assert!(first_entry.content_md.contains("Investigated spec"));
        assert!(first_entry.content_md.contains("linked follow-up"));

        let mentions = &parsed.mentions;
        assert_eq!(mentions.len(), 2);
        assert!(
            mentions
                .iter()
                .any(|mention| mention.task_id.0 == "T-2025-001")
        );
        assert!(
            mentions
                .iter()
                .any(|mention| mention.task_id.0 == "T-2025-002")
        );
    }

    #[test]
    fn parses_ampm_timestamps() {
        let parser = MarkdownParser::new();
        let content = r#"
## Log
- 10:45pm Reviewed plan
"#;
        let parsed = parser.parse(make_note(content));
        assert_eq!(parsed.log_entries.len(), 1);
        assert_eq!(parsed.log_entries[0].timestamp.as_deref(), Some("10:45pm"));
    }

    #[test]
    fn tolerates_indented_headings() {
        let parser = MarkdownParser::new();
        let content = r#"
  ## Tasks
  - [ ] [T-2025-010] Indented heading still works

    ## Log
    - 12:00 Something [T-2025-010]
"#;
        let parsed = parser.parse(make_note(content));
        assert_eq!(parsed.tasks.len(), 1);
        assert_eq!(parsed.tasks[0].id.0, "T-2025-010");
        assert_eq!(parsed.log_entries.len(), 1);
        assert_eq!(parsed.mentions.len(), 1);
    }
}
