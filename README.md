# Yapper
## Yet Another Powerful Productivity Engine, Redux

## What?

- A task management/note-taking core focused on Markdown vaults, JSON-RPC, and Neovim integration.
- Pluggable back-end with core types and model, bring-your-own front-end
- `note-core`: Rust library that models notes, tasks, tags, and log entries from a vault of `.md` files.
- `note-daemon`: JSON-RPC server exposing the domain via stdio for frontends (Neovim plugin is the first target).

## Repository Layout

- `Cargo.toml` – workspace definition.
- `crates/note-core` – vault + parser + index + domain library.
- `crates/note-daemon` – stdio JSON-RPC binary that wires the domain to clients.
- `AGENTS.md` – detailed product spec / system design brief.
- `PROGRESS.md` – running log for future contributors/agents.

## Getting Started

```
NOTE_VAULT_PATH=/path/to/vault cargo run -p note-daemon
```

The daemon watches stdin for JSON-RPC 2.0 requests and writes responses to stdout. Suggested flow for development:

1. Set `NOTE_VAULT_PATH` to a test Markdown vault (contains daily notes with `## Tasks` and `## Log`).
2. Run `cargo run -p note-daemon`.
3. Send JSON-RPC requests (e.g. via `nvim` plugin, `jq`, or ad-hoc scripts).

## Scratch Neovim Module

A throwaway Neovim client lives under `nvim/lua/note_rpc`. To experiment:

1. Add the repo’s `nvim` dir to your `runtimepath`, e.g. in `init.lua`:
   ```lua
   vim.opt.rtp:append("/home/jack/Developer/yapper/nvim")
   require("note_rpc").setup({
     vault_path = "/path/to/vault", -- optional, defaults to cwd
     -- cmd = { "/absolute/path/to/note-daemon" }, -- optional override
   })
   ```
2. Inside Neovim, use:
   - `:NoteDaemonStart [vault_path]` / `:NoteDaemonStop` to control the daemon.
   - `:NoteListTasks` to show a floating task list.
   - `:NoteOpenDaily [YYYY-MM-DD]` to fetch or create a daily note in a new buffer (defaults to today).

This module is intentionally minimal—just enough to drive the JSON-RPC server during development.

## Current Status

- Vault traversal, note parsing (including multiline entries and indented headings), in-memory indexing, JSON-RPC endpoints, and a scratch Neovim frontend exist.
- `Domain::open_daily` creates YYYY-MM-DD notes with a simple template if missing.
- `cargo check/test` will pass once crates.io is reachable (sandbox currently offline).

## Future Wishlist

- [ ] Task ID generation & persistence helpers.
- [ ] Incremental vault watching + better JSON-RPC logging/configuration.
- [ ] Canonical parser improvements (tree-sitter backend, richer log parsing).
- [ ] Polished frontend (full Neovim plugin, TUI).
