# aworkspace

Lightweight workspace manager for multi-repo development.

## The Problem

Working across multiple repositories on a single feature or initiative is cumbersome. You end up with repos scattered across directories, branches everywhere, and no clear organization or context for what you're working on.

`aworkspace` organizes code into **workspaces** — each a directory containing git worktrees from one or more repos, plus metadata that captures goals, constraints, and status. Each workspace is isolated, with its own branches, keeping your work focused and organized.

## How It Works

aworkspace uses **git worktrees** to create isolated working directories for each workspace. All repos are stored as bare clones in a central location, and each workspace gets its own worktrees with dedicated branches.

```
~/.local/share/aworkspace/repos/   # Bare clones (shared)
  github.com/user/repo-a/
  github.com/user/repo-b/

~/Workspaces/                      # Workspaces
  my-feature/
    .aworkspace.toml               # Workspace config
    WORKSPACE.md                   # Goals, context, notes
    CLAUDE.md                      # Agent instructions
    code/
      repo-a/                      # Worktree (branch: my-feature)
      repo-b/                      # Worktree (branch: my-feature)
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
aworkspace add-repo git@github.com:user/repo-a.git
aworkspace add-repo git@github.com:user/repo-b.git

# List all workspaces
aworkspace list

# Show details about current workspace
aworkspace show
```

## Global Options

- `-v, --verbose` — Show detailed output including git subprocess output

## Commands

### `aworkspace new <name>`

Create a new workspace. Creates the directory structure and initializes `.aworkspace.toml`, `WORKSPACE.md`, and `CLAUDE.md`.

### `aworkspace list`

List all workspaces (one name per line).

### `aworkspace show`

Show details about a workspace (repos and their URLs). Infers workspace from current directory.

**Options:**

- `-C, --dir <path>` — Run as if started in this directory

### `aworkspace add-repo <url>`

Add a repository to the current workspace. Creates a bare clone (if needed) and a worktree on a branch named after the workspace (e.g. workspace `my-feature` → branch `my-feature`). Prints the branch name and whether it was created or reused.

**Examples:**
```bash
aworkspace add-repo git@github.com:user/repo.git
aworkspace add-repo https://github.com/user/repo.git
```

**Options:**

- `-C, --dir <path>` — Run as if started in this directory

### `aworkspace init <shell>`

Output shell integration code. Add to your shell profile to enable the `cd` subcommand:

```bash
# For bash/zsh
eval "$(aworkspace init zsh)"
```

This creates a shell wrapper so that `aworkspace cd` actually changes your directory.

### `aworkspace cd <workspace>` (alias: `switch`)

Change to a workspace directory. Requires shell integration via `aworkspace init`.

```bash
aworkspace cd my-feature    # cd to ~/Workspaces/my-feature
```

### PLANNED: `aworkspace new --from <workspace>`

Clone an existing workspace's repo list with fresh branches.

### PLANNED: `aworkspace list -l`

Detailed list format (repo count, branches, dirty state).

### PLANNED: `aworkspace rm [workspace]`

Remove a workspace. Removes worktrees and deletes the workspace directory. Warns if there are uncommitted changes or unpushed branches.

### PLANNED: `aworkspace prune`

Find and remove bare repos that aren't referenced by any workspace.

### PLANNED: `aworkspace update`

Fetch and rebase workspace branches onto their base branch. Skips repos with uncommitted changes.

### PLANNED: `aworkspace reset`

Reset workspace to a clean state. Skips dirty repos unless `--force` is used.

### PLANNED: `aworkspace doctor`

Check your environment for common issues (git config, stale worktrees, config validity).

## Configuration

Config file: `~/.config/aworkspace/config.toml`

```toml
workspaces_dir = "~/Workspaces"
bares_dir = "~/.local/share/aworkspace/repos"
workspace_worktree_subdir = "code"
branch_prefix = ""

[git]
path = "/usr/local/bin/git"  # custom git binary (default: "git")
```

**Options:**

- `workspaces_dir` — Where workspaces are created (default: `~/Workspaces`)
- `bares_dir` — Where bare clones are stored (default: `~/.local/share/aworkspace/repos`)
- `workspace_worktree_subdir` — Subdirectory within workspace for worktrees (default: `"code"`)
- `branch_prefix` — Optional prefix for workspace branches (default: `""` — branch name = workspace name)

### PLANNED: Bookmarks

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

By default, the branch is **named after the workspace** — the same name in every repo of the workspace, with no prefix. Workspace `my-feature` → branch `my-feature`.

The branch name describes *the work*, not the plumbing. Because the local branch name matches the remote name, there's no translation when you push or open a PR — `git push` just works, and the base branch (what you forked from) is recorded as metadata, not baked into the name.

**Creating or reusing a branch** (aworkspace always tells you which):
- Branch doesn't exist → created from the base branch.
- Branch exists but isn't checked out elsewhere → reused (the worktree attaches to it).
- Branch exists and is already checked out in another worktree → **error.** This is an exceptional case (workspace names are unique and teardown is clean, so it usually means a stale worktree left behind). aworkspace won't silently invent a different branch name; it tells you which worktree holds the branch. Pass an explicit branch (`add-repo <url> <branch>`) or clean up with `doctor`/`prune`.

```
workspace 'fix-auth'
  api     → branch fix-auth (new, from main)
  shared  → branch fix-auth (reusing existing)
```

**Optional prefix.** If you want workspace branches namespaced (e.g. to identify them in CI or match a team convention), set a prefix:

```toml
# config.toml
branch_prefix = ""        # default: branch name = workspace name
# branch_prefix = "ws/"   # namespaced: my-feature → ws/my-feature
# branch_prefix = "aw-"   # flat prefix: my-feature → aw-my-feature
```

Note that any prefix reintroduces a local/remote name difference, so leaving it empty is recommended unless you specifically need the namespace.

## Workspace Structure

Each workspace contains:

**`.aworkspace.toml`** — Structured config
```toml
[repos]
repo-a = { url = "git@github.com:user/repo-a.git" }
repo-b = { url = "git@github.com:user/repo-b.git" }
```

**`WORKSPACE.md`** — Human-readable context, goals, notes (avoids collision with repo READMEs)

**`CLAUDE.md`** — Agent instructions explaining workspace isolation rules

**`code/`** — Directory containing all worktrees

## Benefits

- **Organized multi-repo work** — All repos for a feature in one place
- **Isolated branches** — Each workspace gets its own branches, no cross-contamination
- **Context capture** — `WORKSPACE.md` documents what you're doing and why
- **Agent-friendly** — `CLAUDE.md` automatically explains workspace isolation rules to AI agents
- **Efficient disk usage** — Bare repos are shared, worktrees are lightweight

## Git Worktree Notes

### Relative Paths for Devcontainers

**Note:** This is a git configuration, not an aworkspace feature.

Git worktrees use absolute paths by default, which breaks inside devcontainers (the paths reference the host filesystem). Git 2.48+ supports relative worktree paths:

```bash
git config --global worktree.useRelativePaths true
```

With this setting, git creates worktrees with relative paths that work correctly when the workspace is bind-mounted into a container. This is entirely managed by git — aworkspace just creates worktrees using `git worktree add`, and git handles the path format based on your config.

### PLANNED: Submodules

Worktrees don't initialize submodules by default. A future `init_submodules` config option will auto-initialize them when creating worktrees.

## License

MIT
