# Superstar 

## Getting Started

### From Source

```bash
git clone https://github.com/jacobhrussell/superstar.git
go install .
```

## Configuration

`superstar` optionally reads a YAML config from `~/.config/superstar/config.yaml`. If the file doesn't exist, defaults apply.

```yaml
# ~/.config/superstar/config.yaml
default_agent: claude
```

Supported keys:

- `default_agent` — fallback for `session new --agent` when the flag is omitted. Must be one of `claude`, `codex`, `cursor`, `opencode`.