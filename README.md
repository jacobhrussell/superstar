# Superstar 

Superstar is a CLI for managing agent sessions.

To create a new agent session:

```bash
superstar session new --agent claude --dir ~/my/proj/src --name my-feature --prompt "create my feature"
```
- opens a new tmux session named `my-feature`
- creates a git worktree and branch named `my-feature`
- initializes the selected agent 
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
cd superstar
go install .
```

`go install` drops the binary in `$(go env GOPATH)/bin` (default `~/go/bin`). If `superstar` is not found after install, that directory is not on your `PATH`. Add it:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Persist by appending the line to your shell rc (`~/.zshrc`, `~/.bashrc`, etc.) and reloading.

Alternatively, build a binary in the current directory and run it directly or move it onto your `PATH`:

```bash
go build -o superstar .
./superstar session new ...
```

## Configuration

`superstar` optionally reads a YAML config from `~/.config/superstar/config.yaml`.

```yaml
# ~/.config/superstar/config.yaml
default_agent: claude
```

Supported keys:

- `default_agent` — fallback for `session new --agent` when the flag is omitted. Must be one of `claude`, `codex`, `cursor`, `opencode`.