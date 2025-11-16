# Progress Log

## 2024-??
- Converted repo into a Cargo workspace with `note-core` (library) and `note-daemon` (JSON-RPC server) crates.
- Implemented core data models, vault abstraction, Markdown parser for Tasks/Log sections, in-memory index, and domain layer plumbing.
- Added a blocking stdio JSON-RPC loop that exposes the domain methods defined in `AGENTS.md`.
- `cargo fmt` applied; `cargo check` currently blocked because the sandbox cannot reach crates.io (network restricted).
- Introduced a pluggable parser interface (`NoteParser` trait with regex-backed default) so future backends (tree-sitter, etc.) can be swapped in without touching the domain/index layers.
- Parser now supports multiline task/log continuations, am/pm timestamps, indented headings, and includes unit tests covering these cases.
- Added a scratch Neovim module (`nvim/lua/note_rpc`) that spawns the daemon, exposes `:NoteListTasks` and `:NoteOpenDaily`, and renders simple floating buffers for manual testing (stdout/stderr buffering fixed, empty JSON params handled, tolerant of serde struct wrappers coming back from JSON).

### Follow-ups / Next Steps
1. Harden JSON-RPC server (better error handling, logging, config loading) and add automated tests/integration harness once dependency fetch succeeds.
2. Decide on canonical task ID generation utilities and integrate with parser/UI flows.
3. Expand Neovim frontend (buffer sync, task detail views) or replace with polished plugin once the core stabilizes; current Lua module is purposely minimal.
4. Investigate incremental vault watching + background reindexing (currently full reindex per note change).
