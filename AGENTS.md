# Note System Core – Technical Specification (Updated)

This document captures the evolving specification for the Note System Core. It has been updated based on architectural decisions and implementation progress.

## 0. Context & Goals

* Core will be implemented in **Go**.
* Frontends (Neovim plugin, TUI, others) communicate via **JSON-RPC**.
* The system operates over a **Markdown vault** (Obsidian-style folder of `.md` files).
* Markdown files remain the **source of truth**.
* The backend maintains an **in-memory index** (with optional SQLite backing later).
* Tasks are **first-class**, with global IDs, backlinks, and structured metadata.
* Logs and tags are parsed from Markdown.
* **Markdown parsing strategy is pluggable** (line-based regex v1; later Tree-sitter or Markdown AST).

## 1. Overview

The core is an editor-agnostic engine that:

1. Scans a Markdown vault.
2. Parses notes into structured objects (tasks, logs, mentions, tags).
3. Maintains an index for fast queries.
4. Exposes all functionality via JSON-RPC.
5. Allows multiple frontends (Neovim, TUI, CLI) to use the same data model.

## 2. Current Progress (Implementation Status)

Phase 1 scaffolding is in place. The repository now has a Go module, basic
project layout, placeholder daemon entrypoint, and the initial data models. All
subsequent phases still need to be implemented.

### ✅ Phase 1 – Core Scaffolding

* `go.mod` initialized (module `github.com/weakphish/yapper` targeting Go 1.24).
* Directory layout created (`cmd/note-daemon`, `internal/*` packages).
* Minimal daemon entrypoint at `cmd/note-daemon/main.go` that will later host the JSON-RPC server.
* Core data models (`Note`, `Task`, `LogEntry`, `TaskMention`, etc.) defined in `internal/model`.

### 🧭 Ready for Implementation

The following components remain to be implemented next (Phase 2 onward):

* **Vault layer** (filesystem scanning + note loading)
* **NoteParser interface** (pluggable markdown parsing strategy)
* **Regex-based Markdown parser (v1)**
* **IndexStore interface** (in-memory first)
* **VaultIndexManager** (vault + parse + index coordination)
* **Domain/query layer** (high-level operations)
* **JSON-RPC server** (exposing domain methods)

### 📌 Future Milestones (Not yet implemented)

* SQLite-backed index
* Tree-sitter / AST-based parser
* Neovim plugin (Lua)
* TUI frontend

## 3. Architecture Overview Architecture Overview Architecture Overview

**Layers:**

1. Vault (filesystem access)
2. Parser (pluggable markdown interpretation)
3. Index (in-memory initially)
4. Domain logic (queries, summaries)
5. JSON-RPC server

## 4. Pluggable Parser Interface

* A `NoteParser` interface abstracts markdown parsing.
* Initial implementation: regex-based, line-by-line parsing.
* Future implementations: Tree-sitter, Markdown AST (e.g. Goldmark AST), etc.
* Parser emits `ParsedNote` objects used by the index.

## 5. Data Model

* **Note**: id, path, title, date, content.
* **Task**: id, title, status, timestamps, tags, source note.
* **LogEntry**: id, note_id, line number, tags, task refs.
* **TaskMention**: backlink linking tasks to notes.
* **Index**: memory structures for fast lookup.

## 6. JSON-RPC API

Core implements JSON-RPC methods such as:

* `core.reindex`
* `core.open_daily`
* `core.list_tasks`
* `core.task_detail`
* `core.items_for_tag`
* `core.notes_in_range`
* `core.weekly_summary`
* `core.read_note`
* `core.write_note`

## 7. Frontend Model

* Neovim plugin communicates with Go backend via JSON-RPC.
* Plugin does **no semantic parsing**; backend is the source of truth.
* Neovim uses Tree-sitter only for UI (highlighting, motions).
* Other frontends (TUI, web UI) can consume backend structures.

## 8. Future Considerations

* SQLite index for faster startup and cross-tool access.
* Incremental parsing using Tree-sitter.
* Block-level editing operations.
* More advanced task workflow metadata.

## 9. Implementation Roadmap for Agents

This section is for **agents picking up the project with no prior memory**. Follow these steps in order unless explicitly instructed otherwise.

### Phase 1 – Core Scaffolding

1. **Create Go module & project layout**

   * Initialize a Go module (e.g. `github.com/user/note-core`).
   * Suggested structure:

     * `cmd/note-daemon/` – JSON-RPC server entrypoint.
     * `internal/vault/` – filesystem vault logic.
     * `internal/parser/` – NoteParser interface + implementations.
     * `internal/index/` – IndexStore + in-memory index.
     * `internal/domain/` – domain/query layer.
     * `internal/rpc/` – JSON-RPC handler wiring.
   * Add basic build instructions to a README.

2. **Define core data models** (in `internal/domain` or `internal/model`)

   * Implement Go structs for:

     * `Note`, `NoteID`
     * `Task`, `TaskID`, `TaskStatus`
     * `LogEntry`, `LogEntryID`
     * `TaskMention`
   * Ensure JSON tags exist so they can be serialized over JSON-RPC.

### Phase 2 – Vault + Parser (Regex v1)

3. **Implement Vault interface** (in `internal/vault`)

   * Interface to list note paths and load note contents from a root directory.
   * Implement `FileSystemVault` with a configurable root path.

4. **Define NoteParser interface** (in `internal/parser`)

   * Interface method: `Parse(note Note) (ParsedNote, error)`.
   * `ParsedNote` includes `Note`, `[]Task`, `[]LogEntry`, `[]TaskMention`.
   * Document that **parsing strategy is pluggable** (regex now, Tree-sitter/AST later).

5. **Implement RegexNoteParser (v1)**

   * Line-by-line parsing of Markdown using regex.
   * Handle `## Tasks` and `## Log` sections.
   * Extract tasks, log entries, tags (`#tag/subtag`), and task IDs (`[T-xxxx]`).
   * Ensure unit tests exist for typical and edge-case notes.

### Phase 3 – Index & Coordination

6. **Define IndexStore interface** (in `internal/index`)

   * Methods to upsert `ParsedNote`, remove notes, and run basic queries:

     * Get task by ID.
     * List tasks with filters.
     * Get log entries/mentions for a task.
     * List notes by date.
     * List tags and items for a tag.

7. **Implement InMemoryIndexStore**

   * Use Go maps to store:

     * Notes metadata.
     * Tasks by ID.
     * Log entries by ID.
     * Mentions by task ID.
     * Tag → tasks/log entries.
   * Implement `UpsertParsedNote` to replace previous data for a note.

8. **Implement VaultIndexManager**

   * Coordinates Vault + NoteParser + IndexStore.
   * Functions:

     * `FullReindex() error` – scans all notes, parses, indexes.
     * `ReindexNote(path string) error` – reparse a single note.

### Phase 4 – Domain Layer

9. **Implement Domain/query layer** (in `internal/domain`)

   * Wrap `VaultIndexManager` and expose high-level operations:

     * `ReindexAll()`
     * `OpenDaily(date)` – find or create a daily note.
     * `ListTasks(filter)`
     * `TaskDetail(taskID)` – returns task + mentions.
     * `ItemsForTag(tag)`
     * `NotesInRange(start, end)`
     * `WeeklySummary(range)` (can be minimal initially).
   * Domain layer should be **UI-agnostic** and unaware of Neovim.

### Phase 5 – JSON-RPC Server

10. **Implement JSON-RPC server** (in `cmd/note-daemon` + `internal/rpc`)

    * Choose a JSON-RPC 2.0 library or implement minimal support.
    * Wire methods:

      * `core.reindex`
      * `core.open_daily`
      * `core.list_tasks`
      * `core.task_detail`
      * `core.items_for_tag`
      * `core.notes_in_range`
      * `core.weekly_summary`
      * `core.read_note`
      * `core.write_note`
    * Ensure all structs used in responses are JSON-serializable.

### Phase 6 – Testing & Hardening

11. **Add tests**

    * Unit tests for parser, index, domain methods.
    * Integration test that:

      * Creates a small temporary vault.
      * Runs `FullReindex()`.
      * Asserts that tasks/log entries/tags are correctly indexed.

12. **Add configuration & ergonomics**

    * Command-line flags or config for vault path.
    * Simple logging for indexing and RPC requests.

### Phase 7 – Future Enhancements (Optional)

13. **SQLite-backed IndexStore**

    * Implement `IndexStore` using SQLite for persistence.

14. **Alternative NoteParser implementations**

    * Tree-sitter or Markdown AST-based parser.

15. **Frontend integrations**

    * Neovim plugin (Lua) that talks to this daemon over JSON-RPC.
    * Optional TUI using the same JSON-RPC or directly linking the Go packages.

Agents should consult this roadmap, pick the **next unfinished step**, and implement it according to the design above, updating this spec if the design evolves.
