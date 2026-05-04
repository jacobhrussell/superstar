# Superstar 

Superstar is a CLI for managing agent sessions.

To create a new agent session:

```bash
superstar session new --agent claude --dir ~/my/proj/src --name my-feature --prompt "create my feature"
```
- opens a new tmux session named `my-feature`
- creates a git worktree and branch named `my-feature`
- initializes the selected agnet
- sends the initial prompt

To clean it up:

```bash
superstar session delete my-feature
```
- kills the tmux session
- removes the worktree
- deletes the branch

## Getting Started

### From Source

```bash
git clone https://github.com/jacobhrussell/superstar.git
go install .
```

## Configuration

`superstar` optionally reads a YAML config from `~/.config/superstar/config.yaml`.

```yaml
# ~/.config/superstar/config.yaml
default_agent: claude
```

Supported keys:

- `default_agent` — fallback for `session new --agent` when the flag is omitted. Must be one of `claude`, `codex`, `cursor`, `opencode`.