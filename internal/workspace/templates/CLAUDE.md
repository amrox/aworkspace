This workspace was created by `aworkspace`. Read `WORKSPACE.md` for goals,
repos, constraints, and current status.

## Workspace layout

This directory contains git worktrees for one or more repos, all related to
a single initiative. Each worktree is a subdirectory which may have its own
`CLAUDE.md`. `workspace.toml` is the workspace configuration — do not modify it.

## Rules

- Use `WORKSPACE.md` as the source of truth for what this workspace is about.
- Only work within worktrees in this directory — do not modify files outside
  the workspace or in other workspaces.
- Repos in this workspace are independent projects — do not cross-reference
  between them. No repo should import, symlink, or refer to paths in another.
- These are git worktrees, not standalone clones. Do not run `git init`,
  `git clone`, or re-clone repos that are already here.
- Commit and push each repo independently.
- Do not clone, fetch, or push to repos not already checked out here.
- This workspace is temporary scaffolding for development — it will not be
  committed or shared as a unit.