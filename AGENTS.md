# Note System Core – Technical Specification (Rust + JSON-RPC)

## 0. Context & Goals

You (the user) currently:

* Live in **Neovim**.
* Take notes in an **Obsidian-style Markdown vault** (folder of `.md` files).
* Use a **daily note** pattern (one note per day, e.g. `25-03-10.md`), split into:

  * `## Tasks` – checkboxes
  * `## Log` – bullet log of work/ideas
* Use **hierarchical tags** (e.g. `projects/note-app`, `people/jack`) as plain strings.
* Want to **upgrade** this workflow to:

  * Make **tasks first-class**, globally addressable and queryable.
  * Link **log entries ↔ tasks** with backlinks.
  * Keep **plain Markdown as the canonical storage**.
  * Support **rich views**: open tasks, per-project, weekly review, etc.
  * Eventually expose this via:

    * A **Neovim plugin** (first target) over JSON-RPC.
    * Possibly a standalone TUI later.

Key architectural constraints:

* **Rust** as the primary implementation language.
* **Vault of Markdown files is the source of truth**.
* **Index is derived**, in-memory at first, with an optional SQLite backend later.
* **Editor-agnostic core**: Neovim is just one frontend via JSON-RPC.

This spec describes that **core Rust library/binary** and its JSON-RPC API.

---

## 1. High-Level Architecture

Layers (bottom → top):

1. **Vault layer**

   * Filesystem access to a “vault” directory of `.md` files.
   * Knows how to list notes and read file contents.

2. **Parser layer**

   * Parses a `Note` (Markdown file) into structured elements:

     * Tasks
     * Log entries
     * Tags
     * Task mentions (backlinks)

3. **Index layer**

   * Maintains an index of all parsed data:

     * Notes metadata
     * Tasks
     * Log entries
     * Mentions
     * Tag → items mapping
   * First implementation: in-memory.
   * Later: an interchangeable SQLite-backed implementation.

4. **Domain / Query layer**

   * Exposes high-level operations:

     * Open/create daily note
     * List tasks with filters
     * Get task details + mentions
     * List notes in date range
     * Weekly summary, tag views, etc.

5. **JSON-RPC Server (adapter layer)**

   * A thin server exposing Domain operations over JSON-RPC.
   * Neovim plugin calls this process to power UI (buffers, floating windows, etc).

---

## 2. Core Concepts & Data Model

### 2.1 Note

Represents a Markdown file in the vault.

```rust
/// Unique identifier for a note (opaque to clients).
#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub struct NoteId(pub String);

/// Basic note metadata plus content.
#[derive(Clone, Debug)]
pub struct Note {
    pub id: NoteId,
    pub path: std::path::PathBuf,
    pub title: String,            // e.g. "2025-03-10"
    pub date: Option<chrono::NaiveDate>,
    pub content: String,          // full raw markdown content
}
```

Notes are typically daily notes (one per date), but the model should not enforce this; it’s a convention.

### 2.2 Task

Tasks are first-class entities with stable IDs.

```rust
#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub struct TaskId(pub String); // e.g. "T-2025-001"

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum TaskStatus {
    Open,
    InProgress,
    Done,
    Blocked,
}

#[derive(Clone, Debug)]
pub struct Task {
    pub id: TaskId,
    pub title: String,
    pub status: TaskStatus,
    pub created_at: chrono::NaiveDateTime,
    pub updated_at: chrono::NaiveDateTime,
    pub closed_at: Option<chrono::NaiveDateTime>,
    pub tags: Vec<String>,            // e.g. ["projects/note-app", "people/jack"]
    pub description_md: Option<String>,
    pub source_note_id: Option<NoteId>,
}
```

* `Task.id` is a human-readable ID like `T-2025-001`. The core should provide ID generation utilities.
* `status` is canonical here. Notes can show checkboxes reflecting this status.

### 2.3 LogEntry

A log entry is typically a bullet/line in the `## Log` section of a note.

```rust
#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub struct LogEntryId(pub String); // e.g. "<note_id>:<line_number>"

#[derive(Clone, Debug)]
pub struct LogEntry {
    pub id: LogEntryId,
    pub note_id: NoteId,
    pub line_number: usize,         // or byte offset
    pub timestamp: Option<String>,  // e.g. "10:15" if user includes it
    pub content_md: String,         // the full bullet/line markdown
    pub tags: Vec<String>,          // extracted from #tags
    pub task_ids: Vec<TaskId>,      // extracted from references like [T-2025-001]
}
```

### 2.4 TaskMention (Backlink)

Represents a reference from a note/log entry to a task.

```rust
#[derive(Clone, Debug)]
pub struct TaskMention {
    pub task_id: TaskId,
    pub note_id: NoteId,
    pub log_entry_id: Option<LogEntryId>,
    pub excerpt: String,           // short snippet for display in UI
}
```

### 2.5 In-Memory Index

Structure used by the in-memory index implementation:

```rust
#[derive(Default)]
pub struct VaultIndex {
    pub notes: std::collections::HashMap<NoteId, NoteMeta>,
    pub tasks: std::collections::HashMap<TaskId, Task>,
    pub log_entries: std::collections::HashMap<LogEntryId, LogEntry>,
    pub mentions_by_task: std::collections::HashMap<TaskId, Vec<TaskMention>>,
    pub tags_to_tasks: std::collections::HashMap<String, Vec<TaskId>>,
    pub tags_to_log_entries: std::collections::HashMap<String, Vec<LogEntryId>>,
    // convenience maps like notes_by_date are allowed.
}

/// Lightweight note metadata without full content (for lists).
#[derive(Clone, Debug)]
pub struct NoteMeta {
    pub id: NoteId,
    pub path: std::path::PathBuf,
    pub title: String,
    pub date: Option<chrono::NaiveDate>,
}
```

---

## 3. Markdown Conventions & Parsing Rules

The parser must be deterministic and well-specified. It processes each `Note` and emits a `ParsedNote`.

### 3.1 ParsedNote

```rust
pub struct ParsedNote {
    pub note: Note,
    pub tasks: Vec<Task>,          // tasks discovered in this note
    pub log_entries: Vec<LogEntry>,
    pub mentions: Vec<TaskMention>,
}
```

### 3.2 Sections

Convention (not enforced by the core, but parser should support this pattern):

* `## Tasks` section:

  * Contains task lines (checkboxes with task IDs).
* `## Log` section:

  * Contains log entry bullets.

Parser should:

* Find headings by scanning for `^## `.
* Treat the content under `## Tasks` and `## Log` until next `##` or EOF as that section.

### 3.3 Task Lines in Notes

In the `## Tasks` section, task lines look like:

```md
- [ ] [T-2025-001] Implement note linking in app #projects/note-app
- [x] [T-2025-002] Sync task status from notes #projects/note-app #people/jack
```

Parsing rules:

* A task line is a line starting with `- [ ]` or `- [x]`.

* Immediately after the checkbox, we expect `[T-YYYY-NNN]` as the task ID.

  * Use regex like: `^\s*-\s*\[( |x)\]\s+\[(T-[0-9]{4}-[0-9]{3,})\]\s*(.*)$`
  * `T-YYYY-NNN` is a convention, but code should not hard-code year/number lengths beyond reasonable patterns.

* The remainder of the line is the **title + tags**.

* Tags:

  * Any token starting with `#` is a tag string.
  * Keep tag strings exactly as typed (e.g. `projects/note-app`).
  * Do not attempt to type or interpret tags beyond treating them as strings.

* Status mapping:

  * `- [ ]` → `TaskStatus::Open`
  * `- [x]` → `TaskStatus::Done`
  * Other states (`InProgress`, `Blocked`) may be set via UI or derived later.

* Created/updated timestamps:

  * For now, these can be:

    * Set to “now” when first seen.
    * Or left as default and refined later.
  * The agent can provide some sensible defaults.

### 3.4 Log Entries

In the `## Log` section, treat each bullet line as a `LogEntry`. Example:

```md
## Log

- 10:15 Started exploring task [T-2025-001] #projects/note-app
- 10:45 Decided on using stable IDs like [T-2025-001] for cross-note references.
- 11:30 Talked with #people/jack about feature scope [T-2025-003].
```

Parsing rules:

* A log entry is any line beginning with `- ` within the Log section.
* Optional time prefix: capture a leading time-like token if present:

  * e.g. regex `^\s*-\s*([0-9]{1,2}:[0-9]{2})\s+(.*)$`
* Tags:

  * Any `#something` token is a tag.
* Task references:

  * Any `[T-...some id...]` pattern within the content is a task reference.
  * Use a regex like `\[(T-[A-Za-z0-9_-]+)\]` to find IDs.

`LogEntry.id` can be generated as `"<note_id>:<line_number>"` or similar.

### 3.5 Mentions

For each task reference found in a log entry:

* Create a `TaskMention`:

  * `task_id` = referenced ID.
  * `note_id` = note where it appears.
  * `log_entry_id` = id of the corresponding log entry.
  * `excerpt` = the log entry content truncated to a reasonable length.

The parser can return these mentions as part of `ParsedNote`.

---

## 4. Index Layer

### 4.1 IndexStore Trait

Abstract interface so we can plug in an in-memory or SQLite-backed index.

```rust
pub struct TaskFilter {
    pub status: Option<TaskStatus>,
    pub tags: Vec<String>,
    pub text_search: Option<String>,    // simple substring search for now
    pub touched_since: Option<chrono::NaiveDate>,
}

pub struct DateRange {
    pub start: chrono::NaiveDate,
    pub end: chrono::NaiveDate,
}

pub struct TagResult {
    pub tag: String,
    pub tasks: Vec<Task>,
    pub log_entries: Vec<LogEntry>,
}

pub trait IndexStore {
    fn upsert_parsed_note(&mut self, parsed: ParsedNote) -> anyhow::Result<()>;
    fn remove_note(&mut self, note_id: &NoteId) -> anyhow::Result<()>;

    // Query methods
    fn get_task(&self, id: &TaskId) -> Option<Task>;
    fn list_tasks(&self, filter: &TaskFilter) -> Vec<Task>;
    fn get_log_entries_for_task(&self, id: &TaskId) -> Vec<LogEntry>;
    fn get_mentions_for_task(&self, id: &TaskId) -> Vec<TaskMention>;

    fn list_notes_by_date(&self, range: &DateRange) -> Vec<NoteMeta>;
    fn get_note(&self, id: &NoteId) -> Option<Note>; // may load from vault

    fn list_tags(&self) -> Vec<String>;
    fn items_for_tag(&self, tag: &str) -> TagResult;
}
```

### 4.2 InMemoryIndexStore Implementation

* Internally uses `VaultIndex` struct.
* `upsert_parsed_note`:

  * Updates `notes` metadata.
  * Updates `tasks`, `log_entries`, `mentions_by_task`, and tag maps for this note.
  * Must clean out any old entries for this note first.

The agent should implement this as the first concrete `IndexStore`.

---

## 5. Vault & Index Management

### 5.1 Vault Trait

```rust
pub trait Vault {
    fn list_note_paths(&self) -> anyhow::Result<Vec<std::path::PathBuf>>;
    fn read_note(&self, path: &std::path::Path) -> anyhow::Result<Note>;
}
```

### 5.2 FileSystemVault

Implementation:

* Configured with a root directory (`PathBuf`).
* `list_note_paths`:

  * Recursively list all `*.md` files under root.
* `read_note`:

  * Read file content into `Note`.
  * Derive `title` from filename (strip `.md`).
  * Derive `date` from filename if matches `YY-MM-DD` or `YYYY-MM-DD`.

### 5.3 VaultIndexManager

Coordinates vault + parser + index.

```rust
pub struct VaultIndexManager<V: Vault, I: IndexStore> {
    pub vault: V,
    pub index: I,
    // optionally: map of path → last_mtime/hash for incremental updates
}

impl<V: Vault, I: IndexStore> VaultIndexManager<V, I> {
    pub fn full_reindex(&mut self) -> anyhow::Result<()> {
        let paths = self.vault.list_note_paths()?;
        for path in paths {
            let note = self.vault.read_note(&path)?;
            let parsed = parse_note_markdown(note); // free function or method
            self.index.upsert_parsed_note(parsed)?;
        }
        Ok(())
    }

    pub fn reindex_note_path(&mut self, path: &std::path::Path) -> anyhow::Result<()> {
        let note = self.vault.read_note(path)?;
        let parsed = parse_note_markdown(note);
        self.index.upsert_parsed_note(parsed)
    }
}
```

---

## 6. Domain / Query API

On top of `VaultIndexManager` and `IndexStore`, provide a higher-level `Domain` object that exposes operations for the frontends.

```rust
pub struct Domain<V: Vault, I: IndexStore> {
    pub index_mgr: VaultIndexManager<V, I>,
}

impl<V: Vault, I: IndexStore> Domain<V, I> {
    /// Reindex entire vault (expensive). Typically called at startup or on explicit command.
    pub fn reindex_all(&mut self) -> anyhow::Result<()> {
        self.index_mgr.full_reindex()
    }

    /// Get or create a daily note for a specific date.
    pub fn open_daily(&mut self, date: chrono::NaiveDate) -> anyhow::Result<Note> {
        // Strategy: try to find a note with matching date; if absent, create new file and reindex it.
        unimplemented!()
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

    // Further helpers for weekly summaries, etc.:
    pub fn weekly_summary(&self, range: &DateRange) -> WeeklySummary {
        // Implementation detail left to the agent; include tasks created/closed in range, etc.
        unimplemented!()
    }
}
```

`WeeklySummary` can be defined as:

```rust
pub struct WeeklySummary {
    pub new_tasks: Vec<Task>,
    pub completed_tasks: Vec<Task>,
    pub notes: Vec<NoteMeta>,
    pub top_tags: Vec<(String, usize)>, // (tag, count)
}
```

---

## 7. JSON-RPC API Design

We’ll expose the `Domain` functionality via a JSON-RPC server.
The JSON-RPC server can be implemented as a standalone Rust binary that:

* Reads/write JSON-RPC messages on stdin/stdout (good for Neovim integration), or
* Listens on a local TCP/Unix socket (also fine).

For now, assume **stdio JSON-RPC** for easy Neovim integration.

### 7.1 General Structure

* Protocol: JSON-RPC 2.0
* Methods:

  * `core.reindex`
  * `core.list_tasks`
  * `core.task_detail`
  * `core.items_for_tag`
  * `core.notes_in_range`
  * `core.weekly_summary`
  * `core.open_daily`
  * `core.read_note` (get full markdown content)
  * `core.write_note` (update markdown on disk, then reindex that note)

### 7.2 Example Methods

#### `core.reindex`

* **Params**: `{}`
* **Result**: `{ "status": "ok" }`

```json
// Request
{"jsonrpc":"2.0","id":1,"method":"core.reindex","params":{}}

// Response
{"jsonrpc":"2.0","id":1,"result":{"status":"ok"}}
```

#### `core.list_tasks`

* **Params**:

```json
{
  "status": "open" | "in_progress" | "done" | "blocked" | null,
  "tags": ["projects/note-app", "people/jack"],
  "text_search": "linking",
  "touched_since": "2025-03-01" // optional
}
```

* **Result**: `Task[]` (serialized versions)

#### `core.task_detail`

* **Params**:

```json
{ "task_id": "T-2025-001" }
```

* **Result**:

```json
{
  "task": { /* Task */ },
  "mentions": [ /* TaskMention[] */ ]
}
```

#### `core.items_for_tag`

* **Params**:

```json
{ "tag": "projects/note-app" }
```

* **Result**:

```json
{
  "tag": "projects/note-app",
  "tasks": [ /* Task[] */ ],
  "log_entries": [ /* LogEntry[] */ ]
}
```

#### `core.notes_in_range`

* **Params**:

```json
{
  "start": "2025-03-10",
  "end": "2025-03-17"
}
```

* **Result**: `NoteMeta[]`

#### `core.open_daily`

* **Params**:

```json
{ "date": "2025-03-10" }
```

* **Result**: `Note` (with full `content`), creating the note file if it doesn’t exist.

#### `core.read_note`

* **Params**:

```json
{ "note_id": "<opaque-id>" }
```

* **Result**: `Note`

#### `core.write_note`

* **Params**:

```json
{
  "note_id": "<id>",
  "content": "full markdown content to write"
}
```

* Behavior:

  * Overwrite the note file on disk.
  * Reindex that note.
  * Return updated `Note` or `NoteMeta`.

---

## 8. Implementation Plan for the Agent

Suggested implementation steps:

1. **Project layout (Rust)**

   ```text
   note-core/
     Cargo.toml
     src/
       lib.rs           // exports Vault, parser, index, domain
       vault.rs
       parser.rs
       index.rs
       domain.rs

   note-daemon/
     Cargo.toml
     src/
       main.rs          // JSON-RPC server using note-core
   ```

2. **Implement Vault & Note model**

   * `FileSystemVault` with root directory configured (via env var or config file).

3. **Implement basic parser**

   * `parse_note_markdown(Note) -> ParsedNote` with:

     * Section detection (`## Tasks` / `## Log`).
     * Task-line parsing.
     * Log-entry parsing.
     * Tag and task-ID extraction.
     * Mentions generation.

4. **Implement InMemoryIndexStore**

   * `upsert_parsed_note`:

     * Remove old tasks/log entries/mentions for this note.
     * Insert new ones.
   * Basic filters in `list_tasks`, `items_for_tag`, etc.

5. **Implement VaultIndexManager**

   * `full_reindex` using `Vault` + `parse_note_markdown` + `IndexStore`.

6. **Implement Domain**

   * Wrap `VaultIndexManager` and provide `reindex_all`, `list_tasks`, `task_detail`, etc.
   * Stub `weekly_summary` if needed later.

7. **Implement JSON-RPC server**

   * Use a simple JSON-RPC library or manual handling:

     * Read line from stdin
     * Parse JSON
     * Dispatch to `Domain` methods
     * Write response JSON to stdout.

8. **Neovim integration (later)**

   * Write a Lua plugin that:

     * Starts the `note-daemon` process.
     * Sends JSON-RPC requests.
     * Renders results in:

       * Regular buffers (for notes)
       * Floating windows / Telescope pickers (for task lists, summaries, etc.).

