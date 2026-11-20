# aworkspace

Lightweight workspace manager for multi-repo development.

## The Problem

Working across multiple repositories on a single feature or initiative is cumbersome. You end up with repos scattered across directories, branches everywhere, and no clear organization or context for what you're working on.

`aworkspace` organizes code into **workspaces** — each a directory containing git worktrees from one or more repos, plus metadata that captures goals, constraints, and status. Each workspace is isolated, with its own branches, keeping your work focused and organized.

## How It Works

aworkspace uses **git worktrees** to create isolated working directories for each workspace. All repos are stored as bare clones in a central location, and each workspace gets its own worktrees, each detached at the repo's default branch.

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
      repo-a/                      # Worktree
      repo-b/                      # Worktree
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

Add a repository to the current workspace. Creates a bare clone (if needed) and a worktree detached at the repo's default branch.

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

Clone an existing workspace

### PLANNED: `aworkspace list -l`

Detailed list format (repo count, worktrees, etc).

### PLANNED: `aworkspace archive <workspace>`

Archive a workspace (safe by default): detaches its worktrees, keeps the record (`.aworkspace.toml`, `WORKSPACE.md`, notes), and moves it to `_archive/<name>@<date>/` (e.g. `_archive/my-feature@2026-06-30/`). Never deletes shared content — the bare repos stay intact for later rehydrate.

Requires the workspace name — it does not infer from the current directory. Refuses if your shell's cwd is at or under the target workspace (`cd` out first), so archiving never moves the directory out from under you.

**Options:**

- `--force` — skip confirmation.

A clean worktree archives with no prompt. A worktree with **uncommitted or untracked** changes — or **commits on no branch** — archives only after confirmation (or `--force`), since that work is reachable from no ref and exists nowhere else. Commits reachable from a branch, tag, or remote are preserved in the bare repo.

To actually delete an archived workspace, remove its folder directly (`rm -rf _archive/<name>@<date>`) — it's a plain folder. A tool-provided destroyer command is planned for a later release.

### PLANNED: `aworkspace prune`

Find and remove bare repos that aren't referenced by any workspace.

### PLANNED: `aworkspace doctor`

Check your environment for common issues (git config, stale worktrees, config validity).

## Configuration

Config file: `~/.config/aworkspace/config.toml`

```toml
workspaces_dir = "~/Workspaces"
bares_dir = "~/.local/share/aworkspace/repos"
workspace_worktree_subdir = "code"

[git]
path = "/usr/local/bin/git"  # custom git binary (default: "git")
```

**Options:**

- `workspaces_dir` — Where workspaces are created (default: `~/Workspaces`)
- `bares_dir` — Where bare clones are stored (default: `~/.local/share/aworkspace/repos`)
- `workspace_worktree_subdir` — Subdirectory within workspace for worktrees (default: `"code"`)

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

## Branch model

Worktrees start **detached** at each repo's default branch — aworkspace creates no branch. Git allows a branch to be checked out in only one worktree, and all workspaces share one bare repo per remote, so a starting point many workspaces can share must be detached.

Create a branch yourself when you're ready (`git switch -c <name>`). `archive` and `remove-repo` treat commits on no branch as unsafe — they're reachable from no ref and would be lost — so they require confirmation or `--force`.

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
