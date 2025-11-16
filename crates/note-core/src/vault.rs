use std::fs;
use std::path::{Path, PathBuf};

use anyhow::Context;
use anyhow::Result;
use chrono::NaiveDate;

use crate::model::{Note, NoteId};

/// Abstraction over the Markdown vault storage backend.
pub trait Vault {
    /// Returns every Markdown note path under the vault root.
    fn list_note_paths(&self) -> Result<Vec<PathBuf>>;
    /// Reads and materializes a single note from disk.
    fn read_note(&self, path: &Path) -> Result<Note>;
    /// Returns the configured root directory for the vault.
    fn root_path(&self) -> &Path;
}

/// Concrete vault implementation that talks to the local filesystem.
#[derive(Clone, Debug)]
pub struct FileSystemVault {
    root: PathBuf,
}

impl FileSystemVault {
    /// Creates a new vault rooted at the provided path.
    pub fn new(root: PathBuf) -> Self {
        FileSystemVault { root }
    }

    /// Recursively discovers markdown files underneath `dir`.
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

    /// Uses the file stem or fallback display name as a note title.
    fn derive_title(path: &Path) -> String {
        path.file_stem()
            .and_then(|s| s.to_str())
            .map(|s| s.to_string())
            .unwrap_or_else(|| path.display().to_string())
    }

    /// Attempts to parse a date from the filename using YYYY-MM-DD or YY-MM-DD.
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
    /// Enumerates every Markdown note relative to the vault root.
    fn list_note_paths(&self) -> Result<Vec<PathBuf>> {
        Self::gather_markdown_files(&self.root)
    }

    /// Reads the note contents plus metadata for the provided path.
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

    /// Gives callers access to the root so helpers like `open_daily` can create files.
    fn root_path(&self) -> &Path {
        &self.root
    }
}
