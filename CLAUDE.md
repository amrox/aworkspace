# aworkspace

Go CLI for lightweight workspace management — creating and managing directories of git worktrees across multiple repos.

## Tech stack

- Go 1.26, Cobra (CLI framework), TOML config (go-toml/v2)
- Standard layout: `cmd/` for Cobra commands, `internal/workspace/` for core types and logic

## Development

- `go build` / `go test ./...`
- Toolchain managed via mise (`mise.toml`)

## Conventions

- Primary noun is **workspace** (not "project")
- Branch prefix default: `ws/`
- Config file: `~/.config/aworkspace/config.toml`
- Workspace marker: `workspace.toml`
- See `ROADMAP.md` for milestones and design decisions
