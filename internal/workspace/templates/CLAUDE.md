This workspace was created by `aworkspace`. Read `WORKSPACE.md` for goals,
repos, constraints, and current status.

## Workspace layout

This directory contains git worktrees for one or more repos, all related to
a single initiative. Each worktree is a subdirectory which may have its own
`CLAUDE.md`. `.aworkspace.toml` is the workspace configuration — do not modify it.

## Rules

- Use `WORKSPACE.md` as the source of truth for what this workspace is about.
- Only work within worktrees in this directory — do not modify files outside
  the workspace or in other workspaces.
- Repos in this workspace are separate checkouts of independent projects. Do
  not link them on disk: no symlinks, imports, or relative paths from one
  worktree into another. Referring to another repo by name or URL in code and
  docs is fine.
- These are git worktrees, not standalone clones. Do not run `git init`,
  `git clone`, or re-clone repos that are already here.
- Commit and push each repo independently.
- Do not clone, fetch, or push to repos not already checked out here.
- This workspace is temporary scaffolding for development — it will not be
  committed or shared as a unit.

## Branches and worktrees

Each repo's **primary worktree** is `code/<repo>` and starts **detached** at the
repo's default branch — it is on no branch. Before committing work you want to
keep, create a branch (`git -C code/<repo> switch -c <branch>`). Commits made on
no branch can be lost when the workspace is archived.

For isolated or parallel work you may add an **extra worktree** of a repo:

- Put it **beside** the primary, never nested inside it:
  `git -C code/<repo> worktree add -b <branch> ../<repo>--<task>`
- Always create it on a branch (the `-b` above), so its commits survive removal.
- Never remove a repo's primary worktree `code/<repo>`.
- When done, commit, then remove the extra:
  `git -C code/<repo> worktree remove ../<repo>--<task>`

Do not create worktrees outside this workspace directory.