# aworkspace Roadmap

## Design Notes

**Branch prefix:** Default is `"ws/"` (e.g., workspace `my-feature` creates branches `ws/my-feature`). This prefix:
- Clearly marks workspace-managed branches as plumbing (not real topic branches)
- Avoids locking the repo's default branch across multiple workspaces
- Enables clean lifecycle operations (`rm` can safely delete `ws/*` branches)
- Is configurable via `branch_prefix` in config

**Branch collision handling:** If a `ws/<name>` branch already exists or is already checked out, `add-repo` will fail with a clear error. No auto-suffixing or magic renaming. This should be rare since `ws/` is aworkspace's namespace. If it happens, it usually means a workspace was removed without cleanup (which `doctor` would flag).

**Open question (0.1 decision required):** Workspace context file naming. Options:
- `WORKSPACE.md` — Avoids collision with repo READMEs, clear purpose. Current leaning.
- `README.md` — More conventional, but conflicts when repos have their own READMEs in `code/`
- Workspace-level `README.md` at root + repo READMEs in `code/*/` — Could work but feels redundant

Decision needed before implementing `new` command. Current direction: `WORKSPACE.md` for clarity and avoiding collisions.

**Agent context:** Workspaces should include a default `CLAUDE.md` that explains workspace structure to agents. Key rules that need to be communicated:
- Repos in the workspace are independent projects (don't cross-reference between them)
- Each repo will be committed separately
- The workspace is temporary scaffolding for development

This needs to be configurable:
- Default template for `aworkspace new` should include a sensible `CLAUDE.md`
- Config option to disable auto-creation of `CLAUDE.md`
- Support for custom workspace templates (user-defined scaffolding)

Without this, every agent session requires manual explanation of workspace isolation rules.

**Checkout roles (edit vs reference):** Each repo checkout in a workspace could be tagged as either "edit" (actively being developed) or "reference" (read-only context for agents). For example, a workspace might have `my-service` as the edit target while `shared-lib` and `api-spec` are references. This distinction is purely advisory — it helps agents understand which code they should be modifying vs. just reading for context. Could be a field in `workspace.toml` per-repo entry.

## Milestone 0.1 - Core Functionality

Goal: Get the basic workspace management working. Create, list, and manage workspaces with multiple repos.

### Must Have

- [ ] **Core data types** — `Workspace`, `Repo`, `Config` structs in `internal/workspace/`
- [ ] **Config loading** — Read/write `~/.config/aworkspace/config.toml` with defaults
- [ ] **Workspace discovery** — Find workspace by walking up from cwd to locate `workspace.toml`
- [ ] **`aworkspace new <name>`** — Create workspace directory with `workspace.toml`, `WORKSPACE.md`, and `CLAUDE.md`
  - `CLAUDE.md` includes default workspace isolation rules for agents
  - Makes agents immediately understand that repos are independent projects
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
- [ ] **Configurable `CLAUDE.md` generation** — Config option to disable auto-creation of `CLAUDE.md`
- [ ] **Workspace templates** — User-defined templates for `new` command (scaffolding, standard files, custom CLAUDE.md)

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
- [ ] **Optional workspace subtitle** — One-line description field in `workspace.toml` (e.g., `subtitle = "Q2 nav rewrite"`), shown in `list -l`

## Future

- Web dashboard for session management
- Agent orchestration integration
- Workspace templates
- Workspace sharing/export
