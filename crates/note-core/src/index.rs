use std::collections::HashSet;

use anyhow::Result;

use crate::model::{
    DateRange, LogEntry, LogEntryId, Note, NoteId, NoteMeta, ParsedNote, TagResult, Task,
    TaskFilter, TaskId, TaskMention, VaultIndex,
};

/// Trait describing a storage backend for parsed vault data.
pub trait IndexStore {
    /// Inserts or replaces the parsed representation of a single note.
    fn upsert_parsed_note(&mut self, parsed: ParsedNote) -> Result<()>;
    /// Removes everything associated with the provided note from the index.
    fn remove_note(&mut self, note_id: &NoteId) -> Result<()>;

    /// Fetches a single task by ID.
    fn get_task(&self, id: &TaskId) -> Option<Task>;
    /// Lists tasks matching the provided filter criteria.
    fn list_tasks(&self, filter: &TaskFilter) -> Vec<Task>;
    /// Returns log entries that reference the provided task.
    fn get_log_entries_for_task(&self, id: &TaskId) -> Vec<LogEntry>;
    /// Returns backlinks discovered for the provided task.
    fn get_mentions_for_task(&self, id: &TaskId) -> Vec<TaskMention>;

    /// Returns `NoteMeta` entries whose dates fall within the range.
    fn list_notes_by_date(&self, range: &DateRange) -> Vec<NoteMeta>;
    /// Fetches the fully loaded note (including Markdown content).
    fn get_note(&self, id: &NoteId) -> Option<Note>;

    /// Lists every tag present in the index.
    fn list_tags(&self) -> Vec<String>;
    /// Returns all tasks/log entries tied to a single tag.
    fn items_for_tag(&self, tag: &str) -> TagResult;
}

/// In-memory `IndexStore` backed by hash maps—useful for now and tests.
#[derive(Default)]
pub struct InMemoryIndexStore {
    data: VaultIndex,
}

impl InMemoryIndexStore {
    /// Creates a brand-new empty index store.
    pub fn new() -> Self {
        Self::default()
    }

    /// Cleans tag/mention lookups when a task is removed.
    fn remove_task_internal(&mut self, task_id: &TaskId) {
        if let Some(task) = self.data.tasks.remove(task_id) {
            for tag in task.tags {
                if let Some(ids) = self.data.tags_to_tasks.get_mut(&tag) {
                    ids.retain(|id| id != task_id);
                    if ids.is_empty() {
                        self.data.tags_to_tasks.remove(&tag);
                    }
                }
            }
        }
        self.data.mentions_by_task.remove(task_id);
        self.data.task_refs_by_task.remove(task_id);
    }

    /// Removes a log entry and clears any tag/task reverse references.
    fn remove_log_entry_internal(&mut self, entry_id: &LogEntryId) {
        if let Some(entry) = self.data.log_entries.remove(entry_id) {
            for tag in entry.tags {
                if let Some(ids) = self.data.tags_to_log_entries.get_mut(&tag) {
                    ids.retain(|id| id != entry_id);
                    if ids.is_empty() {
                        self.data.tags_to_log_entries.remove(&tag);
                    }
                }
            }
            for task_id in entry.task_ids {
                if let Some(refs) = self.data.task_refs_by_task.get_mut(&task_id) {
                    refs.retain(|id| id != entry_id);
                    if refs.is_empty() {
                        self.data.task_refs_by_task.remove(&task_id);
                    }
                }
            }
        }
    }
}

impl IndexStore for InMemoryIndexStore {
    /// Applies the parsed note to the index, replacing previous data for that note.
    fn upsert_parsed_note(&mut self, parsed: ParsedNote) -> Result<()> {
        let ParsedNote {
            note,
            tasks,
            log_entries,
            mentions,
        } = parsed;
        let note_id = note.id.clone();
        self.remove_note(&note_id)?;

        let meta = NoteMeta {
            id: note.id.clone(),
            path: note.path.clone(),
            title: note.title.clone(),
            date: note.date,
        };
        self.data.notes.insert(note_id.clone(), meta);
        self.data.note_content.insert(note_id.clone(), note);

        let mut task_ids = Vec::new();
        for task in tasks {
            for tag in &task.tags {
                self.data
                    .tags_to_tasks
                    .entry(tag.clone())
                    .or_default()
                    .push(task.id.clone());
            }
            task_ids.push(task.id.clone());
            self.data.tasks.insert(task.id.clone(), task);
        }
        if !task_ids.is_empty() {
            self.data.note_to_task_ids.insert(note_id.clone(), task_ids);
        }

        let mut log_entry_ids = Vec::new();
        for entry in log_entries {
            for tag in &entry.tags {
                self.data
                    .tags_to_log_entries
                    .entry(tag.clone())
                    .or_default()
                    .push(entry.id.clone());
            }
            for task_id in &entry.task_ids {
                self.data
                    .task_refs_by_task
                    .entry(task_id.clone())
                    .or_default()
                    .push(entry.id.clone());
            }
            log_entry_ids.push(entry.id.clone());
            self.data.log_entries.insert(entry.id.clone(), entry);
        }
        if !log_entry_ids.is_empty() {
            self.data
                .note_to_log_entry_ids
                .insert(note_id.clone(), log_entry_ids);
        }

        for mention in mentions {
            self.data
                .mentions_by_task
                .entry(mention.task_id.clone())
                .or_default()
                .push(mention);
        }

        Ok(())
    }

    /// Removes all trace of the given note and any derived entities.
    fn remove_note(&mut self, note_id: &NoteId) -> Result<()> {
        self.data.notes.remove(note_id);
        self.data.note_content.remove(note_id);

        if let Some(ids) = self.data.note_to_task_ids.remove(note_id) {
            for id in ids {
                self.remove_task_internal(&id);
            }
        }
        if let Some(ids) = self.data.note_to_log_entry_ids.remove(note_id) {
            for id in ids {
                self.remove_log_entry_internal(&id);
            }
        }

        for mentions in self.data.mentions_by_task.values_mut() {
            mentions.retain(|mention| &mention.note_id != note_id);
        }

        Ok(())
    }

    /// Returns a cloned task if the ID exists.
    fn get_task(&self, id: &TaskId) -> Option<Task> {
        self.data.tasks.get(id).cloned()
    }

    /// Applies filtering logic for status/tags/search/time slices across tasks.
    fn list_tasks(&self, filter: &TaskFilter) -> Vec<Task> {
        self.data
            .tasks
            .values()
            .cloned()
            .filter(|task| match filter.status {
                Some(ref status) => &task.status == status,
                None => true,
            })
            .filter(|task| {
                if filter.tags.is_empty() {
                    return true;
                }
                let task_tags: HashSet<_> = task.tags.iter().collect();
                filter.tags.iter().all(|tag| task_tags.contains(&tag))
            })
            .filter(|task| {
                if let Some(ref search) = filter.text_search {
                    let search = search.to_lowercase();
                    task.title.to_lowercase().contains(&search)
                        || task
                            .description_md
                            .as_ref()
                            .map(|desc| desc.to_lowercase().contains(&search))
                            .unwrap_or(false)
                } else {
                    true
                }
            })
            .filter(|task| {
                if let Some(date) = filter.touched_since {
                    task.updated_at.date() >= date
                        || task
                            .closed_at
                            .map(|closed| closed.date() >= date)
                            .unwrap_or(false)
                } else {
                    true
                }
            })
            .collect()
    }

    /// Retrieves all log entries that reference the provided task.
    fn get_log_entries_for_task(&self, id: &TaskId) -> Vec<LogEntry> {
        self.data
            .task_refs_by_task
            .get(id)
            .into_iter()
            .flat_map(|ids| ids.iter())
            .filter_map(|entry_id| self.data.log_entries.get(entry_id))
            .cloned()
            .collect()
    }

    /// Returns backlink mentions for a task if recorded.
    fn get_mentions_for_task(&self, id: &TaskId) -> Vec<TaskMention> {
        self.data
            .mentions_by_task
            .get(id)
            .cloned()
            .unwrap_or_default()
    }

    /// Lists `NoteMeta` entries ordered by date within the requested range.
    fn list_notes_by_date(&self, range: &DateRange) -> Vec<NoteMeta> {
        let mut notes: Vec<_> = self
            .data
            .notes
            .values()
            .filter(|meta| {
                if let Some(date) = meta.date {
                    date >= range.start && date <= range.end
                } else {
                    false
                }
            })
            .cloned()
            .collect();
        notes.sort_by_key(|meta| meta.date);
        notes
    }

    /// Returns a note with full content if present in the index.
    fn get_note(&self, id: &NoteId) -> Option<Note> {
        self.data.note_content.get(id).cloned()
    }

    /// Produces a sorted list of all known tags.
    fn list_tags(&self) -> Vec<String> {
        let mut tags: HashSet<String> = HashSet::new();
        tags.extend(self.data.tags_to_tasks.keys().cloned());
        tags.extend(self.data.tags_to_log_entries.keys().cloned());
        let mut tags: Vec<_> = tags.into_iter().collect();
        tags.sort();
        tags
    }

    /// Fetches every task/log entry associated with the provided tag string.
    fn items_for_tag(&self, tag: &str) -> TagResult {
        let tasks = self
            .data
            .tags_to_tasks
            .get(tag)
            .into_iter()
            .flat_map(|ids| ids.iter())
            .filter_map(|id| self.data.tasks.get(id))
            .cloned()
            .collect();

        let log_entries = self
            .data
            .tags_to_log_entries
            .get(tag)
            .into_iter()
            .flat_map(|ids| ids.iter())
            .filter_map(|id| self.data.log_entries.get(id))
            .cloned()
            .collect();

        TagResult {
            tag: tag.to_string(),
            tasks,
            log_entries,
        }
    }
}
