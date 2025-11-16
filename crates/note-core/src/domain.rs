use std::collections::HashMap;
use std::fs;
use std::path::Path;

use anyhow::{Result, anyhow};
use chrono::NaiveDate;

use crate::index::IndexStore;
use crate::model::{
    DateRange, Note, NoteId, NoteMeta, TagResult, Task, TaskFilter, TaskId, TaskMention,
    WeeklySummary,
};
use crate::parser::NoteParser;
use crate::vault::Vault;

pub struct VaultIndexManager<V: Vault, I: IndexStore> {
    pub vault: V,
    pub index: I,
    parser: Box<dyn NoteParser>,
}

impl<V: Vault, I: IndexStore> VaultIndexManager<V, I> {
    pub fn new(vault: V, index: I, parser: Box<dyn NoteParser>) -> Self {
        Self {
            vault,
            index,
            parser,
        }
    }

    pub fn full_reindex(&mut self) -> Result<()> {
        let paths = self.vault.list_note_paths()?;
        for path in paths {
            let note = self.vault.read_note(&path)?;
            let parsed = self.parser.parse(note);
            self.index.upsert_parsed_note(parsed)?;
        }
        Ok(())
    }

    pub fn reindex_note_path(&mut self, path: &Path) -> Result<()> {
        let note = self.vault.read_note(path)?;
        let parsed = self.parser.parse(note);
        self.index.upsert_parsed_note(parsed)
    }
}

pub struct Domain<V: Vault, I: IndexStore> {
    pub index_mgr: VaultIndexManager<V, I>,
}

impl<V: Vault, I: IndexStore> Domain<V, I> {
    pub fn new(index_mgr: VaultIndexManager<V, I>) -> Self {
        Domain { index_mgr }
    }

    pub fn reindex_all(&mut self) -> Result<()> {
        self.index_mgr.full_reindex()
    }

    pub fn list_tasks(&self, filter: &TaskFilter) -> Vec<Task> {
        self.index_mgr.index.list_tasks(filter)
    }

    pub fn task_detail(&self, id: &TaskId) -> Option<(Task, Vec<TaskMention>)> {
        let task = self.index_mgr.index.get_task(id)?;
        let mentions = self.index_mgr.index.get_mentions_for_task(id);
        Some((task, mentions))
    }

    pub fn items_for_tag(&self, tag: &str) -> TagResult {
        self.index_mgr.index.items_for_tag(tag)
    }

    pub fn notes_in_range(&self, range: &DateRange) -> Vec<NoteMeta> {
        self.index_mgr.index.list_notes_by_date(range)
    }

    pub fn list_tags(&self) -> Vec<String> {
        self.index_mgr.index.list_tags()
    }

    pub fn read_note(&self, id: &NoteId) -> Option<Note> {
        self.index_mgr.index.get_note(id)
    }

    pub fn write_note(&mut self, id: &NoteId, content: &str) -> Result<Note> {
        let note = self
            .index_mgr
            .index
            .get_note(id)
            .ok_or_else(|| anyhow!("note {} not indexed", id.0))?;
        fs::write(&note.path, content)?;
        self.index_mgr.reindex_note_path(&note.path)?;
        self.index_mgr
            .index
            .get_note(id)
            .ok_or_else(|| anyhow!("note {} unavailable after write", id.0))
    }

    pub fn open_daily(&mut self, date: NaiveDate) -> Result<Note> {
        let range = DateRange {
            start: date,
            end: date,
        };
        if let Some(existing) = self
            .index_mgr
            .index
            .list_notes_by_date(&range)
            .into_iter()
            .find(|meta| meta.date == Some(date))
        {
            if let Some(note) = self.index_mgr.index.get_note(&existing.id) {
                return Ok(note);
            }
        }

        let file_name = format!("{}.md", date.format("%Y-%m-%d"));
        let path = self.index_mgr.vault.root_path().join(file_name);

        if !path.exists() {
            if let Some(parent) = path.parent() {
                fs::create_dir_all(parent)?;
            }
            let template = format!("# {}\n\n## Tasks\n\n## Log\n", date.format("%Y-%m-%d"));
            fs::write(&path, template)?;
        }

        self.index_mgr.reindex_note_path(&path)?;
        let id = NoteId(path.to_string_lossy().to_string());
        self.index_mgr
            .index
            .get_note(&id)
            .ok_or_else(|| anyhow!("note {} unavailable after creation", id.0))
    }

    pub fn weekly_summary(&self, range: &DateRange) -> WeeklySummary {
        let all_tasks = self.index_mgr.index.list_tasks(&TaskFilter::default());
        let new_tasks: Vec<_> = all_tasks
            .iter()
            .cloned()
            .filter(|task| {
                let created = task.created_at.date();
                created >= range.start && created <= range.end
            })
            .collect();

        let completed_tasks: Vec<_> = all_tasks
            .iter()
            .cloned()
            .filter(|task| {
                task.closed_at
                    .map(|closed| {
                        let date = closed.date();
                        date >= range.start && date <= range.end
                    })
                    .unwrap_or(false)
            })
            .collect();

        let notes = self.index_mgr.index.list_notes_by_date(range);
        let mut tag_counts: HashMap<String, usize> = HashMap::new();
        for task in &all_tasks {
            for tag in &task.tags {
                *tag_counts.entry(tag.clone()).or_default() += 1;
            }
        }
        let mut top_tags: Vec<_> = tag_counts.into_iter().collect();
        top_tags.sort_by(|a, b| b.1.cmp(&a.1));
        top_tags.truncate(10);

        WeeklySummary {
            new_tasks,
            completed_tasks,
            notes,
            top_tags,
        }
    }
}
