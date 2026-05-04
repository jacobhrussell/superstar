# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build / test / run

- Build: `go build -o superstar .` (or `go install .`)
- Run: `./superstar <subcommand>` — e.g. `./superstar session new`, `./superstar project list`, `./superstar config init`
- Test all: `go test ./...`
- Test one package: `go test ./cmd`
- Single test: `go test ./cmd -run TestSessionNewPreRunResolvesAgent`
- Verbose: `go test ./cmd -v -run <Name>`

`tmux` must be on `PATH` for any session command that actually starts a session; the unit tests do not require it.

## Architecture

Single-binary Cobra CLI. `main.go` only calls `cmd.Execute()`; everything else lives in package `cmd/` as a flat package with one file per command. Commands self-register through `init()` blocks that attach to one of the parent commands: `rootCmd` (`root.go`), `sessionCmd` (`session.go`), `projectCmd` (`project.go`), or `configCmd` (`config.go`).

Three concepts to keep straight:

1. **Config file** (`~/.config/superstar/config.yaml`) — the source of truth. Schema is `Config{ DefaultAgent, Projects map, Sessions map }` in `cmd/config_io.go`. Always read with `loadConfig()` and write with `saveConfig()`; both handle the missing-file case and ensure the maps are non-nil. Viper is initialized in `root.go` only so `viper.GetString("default_agent")` works as a fallback in `session new`; do not use Viper for the projects/sessions maps.

2. **Projects** — named `(dir, agent)` presets. Used as defaults when creating a session via `--project`. CRUD lives in `cmd/project_*.go`.

3. **Sessions** — a session is a tmux session + (optionally) a git worktree + an agent process. The full lifecycle is in `cmd/session_new.go`:
   - tmux session name is always suffixed: `<name>-superstar` (see `tmuxName` / `sessionNameFromTmux` in `cmd/tmux.go`). Use these helpers — never concatenate the suffix manually.
   - If the target dir is a git repo, a worktree is created at `<dir>-<name>` on a new branch `<name>`. On any later failure, `defer` cleans up worktree → branch → tmux session in reverse order. Preserve this rollback chain when modifying the create flow.
   - The agent is launched by `tmux send-keys` of the agent name as a command. Initial prompt (if given) is sent after a 2s sleep so the agent's TUI has time to render, then a 300ms pause before pressing Enter. These delays are load-bearing — the agent TUIs swallow input typed too early.
   - Valid agents are hardcoded in `cmd/agents.go`: `claude`, `codex`, `cursor`, `opencode`. `validateAgent` is the single gate.

## Conventions worth knowing

- **Interactive vs. non-interactive**: commands check `isInteractiveTerm()` (TTY on stdin+stdout) to decide whether to prompt with `huh` forms or hard-fail on missing flags. When adding a flag, support both paths.
- **TUI lists** (`session list`, `project list`) are Bubble Tea models with a `bubbles/list` and a selected-item field set on Enter. `session list` filters out config entries that have no live tmux session — stale entries can be cleaned with `session delete`.
- **Cobra flag state is package-global** (e.g. `sessionNewDir`, `projectEditAgent`). Tests must reset this state between cases — see `resetSessionNewState()` in `session_new_test.go` and use it as the pattern for new tests. Tests that touch the config file use `t.Setenv("HOME", t.TempDir())` to sandbox it.
- **Editing config**: prefer mutating the loaded struct then calling `saveConfig` over rewriting YAML by hand. The marshaling round-trip is the contract.
