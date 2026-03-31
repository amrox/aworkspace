# aworkspace Roadmap

## Milestone 0.1 - Core Functionality

Goal: Get the basic workspace management working. Create, list, and manage workspaces with multiple repos.

### Must Have

- [ ] **Core data types** — `Workspace`, `Repo`, `Config` structs in `internal/workspace/`
- [ ] **Config loading** — Read/write `~/.config/aworkspace/config.toml` with defaults
- [ ] **Workspace discovery** — Find workspace by walking up from cwd to locate `workspace.toml`
- [ ] **`aworkspace new <name>`** — Create workspace directory with `workspace.toml` and `README.md`
- [ ] **`aworkspace list`** — List all workspaces (basic: name only)
- [ ] **`aworkspace show`** — Display current workspace info (repos, branches, status)
- [ ] **`aworkspace add-repo <url> [branch]`** — Clone bare repo + create worktree
  - Handle bare clone creation
  - Create worktree with branch
  - Update `workspace.toml` with repo metadata
  - Support branch naming with configurable prefix
  - **Support multiple different repos per workspace**
- [ ] **Basic tests** — Unit tests for workspace discovery, config loading, path handling

### Nice to Have

- [ ] **`aworkspace cd <name>`** — Output path for shell integration
- [ ] **`aworkspace rm [workspace]`** — Remove workspace (with safety checks)
- [ ] **Better error messages** — User-friendly errors with suggestions

### Out of Scope for 0.1

- Multiple worktrees of the same repo (same bare, different branches)
- Bookmarks
- `update`/`reset` commands
- `prune` command
- `doctor` command
- Submodule support
- Web dashboard

## Milestone 0.2 - Multi-Repo & Polish

- [ ] Shell completion (bash/zsh/fish)
- [ ] Submodule support (with git-worktree-tools integration)
- [ ] `aworkspace rm` with uncommitted change detection
- [ ] `aworkspace prune` for orphaned bare repos
- [ ] `aworkspace update` — fetch and rebase workspace branches
- [ ] `aworkspace reset` — reset workspace to clean state
- [ ] Bookmarks for common git hosts
- [ ] `--from` flag for cloning workspaces

## Milestone 0.3 - Quality of Life

- [ ] `aworkspace doctor` — environment checks
- [ ] Better git URL parsing (support all formats)
- [ ] Migration tool for POC workspaces
- [ ] Homebrew formula
- [ ] **Rich `list` output** — flexible formatting like `ls -l`
  - `-l, --long` — detailed format (status, repo count, branches, last modified)
  - `-1` — single column (name only, one per line)
  - `--format <template>` — custom output format
  - Sortable fields (name, created, modified, status)
  - Filter by status or other attributes

## Future

- Web dashboard for session management
- Agent orchestration integration
- Workspace templates
- Workspace sharing/export
