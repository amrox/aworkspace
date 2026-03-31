# aworkspace

Lightweight workspace manager for multi-repo development.

## The Problem

Working across multiple repositories on a single feature or initiative is cumbersome. You end up with repos scattered across directories, branches everywhere, and no clear organization or context for what you're working on.

`aworkspace` organizes code into **workspaces** — each a directory containing git worktrees from one or more repos, plus metadata that captures goals, constraints, and status. Each workspace is isolated, with its own branches, keeping your work focused and organized.

## How It Works

aworkspace uses **git worktrees** to create isolated working directories for each workspace. All repos are stored as bare clones in a central location, and each workspace gets its own worktrees with dedicated branches.

```
~/Repos/                        # Bare clones (shared)
  repo-a.git
  repo-b.git

~/Workspaces/                   # Workspaces
  my-feature/
    workspace.toml              # Workspace config
    README.md                   # Goals, context, notes
    repos/
      repo-a/                   # Worktree (branch: my-feature)
      repo-b/                   # Worktree (branch: my-feature)
```

## Installation

```bash
go install github.com/amrox/aworkspace@latest
```

Or build from source:

```bash
git clone https://github.com/amrox/aworkspace
cd aworkspace
go build .
```

## Quick Start

```bash
# Create a new workspace
aworkspace new my-feature

# Add repos (creates bare clones and worktrees automatically)
cd ~/Workspaces/my-feature
aworkspace add-repo github.com/user/repo-a
aworkspace add-repo github.com/user/repo-b my-custom-branch

# List all workspaces
aworkspace list

# Show details about current workspace
aworkspace show

# Check your environment
aworkspace doctor
```

## Commands

### `aworkspace new <name>`

Create a new workspace. Creates the directory structure and initializes `workspace.toml` and `README.md`.

**Options:**
- `--from <workspace>` — Clone an existing workspace's repo list with fresh branches

### `aworkspace list`

List all workspaces with their status.

**Options:**
- `-l, --long` — Show detailed info (repo count, branches, dirty state)

### `aworkspace show`

Show details about a workspace (repos, branches, status). Infers workspace from current directory.

**Options:**
- `-p, --workspace <name>` — Specify workspace explicitly

### `aworkspace add-repo <url-or-bookmark> [branch]`

Add a repository to the current workspace. Creates a bare clone (if needed) and a worktree.

**Examples:**
```bash
# Full URL
aworkspace add-repo git@github.com:user/repo.git

# Using a bookmark (see Configuration)
aworkspace add-repo my-tool
aworkspace add-repo work:gitlab-runners
```

**Options:**
- `-p, --workspace <name>` — Add to a specific workspace

### `aworkspace cd <workspace>` (alias: `switch`)

Change to a workspace directory. Outputs the path for shell integration.

To enable `cd` functionality, add to your shell profile:

```bash
# For bash/zsh
eval "$(aworkspace cd --setup)"

# Or manually:
aw() { cd "$(aworkspace cd "$@")"; }
```

Then use:
```bash
aw my-feature    # cd to ~/Workspaces/my-feature
```

### `aworkspace rm [workspace]`

Remove a workspace. Removes worktrees and deletes the workspace directory. Warns if there are uncommitted changes or unpushed branches.

If no workspace is specified, uses the current directory.

### `aworkspace prune`

Find and remove bare repos that aren't referenced by any workspace. Shows what will be deleted and prompts for confirmation.

### `aworkspace update`

Update all repos in the current workspace. For each worktree:
1. Fetches latest from origin
2. If the workspace branch is clean, rebases onto the configured base branch (default: `main`)
3. Skips repos with uncommitted changes (reports them)

Safe by default — never discards uncommitted work.

**Options:**
- `--no-rebase` — Fetch only, don't rebase

### `aworkspace reset`

Reset the current workspace to a clean state. For each worktree:
1. Checks for uncommitted changes
2. If clean: switches to workspace branch and fetches/resets to match remote
3. If dirty: skips and reports (use `--force` to discard changes)

Use this to get back to a known-good state, especially after experimental changes.

**Options:**
- `--force` — Discard uncommitted changes without prompting (destructive)

### `aworkspace doctor`

Check your environment for common issues:
- Git version (needs >= 2.48 for relative worktree paths)
- `worktree.useRelativePaths` config
- Stale worktrees

## Configuration

Config file: `~/.config/aworkspace/config.toml`

```toml
workspaces_dir = "~/Workspaces"
repos_dir = "~/Repos"
branch_prefix = ""
init_submodules = false
```

**Options:**
- `workspaces_dir` — Where workspaces are created (default: `~/Workspaces`)
- `repos_dir` — Where bare clones are stored (default: `~/Repos`)
- `branch_prefix` — Optional prefix for auto-generated branches
- `init_submodules` — Whether to initialize submodules in worktrees (default: `false`)

### Bookmarks

Bookmarks file: `~/.config/aworkspace/bookmarks.toml`

Define shortcuts for common git hosts and organizations:

```toml
[default]
host = "github.com"
user = "amrox"

[work]
host = "gitlab.company.com"
user = "team-infra"
```

Then use them:

```bash
aworkspace add-repo my-tool              # -> git@github.com:amrox/my-tool.git
aworkspace add-repo work:gitlab-runners  # -> git@gitlab.company.com:team-infra/gitlab-runners.git
```

## Branch Naming

By default, aworkspace creates branches named after the workspace (e.g., workspace `my-feature` → branch `my-feature`). You can configure a prefix:

```toml
# config.toml
branch_prefix = "ws/"
```

Then workspace `my-feature` creates branches `ws/my-feature`.

This is useful for:
- **Organization** — keeping workspace branches separate from other branch types
- **Tooling integration** — CI/scripts can identify workspace branches by prefix
- **Team conventions** — matching existing naming schemes

When adding a repo, you can override the branch name:

```bash
aworkspace add-repo my-repo custom-branch-name
```

## Workspace Structure

Each workspace contains:

**`workspace.toml`** — Structured config
```toml
name = "my-feature"
created = "2026-03-31"
status = "active"

[[repos]]
name = "repo-a"
url = "git@github.com:user/repo-a.git"
branch = "my-feature"
bare = "~/Repos/repo-a.git"

[[repos]]
name = "repo-b"
url = "git@github.com:user/repo-b.git"
branch = "my-feature"
bare = "~/Repos/repo-b.git"
```

**`README.md`** — Human-readable context, goals, notes

**`repos/`** — Directory containing all worktrees

## Benefits

- **Organized multi-repo work** — All repos for a feature in one place
- **Isolated branches** — Each workspace gets its own branches, no cross-contamination
- **Context capture** — `README.md` documents what you're doing and why
- **Efficient disk usage** — Bare repos are shared, worktrees are lightweight
- **Works well with AI coding agents** — Focused scope and context files help tools understand boundaries

## Git Worktree Notes

### Relative Paths for Devcontainers

Git worktrees use absolute paths by default, which breaks inside devcontainers. Git 2.48+ supports relative paths:

```bash
git config --global worktree.useRelativePaths true
```

`aworkspace doctor` checks for this.

### Submodules

Worktrees don't initialize submodules by default. Set `init_submodules = true` in your config to auto-initialize them when creating worktrees.

## License

MIT
