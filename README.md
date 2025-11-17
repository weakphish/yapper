# Yapper
## Yet Another Powerful Productivty Engine, Redux
Core of a note-taking/task management system that I wish existed already.

## Key Features

- Core service that parses & indexes Markdown notes
    - Pluggable parsing and indexing layers allow for modularity
- Exposed indexed vault over JSON-RPC
    - Allows for N-many front-ends to exist, starting with a Neovim plugin

## Current Progress

Phase 1 (scaffolding) and the foundational pieces of Phases 2–3 are complete:

- Filesystem vault layer with helpers to list/load Markdown notes.
- Core parser contracts (`NoteParser`, `ParsedNote`) for pluggable strategies.
- In-memory index implementation plus a `VaultIndexManager` that coordinates the vault, parser, and index.

The next major milestone is implementing the regex-based parser and wiring the domain/query layer on top of the index manager.

## Project Layout

- `cmd/note-daemon/` – entrypoint for the JSON-RPC daemon.
- `internal/model/` – data models for notes, tasks, log entries, and mentions.
- `internal/vault/` – filesystem access layer.
- `internal/parser/` – pluggable Markdown parsing strategies.
- `internal/index/` – index interfaces, in-memory store, and vault/index coordination.
- `internal/domain/` – domain/query logic.
- `internal/rpc/` – JSON-RPC wiring.

## Building & Running

The current scaffold builds a minimal daemon placeholder that simply logs that
it has started. Future phases will extend it with the vault, parser, index, and
RPC implementations.

```sh
# build a binary
go build ./cmd/note-daemon

# or run it directly
go run ./cmd/note-daemon

# run tests (currently exercises the index + manager layers)
go test ./...
```
