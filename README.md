# Note System Core

This repository hosts the backend for a Markdown-oriented task and note system.
The requirements and phase-by-phase roadmap live in `AGENTS.md`. Agents should
consult that document for the most current plan before making changes.

## Project Layout

- `cmd/note-daemon/` – entrypoint for the JSON-RPC daemon (placeholder for now).
- `internal/model/` – data models for notes, tasks, log entries, and mentions.
- `internal/vault/` – filesystem access layer (to be implemented in Phase 2).
- `internal/parser/` – pluggable Markdown parsing strategies.
- `internal/index/` – in-memory/persistent indexing layer.
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
```

## Development Process

1. Review `AGENTS.md` to determine the next unfinished roadmap item.
2. Implement the described functionality, keeping Markdown files as the source
   of truth.
3. Prefer small, well-tested packages inside `internal/` to keep the public API
   limited to the daemon binary.
4. Update documentation and tests as you progress through the phases.
