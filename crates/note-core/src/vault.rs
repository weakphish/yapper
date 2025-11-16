use std::fs;
use std::path::{Path, PathBuf};

use anyhow::Context;
use anyhow::Result;
use chrono::NaiveDate;

use crate::model::{Note, NoteId};

pub trait Vault {
    fn list_note_paths(&self) -> Result<Vec<PathBuf>>;
    fn read_note(&self, path: &Path) -> Result<Note>;
    fn root_path(&self) -> &Path;
}

#[derive(Clone, Debug)]
pub struct FileSystemVault {
    root: PathBuf,
}

impl FileSystemVault {
    pub fn new(root: PathBuf) -> Self {
        FileSystemVault { root }
    }

    fn gather_markdown_files(dir: &Path) -> Result<Vec<PathBuf>> {
        let mut stack = vec![dir.to_path_buf()];
        let mut found = Vec::new();

        while let Some(current) = stack.pop() {
            let entries = fs::read_dir(&current)
                .with_context(|| format!("failed to list entries under {}", current.display()))?;

            for entry in entries {
                let entry = entry?;
                let path = entry.path();
                if path.is_dir() {
                    stack.push(path);
                } else if path
                    .extension()
                    .and_then(|ext| ext.to_str())
                    .map(|ext| ext.eq_ignore_ascii_case("md"))
                    .unwrap_or(false)
                {
                    found.push(path);
                }
            }
        }

        found.sort();
        Ok(found)
    }

    fn derive_title(path: &Path) -> String {
        path.file_stem()
            .and_then(|s| s.to_str())
            .map(|s| s.to_string())
            .unwrap_or_else(|| path.display().to_string())
    }

    fn derive_date(path: &Path) -> Option<NaiveDate> {
        let name = path.file_stem()?.to_string_lossy();
        if let Ok(date) = NaiveDate::parse_from_str(&name, "%Y-%m-%d") {
            return Some(date);
        }
        if let Ok(date) = NaiveDate::parse_from_str(&name, "%y-%m-%d") {
            return Some(date);
        }
        None
    }
}

impl Vault for FileSystemVault {
    fn list_note_paths(&self) -> Result<Vec<PathBuf>> {
        Self::gather_markdown_files(&self.root)
    }

    fn read_note(&self, path: &Path) -> Result<Note> {
        let content = fs::read_to_string(path)
            .with_context(|| format!("failed to read {}", path.display()))?;
        let title = Self::derive_title(path);
        let date = Self::derive_date(path);
        let id = NoteId(path.to_string_lossy().to_string());

        Ok(Note {
            id,
            path: path.to_path_buf(),
            title,
            date,
            content,
        })
    }

    fn root_path(&self) -> &Path {
        &self.root
    }
}
