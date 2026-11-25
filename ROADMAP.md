# aworkspace Roadmap

## Design Notes

**Branch model (decided): aworkspace does not manage branches.** Every repo starts with **no branch**, checked out at the remote's default branch. You (or an agent) create branches when work starts. The tool never creates, names, reuses or deletes a branch.

**Why: a workspace is a place, not a task.** Real usage is one semi-permanent default workspace for small tasks, plus long-lived workspaces at the epic/initiative level. Both ship many branches over their life, so a "one workspace = one branch" rule fits neither. Quick throwaway tasks are out of scope: they go in the default workspace or in an agent's own worktree.

- **Workspace = context:** the repo set, roles, notes and record. Long-lived by design.
- **Branch = a task inside it:** named for the work, owned by whoever created it.

**Mechanics:**

- `add-repo <url>` runs `git worktree add --detach <dest> origin/<default>`, where `<default>` is the remote's HEAD. There is no branch argument and no `--base` flag.
- Start work with `git checkout -b <name>` in the worktree. It works from the no-branch state and carries along uncommitted changes and any commits already made there.
- `remove-repo` and `archive` keep every branch in the bare. Branch cleanup is the user's (`git branch -d`).

**Why no-branch is forced, not a taste choice:** git allows a branch to be checked out in only one worktree, and all workspaces share one bare repo per remote. Any starting point that many workspaces can share must therefore be detached. (Checking out `origin/main` directly is also a detached HEAD.)

**What this removes:** the workspace-named branch, `branch_prefix`, collision handling, branch ownership/markers, `--base`/`--branch` flags on `add-repo`, and the stored `base_branch`.

**What it requires instead:**

- **A record of where each repo sits** — branch name (if any) plus SHA, read from live HEAD when the record is written (archive, sync). Rehydrate and `new --from` rebuild from this. It is recorded state, not a creation-time setting.
- **An orphan-commit check.** Commits made before branching are reachable from no ref and are lost when the worktree is detached. `archive` and `remove-repo` treat them as dirty (confirm / `--force`). Detection: `git rev-list HEAD --not --branches --remotes --tags` non-empty.
- **`status`/`show` read live HEAD** for every repo; nothing trusts a stored branch name.

**Visibility carries the model.** Many git users are uneasy with "detached HEAD", so the output says "no branch" and always names the next step:

```
workspace 'nav-rewrite'
  api     → at origin/main (no branch). Start work: git checkout -b <name>
  shared  → at origin/main (no branch). Start work: git checkout -b <name>
```

`status` flags the one risky case, e.g. `api: 2 commits not on any branch`. The README states the contract once: repos start at the default branch with no branch of their own; create one when you start work.

**Extra worktrees: short-lived work inside a long-lived workspace (decided — convention, no command).** The workspace is medium-to-permanent; branches and extra worktrees inside it can be as short-lived as you like. No hard rule, same as git branches. A repo's **primary worktree** (`code/<repo>`, the one `add-repo` made) is the durable home base; you or an agent may add **extra worktrees** of the same repo for isolated tasks. This is ordinary git: `git worktree add` run from inside a worktree registers with the same shared bare. It also keeps agents inside the workspace instead of scattering worktrees elsewhere.

Conventions (recommended, not enforced):

- **Location: beside the primary, never nested inside it** — `code/<repo>--<task>`, e.g. `git -C code/api worktree add -b fix-auth ../api--fix-auth`. The whole worktree subdir is dropped on archive, so an extra beside the primary needs no new bookkeeping, and relative paths to sibling repos (`../shared`) keep working. Nesting inside the primary (some agent tools default to this, e.g. `<repo>/.claude/worktrees/`) breaks sibling paths and lets `remove-repo` take the nested worktree with it.
- **The name is never parsed.** Git knows which bare each worktree belongs to; the suffix is for humans (same rule as archive folder names). The separator is therefore free to change — `--` is the working choice, `+` the alternative under trial.
- **Not recorded.** Only the primary is in `.aworkspace.toml`. Extras are read live from git/the filesystem, the same way branches are. Rehydrate restores primaries only.
- **Always on a branch.** An extra worktree gets a branch at creation, so its commits survive removal.
- **Creator cleans up:** commit, then `git worktree remove <path>`.
- **Agents are told.** The generated `CLAUDE.md` carries this guidance (location, branch, cleanup), agent-agnostic:

  ```markdown
  ## Worktrees in this workspace

  - Each repo's primary worktree is `code/<repo>`. Do not remove it.
  - For isolated work, create an extra worktree beside it, never inside it:
    `git -C code/<repo> worktree add -b <branch> ../<repo>--<task>`
  - Always create it on a branch, so commits survive removal.
  - When done, commit, then remove it: `git worktree remove code/<repo>--<task>`
  ```

The convention cannot be relied on — an agent's built-in worktree feature may use its own location — so safety comes from ownership, not placement: archive discovers owned worktrees by asking each bare what worktrees it owns (see "archive: dirty detection"), so an extra on our bare is handled wherever it sits. A misplaced worktree is untidy, never dangerous.

- **`remove-repo` with extras present:** removes the primary only, leaves the extras, and prints a warning listing them (the archive scan still covers them later).
- **Later, if missed:** `show` lists extras it finds; a command to create them.
- (Alternative layout considered: a separate `worktrees/<repo>/<task>` dir. Tidier `code/`, but needs a second ignore/archive rule and breaks `../<sibling>` paths. Revisit only if `code/` gets cluttered in practice.)

**Deferred conveniences — add only when missed:**

- `add-repo --at <ref>` to start at a tag or another branch (today: move the worktree by hand). Watch for repeated hand-pinning of reference repos to release tags.
- A start-a-task helper that creates the same branch name across chosen repos.

**`update` (0.2), open:** for a repo with no branch and no local commits, re-detach at the freshly fetched default branch. Repos on a branch are the user's to rebase; the tool may fetch only.

**Healing old bares (decided — `update` repairs, `doctor` reports).** A bare needs the fetch mapping (`remote.origin.fetch = +refs/heads/*:refs/remotes/origin/*`) and `origin/HEAD`; without them `origin/<branch>` does not exist in its worktrees and the default branch cannot be resolved. `git clone --bare` sets up neither. `add-repo` sets both on every run, for a new or an existing bare, but nothing else does, so a bare made by an older build stays broken until `add-repo` touches that repo again. `update` runs the same setup before it fetches, so it heals as a side effect. `doctor` checks each bare and reports, without changing anything. `prune` depends on this: its unpushed-commit check compares local branches with remote-tracking refs, so a bare with no mapping reports every branch as unpushed.

**Earlier designs, set aside** (all assumed one task per workspace; details are in git history): a `ws/<workspace>` prefix; a branch named after the workspace (git's own precedent: `git worktree add <path>` names the branch after the path's basename) with reuse/error collision handling; and a per-branch git marker recording ownership.

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

**Checkout roles (edit vs reference):** Each repo in a workspace can be tagged "edit" (actively developed) or "reference" (read-only context). For example, `my-service` is the edit target while `shared-lib` and `api-spec` are references. Stored per repo in `.aworkspace.toml` (default `edit`):

```toml
[repos.shared-lib]
url = "git@github.com:org/shared-lib.git"
role = "reference"
```

- **Purely advisory.** It tells agents which code to modify and which to only read. It has no mechanical effect: under the branch model above every repo starts the same way (no branch), and a repo is "being edited" simply once someone branches in it.
- **A reference is still a full worktree,** not a symlink to one shared checkout or an exported snapshot. That keeps real files on disk for grep/LSP/agents, a version pinned per workspace, full history (`git log`, `git blame`, `git grep <pattern> <rev>` all work from a no-branch worktree), and the option to start editing it later. The cost is only the checked-out files; objects live in the shared bare.
- **Bares are full clones — never shallow or partial.** The point is to clone once and have all history reachable from every workspace.
- **Pinning a reference to a version** (e.g. a release tag) is done by hand for now: `git checkout --detach v1.2.3` in the worktree. The recorded SHA preserves it across rehydrate.

**Worktree subdir scoping (and "the workspaces root" as a first-class entity):** The worktree subdir name (`worktree_subdir`, e.g. `code/`) should be **scoped to the workspaces root, not to global config read at per-workspace creation time.** The current flaw: the value is resolved from transient global config when each workspace is created, so one root accumulates workspaces stamped with whatever global config said on each creation day — producing intra-root heterogeneity (`A/code/`, `B/worktrees/`). The configurability isn't the problem; the scope and timing are. Fix with three-level resolution (like git config / mise):

- **global config** (`~/.config/aworkspace/config.toml`) — default for *new roots* only
- **root config** (`~/Workspaces/.aworkspace/meta.toml`) — the value for *every* workspace in this root, written once from the global default at root init
- **`.aworkspace.toml`** — records the resolved value per workspace (self-describing; drives `.gitignore` generation; rarely hand-overridden)

At `new`: read the root value (not global) → fall back to global only to initialize the root → write the resolved value into `.aworkspace.toml`. Result: a root is homogeneous by construction (changing global config can't intermix dirs in an existing root), configurability survives (different roots can differ), and the only thing lost — different worktree dirs *within a single root* — is the footgun, not a feature.

This formalizes **the workspaces root** as a managed entity rather than "whatever `WorkspacesDir` points at." It's the thing you `git init`/`jj git init`, and the root config file is the natural neighbor of the root-level universal `.gitignore` (worktree dirs, `tmp/`, `secrets/`) — both are properties of the collection, set once at the root.

Related rule: **worktrees live in a dedicated, non-empty subdir** (name configurable, emptiness disallowed). `worktree_subdir = ''` puts worktrees at the workspace root intermixed with the record, destroying the clean ignore/layer boundary. (If `''` ever must be supported, the generated `.gitignore` would enumerate worktree dirs by repo name from `.aworkspace.toml` — simpler to just disallow empty.)

**Decided:** root config lives at `.aworkspace/meta.toml` within the workspaces root. Required (created by `CreateWorkspaceRoot`).

**Format versioning (decided — minimum):** The on-disk format carries a version stamp so a newer binary can tell an old workspace from a broken one. Motivation: upgrading a consume-only machine across 24 commits left nine workspaces broken or stale (renamed marker file, missing root `meta.toml`, stale generated `CLAUDE.md`), with nothing to detect or repair it.

- **One integer, not semver.** `schema = 1` in the root `meta.toml` and in each `.aworkspace.toml`. A missing stamp means 1; it is written on the next save.
- **Per-workspace as well as root,** because archived workspaces are frozen at old versions and are loaded later by rehydrate and `new --from`.
- **Bump only on breaking changes.** Adding a field is not breaking — the TOML loader ignores unknown keys.
- **The binary reads only the current version.** No "current plus one back" read paths. Older workspaces are *migrated*, not read: each bump adds one small function (v1→v2, v2→v3) and chaining them upgrades any old workspace. No migration code exists until the first breaking change.
- **Newer than the binary is refused** with one clear message telling the user to upgrade aworkspace. (Matters once two machines sync a journal.)
- **`list` checks root metadata once,** before loading workspaces, and fails once — not one warning per workspace.
- The legacy pre-0.1 `workspace.toml` layout is not migrated.

**Deferred (the stale generated-guidance problem):** `new` writes `CLAUDE.md` once and never again, and the file is also user-editable, so the tool can never safely refresh it. Preferred fix when it bites again: split ownership — a tool-owned, re-rendered `AWORKSPACE.md` (refreshed by single-workspace commands, never by `list`/`cd`/completion), a user-owned `CLAUDE.md` stub that imports it, and a `doctor --fix` that migrates formats and swaps unmodified old `CLAUDE.md` files (matched by past-template hash) for the stub. Rendering the guide from metadata is also where per-repo roles and the configurable worktree subdir would surface.

## Workspace Lifecycle: rm, archive, prune

Design from the 2026-05-29 brainstorm. This is the model behind `rm`/`prune` (see 0.1–0.2 below) and the archive concept.

### Guiding philosophy

> aworkspace marries git worktrees + plain-text context. What you do with the context is yours.

aworkspace produces durable, greppable, plain-text artifacts and stays out of the way of whatever you layer on top — git, jj, Obsidian, Syncthing, or nothing. It does not own your files, enforce policy, or reinvent storage/backup machinery. The recurring stance for risky operations is **visibility, never enforcement**.

### Two independent lifecycle axes

1. **Status (active / inactive)** — a flag in `.aworkspace.toml`. The workspace stays fully intact on disk; "inactive" just hides it from `list` by default. Instantly reversible, cheap. Easy to layer on later; not the current focus.
2. **Physical state (live → archived → purged)** — a transformation. Archiving detaches worktrees and moves the record aside; purging destroys the record. This is the real design work.

### Layer model (Docker analogy)

A workspace is three materials with different reconstructibility, which dictates the safety rules:

| Layer | Workspace content | Reconstructible? | Removed by |
|---|---|---|---|
| Container | Worktrees + working-tree changes (`code/`) | Mostly (remote + branch) | `rm` (default) |
| Image | The **record**: `.aworkspace.toml`, `WORKSPACE.md`, notes/resources | No — irreplaceable | destroyer (`purge`/`delete`, deferred; manual `rm` of the archive in 0.1) |
| Base layers | Shared **bare repos** (live outside the workspace) | Yes (re-clone) | `prune` only |

Safety property: **no single-workspace command can ever delete shared state.** Bare repos are refcounted layers; only the explicit, global `prune` reclaims dangling ones.

### Command family (safe by default)

- **`archive` (default, safe).** Detaches worktrees (the container), keeps the record (the image), moves the workspace dir to `_archive/`. Never deletes shared content. This is the everyday verb — it's named for what it *does* (the connotation matches the safe behavior), unlike `rm`, which over-promised destruction. (Earlier draft called this `rm`; renamed because `rm` reads as "destroy" while the default preserves.)
- **The destroyer (`purge`/`delete`) is deferred past 0.1.** Archives are plain folders, so the v1 recourse for "I really want this gone" is a manual `rm -rf _archive/<name>`. A tool-provided destroyer can come later once we've felt whether it's missed; making it a *separate command* (not an `archive --purge` flag) means destruction requires typing a different, scarier word — consistent with "speed bump proportional to irreversibility." Name still open (`purge` vs `delete` vs `destroy`).
- **`prune`** — global GC of bare repos no *live* workspace references. A separate operation, never a side effect of `archive` (see 0.2). Detects unpushed commits before deleting a bare (see "prune" below).

Primary motivation for archive is **preserving the record/memory**. Resumability (re-create worktrees on the same branches, or `new --from` an archived workspace) is a secondary bonus that falls out of keeping `.aworkspace.toml` + final commit SHAs in the record.

### archive: dirty detection & confirmation (decided)

Under **A1** — archive never deletes shared *content* (branches/objects/the bare stay put; it only detaches each worktree, which updates that worktree's own registration) — archive can only ever destroy state that lives **only in the worktree**. So dirty detection reduces to one question per repo: are there **uncommitted or untracked** changes? Committed work on a branch — pushed or not — is safe on the kept branch, so unpushed commits are reported as FYI, never a gate. (Untracked is counted *after* the ignore rules, so `tmp/`/`secrets/` don't cry wolf.)

**One exception — commits on no branch.** Repos start with no branch (see "Branch model" in Design Notes). Commits made before branching exist only in the worktree's HEAD and are orphaned when it is detached, so they count as dirty: `git rev-list HEAD --not --branches --remotes --tags` non-empty → confirm / `--force`. The same check applies to `remove-repo`.

**Scope: ownership, not location.** The governing invariant — same as branches — is *aworkspace manages every worktree on a bare it created, and nothing else.* It never creates branches or commits, so it never deletes them; it does create bares and worktrees, so it detaches those. Ownership, not the `code/` location, drives archive — the `code/` convention can't be relied on (agents place worktrees elsewhere; see "Extra worktrees" in Design Notes), so a location-based rule would be built on sand.

**Discovery is bare-driven, not a filesystem walk.** For each repo in the metadata, take its bare and ask git what worktrees it owns — `git -C <bare> worktree list --porcelain` — then keep the ones whose path is inside this workspace dir. That set is exactly the worktrees we own here: primaries plus any extras an agent created off the same bare (`git worktree add` from inside a primary inherits that primary's common-dir, so the extra is on our bare), wherever they sit. The bare is authoritative for "what are my worktrees," so no `.git` stat, no `rev-parse --git-common-dir` classification, no clone/foreign ambiguity. Foreign worktrees and standalone clones are on bares we did not create, so they never appear in our bares' lists — excluded for free, never discovered, never touched.

- **Owned worktrees** get the two checks (uncommitted/untracked, commits on no branch) as a gate, then on archive are detached with `git worktree remove` — before the workspace dir is moved, while registrations are still valid. Their branches live in the bare and survive.
- **The workspace-membership filter is load-bearing.** Bares are shared across workspaces, so `worktree list` returns worktrees in *other* workspaces too; keep only those under this workspace dir. Canonicalize both sides (`filepath.EvalSymlinks` + separator-aware prefix, not raw `HasPrefix`) — git reports resolved paths and macOS `/var`→`/private/var` will otherwise mismatch.

There is no "rm the whole worktree subdir" step: archive detaches the worktrees it owns, so the gate scope and the action scope are the same set by construction. After detaching, the leftover folder — the record, plus any non-worktree files and any *foreign* worktrees — is moved to `_archive/` with `os.Rename`.

**Foreign worktrees break on the move — lossless, not clean (honest edge).** Archive doesn't operate on a foreign worktree's git state, but the `os.Rename` of the parent dir moves it on disk, which stales its registration in its own (foreign) bare — exactly like a manual `mv` of a worktree. Files are preserved (no data loss); the worktree→bare link is broken (`git worktree list` in that bare lies; a later `prune` there orphans it); git from inside the moved worktree still works until that bare is pruned. Recoverable with `git worktree repair <moved-path>`. We accept this in the early version: the clean alternatives both cost more than it's worth — *refuse-when-foreign-present* is the gate we deferred, and *repairing/moving* the foreign worktree would be managing one we disclaimed. The deferred foreign **warning + `doctor`** check (see below) is what makes it honest — it tells the user we moved their foreign worktree and to run `git worktree repair`.

**Deferred past the early version (visibility only, no data-loss risk):** detecting foreign worktrees (needs a filesystem walk, since they're not on our bares) to emit the archive-time warning and a `doctor` report. Foreign handling is already *correct* without it — never gated, detached, or repaired — this only *tells* the user what the move did.

Confirmation falls out of "what gets destroyed":

| `archive` | clean worktree | dirty (uncommitted/untracked) |
|---|---|---|
| behavior | proceed, print summary | confirm / `--force` |

- Archive is friction-free only for a *clean* worktree; the dirty case has a real destructive edge, which is exactly where the confirm earns its keep.
- **Never a hard block.** confirm/`--force` is a speed bump proportional to irreversibility, consistent with visibility-never-enforcement.
- Dirty state is reported **per owned worktree** (which path, how many uncommitted/untracked), not a workspace-level boolean.

**Target & cwd:**
- `archive <name>` **requires the workspace name** — it never infers from the current directory. Inferring would mean standing inside the workspace being archived, which is exactly the unsafe case below.
- `archive` **refuses when the invoking shell's cwd is at or under the target workspace dir** — the dir is renamed (and owned worktrees detached), so the cwd would be moved/deleted out from under the shell, leaving a half-dead state. Tells you to `cd` out first. Precedent: `git branch -d` refuses the checked-out branch; Postgres `DROP DATABASE` refuses the connected database.
- The guard is **best-effort, not airtight**: it only inspects the invoking shell. Another terminal sitting inside the same workspace still gets its cwd pulled out from under it — "current workspace" is per-shell, not global state aworkspace tracks. Accepted tradeoff; the guard catches the common case (archiving the one you're standing in).

**Library contract** (the CLI maps prompts onto this): a `Force` flag skips the guard; without it, archive-of-dirty returns a sentinel error the CLI turns into a confirmation prompt.

### prune: unpushed-commit detection (decided)

`prune` is where the unpushed-commit risk A1 deferred finally lands: **archive keeps branches in the bare, and prune deletes the bare** — so prune is the operation that can permanently destroy unpushed work. Detection therefore *must* live here; without it, A1's "unpushed is safe because we keep the branch" is a promise with no enforcement point.

It's the same "does anything exist only here?" question as archive, one layer down:

| Operation | Layer destroyed | "exists only here" = |
|---|---|---|
| `archive` | worktree (container) | uncommitted / untracked |
| `prune` | bare (base layer) | **unpushed commits** |

So prune gets archive's treatment, applied to the bare:
- **Bare fully reconstructible** — every local branch tip is reachable from a remote-tracking ref → a re-clone gets it all back → **prune freely** (this is the whole point of bares being the reconstructible base layer).
- **Bare holds unpushed commits** → **confirm / `--force`**, listing which branches carry them. Never a hard block, but the speed bump is mandatory because the loss is permanent.

Detection is cheap and **offline** (no network): in the bare, `git rev-list --branches --not --remotes` yields every commit on a local branch not reachable from any remote-tracking ref. Non-empty → unpushed work exists; enumerate the branches for the warning. It uses the bare's *last-fetched* remote refs, which is the right call — prune shouldn't phone home.

**Two stages, kept separate:**
1. **Candidacy** — which bares have no *live* workspace reference (the GC roots). Archived workspaces do **not** grant immunity — the unpushed check is the real backstop, so a pushed (re-cloneable) bare is safe to prune regardless of archived references, and an unpushed one gets caught at the safety stage either way.
2. **Safety** — of the candidates, which hold unpushed commits → confirm.

### Archive substrate: plain move-aside folder

Archiving moves the workspace dir (minus dropped layers) to `_archive/` as a **plain folder of files**. Not a tarball (opaque — can't grep or `cd` into it), not git-baked (binaries bloat history; couples aworkspace to a VCS). If you want the archive in version control, that's the orthogonal VCS-tracking story below — your call, not aworkspace's.

### Archive folder naming: `name@date` (decided)

Archive-forever guarantees collisions: you can `new feature1`, archive it, then `new feature1` again (the name frees up once it leaves the live root). So archives are named **`<name>@<YYYY-MM-DD>`** — e.g. `_archive/feature1@2026-06-30`.

- **Always dated, not just on collision.** The date is useful at a glance in `ls _archive/`, sorts re-archives of a name together chronologically, and avoids special-casing the first one.
- **`@` delimiter** is unambiguous precisely because workspace names are validated `[A-Za-z0-9_-]` (no `@`), so it can never appear in a name. (`_`/`-` would blur the name/date boundary.)
- **Same-day, same-name collisions** get a first-free counter: `feature1@2026-06-30`, `feature1@2026-06-30-2`, … (rare path; a terminating "increment to first free" loop).
- **The folder name is never parsed** — frontmatter (`name`, `archived` date) is the authority and `reindex` reads *that*, not the folder. So the name only has to be *readable + unique on disk*; a manual `mv` of an archive breaks nothing. The folder date equals the frontmatter `archived` date (one timestamp).
- **`Archive` returns the resolved path** so the CLI can report where it went and callers/tests don't have to recompute the date.
- **Never overwrite or merge** into an existing archive dir — always land in a fresh, deconflicted path.

### Index: self-describing folders + regenerable ledger

- **Source of truth = the frozen folders.** Archiving stamps structured frontmatter into the record (archived date; repos + branches + final SHAs; a 1–3 sentence summary, optional keywords/ticket). Travels with the folder, survives a manual `mv`, can't be orphaned.
- **`_archive/INDEX.md` = a derived cache** — a compact one-line-per-workspace ledger, written at archive time and fully regenerable by walking the archives (`archive reindex`). Gives agents/tools a token-cheap scan without walking N folders. It is never the authority: delete it and it rebuilds from the folders.
- **The asymmetry is intentional.** Live workspaces are walk-only (small N, mutating — an index would just drift). Archived workspaces get the ledger (large N, frozen — no drift possible). `list` shows live by default; `list --archived` reads the ledger. The friction asymmetry (live = rich, archived = one-liner) is what rewards archiving instead of hoarding (the "Kanban Done column that never ages out" problem).
- **"Frozen" is a convention, not enforced.** Editing or deleting files in an archive breaks nothing — `reindex` reconciles the cache, and the recourse to "I don't want this archived file anymore" is just `rm thefile` (it's a plain folder).

### File-handling conventions: two filters, named directories

Two independent filters: the generated `.gitignore` ("VCS-tracked?") and the archive-exclude list ("Archived?"). Special directories are named points on that grid — conventions with escape hatches, not enforcement (a configurable policy engine can come later; don't build one now):

| Bucket | Examples | VCS-tracked | Archived |
|---|---|---|---|
| Record | `WORKSPACE.md`, notes, `*.toml`/`*.md` | ✅ | ✅ |
| Sensitive | `secrets/`, `.env` | ❌ | ✅ (kept, but stays ignored) |
| Ephemeral | `tmp/` | ❌ | ❌ |
| Reconstructible | `code/` (worktrees) | ❌ | ❌ |

- **`secrets/`** — VCS-ignored **everywhere** via a path-agnostic pattern (so it stays ignored inside `_archive/…/` too) → a self-contained, resumable env that never pushes. Kept on archive (dev-grade keys, local-only, bounded liability; delete manually if undesired). Print a warning when archiving secrets, e.g. `[WARNING] 3 files in secrets/ will be archived.`
- **`tmp/`** — ignore + drop. Pure scratch.
- Caveat: VCS-ignore only guards the git-push foot-gun. Non-VCS sync (Time Machine, Syncthing) of an archive still carries secrets along — the user's tool choice, out of scope. aworkspace is not a watchdog.

### Big files: visibility, never enforcement

Archives are forever, so heavy files in the record layer are paid for forever. At archive time (and optionally in `list`/`show`), surface size and flag the heaviest offenders, e.g. `Archiving feature1: 2.1 GB, 1.9 GB of which is assets/demo.mov. Continue?`. The warn threshold is **configurable**. aworkspace never blocks or quarantines — the same onus the user already carries for source code.

**Non-goal:** aworkspace does not compress, dedup, or transform archived files. Those capabilities already exist one layer down (APFS/ZFS CoW, Time Machine, borg/restic, git-LFS) and the plain-folder output keeps them all usable. Compression also fights the plain-text value; users who want it can script it over the folder.

### VCS tracking: fully hands-off

- Generate a sensible `.gitignore` at `new` time (ignores the worktree subdir + `tmp/` + `secrets/`), disable-able in config — same pattern as auto-`CLAUDE.md`. One file serves git **and** jj (jj reads `.gitignore`).
- **Never `git init` for the user.** A dormant `.gitignore` in a non-git folder is harmless; initializing a repo is reaching into the user's territory. Lay the groundwork; never opt in for them.
- Ignoring `code/` makes the VCS boundary line up exactly with the layer boundary (track record + archive, ignore worktrees, never see bare repos) and sidesteps the embedded-repo problem.
- Tips & tricks doc section ("Using aworkspace with git/jj"): track the workspaces dir on a **single linear history — never branch** (the cross-product of every project's worktree state is the head-spin). jj's working-copy-as-commit auto-snapshot is a strong fit for a frictionless memory journal. jj's LFS/submodule gaps don't affect this use case — the journal is plain text and worktrees are ignored, while project repos inside `code/` keep using git + LFS + submodules independently. The VCS choice for the journal is decoupled from what the projects need.
- **Backup vs. sync guidance** (docs): adding a workspaces root to VCS is useful for historical tracking and off-machine backup. Syncing the repo across multiple machines is up to the user — conflicts, branching, and merge are VCS workflow, not the tool's job. Secrets and other untracked files don't travel with the repo by design.
- **Possible later: an opt-in `sync`** for the one case that stays simple — a single user across a few machines, no live collaboration. Shape: commit the journal, `git pull --rebase`, push; then push *committed* work branches (print what will be pushed and make it confirmable — it creates remote branches teammates see and can trigger CI); report repos with uncommitted changes but **never auto-commit code**. Works at one-workspace and all-workspaces level, and reuses the dirty/unpushed checks. Auto-commit of the journal on `new`/`add-repo`/`archive` could follow, also opt-in. Requires: work branches pushed to their own remotes, and no machine-specific absolute paths in metadata. On the second machine the flow is pull → rehydrate → work.

### Rehydrate: archive's inverse, and the backup-loop closer

A command that makes on-disk reality match the record — recreate missing bares and worktrees from `.aworkspace.toml` (URLs + each repo's recorded branch and SHA). **Name is a working title; candidates: `rehydrate` / `restore` / `materialize` / `doctor --fix`.** The capability is what matters, not the name.

It's more load-bearing than a mere repair tool: **because `code/` is gitignored, a VCS backup deliberately omits the worktrees and bares.** So a fresh clone of the journal on a new machine is just the record — nothing runnable. Rehydrate is the *only* thing that turns that record back into a working setup, which means the moment VCS-backup is a goal, rehydrate stops being optional. It's a single primitive that serves three needs:

- **New machine / disaster recovery** — clone the journal, rehydrate the worktrees.
- **Resume an archived workspace** — archive detached the worktrees; rehydrate reattaches them.
- **Repair** — `code/` got deleted or corrupted; rehydrate restores it.

**Restore boundary** (maps onto the layer model — restores the *reconstructible* layer from the *record*, nothing more):

| Layer | Restored? | From |
|---|---|---|
| Record (notes, `.aworkspace.toml`) | ✅ already present | the VCS clone |
| Bares + worktrees | ✅ recreated | re-clone + worktree add on the recorded branch if it still exists, else no-branch at the recorded SHA |
| `secrets/` | ❌ not restored | gitignored — user re-provides |
| Uncommitted worktree changes | ❌ gone | never in VCS/bares unless pushed |

Secrets not syncing is correct and desired, but it's a sharp edge on a fresh restore — so rehydrate should *say so*: a one-line `secrets/ is gitignored and was not restored — re-add your env files` (visibility, never enforcement, same as the big-files warning). Out of scope: multi-machine conflicts, branching, and sharing of the journal repo — that's the user's VCS workflow.

### Seeding: `new --from` (fork / template)

`new <name> --from <workspace>` creates a new workspace seeded from an existing one. It is a **seeded copy, not a link**: the two share an origin and nothing else — no syncing, no parent/child relationship. ("Fork" in the git-hosting sense; not "branch".)

Distinct from rehydrate: rehydrate restores the *same* workspace; `--from` makes a *new* one.

**One mechanism, three sources:**

- **A live workspace** — spin a parallel line of work off an epic or the default workspace, keeping its context.
- **An archived workspace** — restart old work under a new name.
- **A template** — a workspace kept only to be copied from. Templates are not a separate feature: a template is a workspace you seed from. (Scaffolding for plain `new` — custom `CLAUDE.md`, standard files — may still want its own config; revisit once `--from` exists.)

**What is copied:**

| Content | Copied? |
|---|---|
| Repo set and roles (`.aworkspace.toml`) | ✅ |
| Context files (`WORKSPACE.md`, `CLAUDE.md`, notes/resources) | ✅ |
| `tmp/`, `secrets/` | ❌ (same warning as rehydrate for `secrets/`) |
| Worktrees / uncommitted changes | ❌ — fresh worktrees are created |

- **Provenance only:** the new workspace records `seeded_from = "<name>"` (plus date). Informational — nothing reads it to keep the two in sync.

**Open: where the copy's repos start.**

- **At each repo's default branch** — what a template wants (fresh start, latest code).
- **At the commits the source is sitting on** — what a mid-epic fork wants (continue from here). The copy cannot reuse the source's branches while the source is live (one branch, one worktree), so this means no-branch at the same SHAs.

Possibly a flag, with the default branch as the default.

### Parked implementation TODOs

- **Archived folder naming/deconfliction** — **decided:** `name@date` with a same-day `-N` fallback; friendly name preserved in frontmatter. See "Archive folder naming" above.
- `archive reindex` — rebuild `_archive/INDEX.md` from the folders.
- Possible `archive clean` to bulk-prune old archives (but it's flat files — easy to do manually).
- Possible `import-resource` using APFS `clonefile` for cheap CoW copies — only if aworkspace is ever the one doing the copying.
- `inactive` flag in `.aworkspace.toml` + `list` filtering — an easy layer-on.

## Milestone 0.1 - Core Functionality

Goal: Get the basic workspace management working. Create, list, and manage workspaces with multiple repos.

### Must Have

- [x] **Core data types** — `Workspace`, `Repo`, `Config` structs in `internal/workspace/`
- [x] **Config loading** — Read/write `~/.config/aworkspace/config.toml` with defaults
- [x] **Workspace discovery** — Find workspace by walking up from cwd to locate `.aworkspace.toml`
- [x] **`aworkspace new <name>`** — Create workspace directory with `.aworkspace.toml`, `WORKSPACE.md`, and `CLAUDE.md`
  - `CLAUDE.md` includes default workspace isolation rules for agents
  - Makes agents immediately understand that repos are independent projects
- [x] **`aworkspace list`** — List all workspaces (basic: name only)
- [x] **`aworkspace show`** — Display current workspace info (repos, branches, status)
- [x] **`aworkspace add-repo <url>`** — Clone bare repo + create worktree
  - Handle bare clone creation
  - Update `.aworkspace.toml` with repo metadata
  - **Support multiple different repos per workspace**
  - [ ] **Switch to the no-branch model** — worktree starts detached at the remote's default branch; remove the branch argument, `defaultBranchName` and `branch_prefix`; print the "no branch / start work" line. See "Branch model" in Design Notes.
- [ ] **Orphan-commit check** — `archive` and `remove-repo` treat commits on no branch as dirty (confirm / `--force`); `show`/`status` flag them.
- [x] **`-C <path>` flag** — Override working directory for commands that depend on cwd (`show`, `add-repo`)
- [x] **Basic tests** — Unit tests for workspace discovery, config loading, path handling

### Nice to Have

- [x] **`aworkspace cd <name>`** — Output path for shell integration
- [x] **`aworkspace init <shell>`** — Shell integration setup (emits shell wrapper for `cd`)
- [x] **Tab completion for `aworkspace cd`** — Complete workspace names; critical for usability of `cd`
- [x] **`aworkspace archive [workspace]`** — Safe-by-default archive (detach worktrees, keep the record, move to `_archive/`). Dirty (uncommitted/untracked) worktrees need `--force`/confirm; clean ones go through silently. No tool-provided destroyer in 0.1 — manual `rm -rf _archive/<name>` is the recourse. See "Workspace Lifecycle" above.
- [x] **`aworkspace remove-repo <repo>`** — Inverse of `add-repo`: detach the repo's worktree, drop it from `.aworkspace.toml`. Dirty worktree needs `--force`. Refuses when cwd is at/under the worktree (guard parity with `archive`). Keeps the branch in the bare.
  - Branch cleanup: **decided — none.** The tool never deletes branches (see "Branch model").
- [ ] **Extra-worktree safety** — `archive` discovers owned worktrees bare-driven (`git -C <bare> worktree list`, filtered to this workspace), so primaries *and* agent-created extras on our bares get the dirty + orphan-commit gate and `git worktree remove` (replaces `rm -rf code/`); drop the archive kill switch. Foreign worktrees excluded for free (not on our bares). `remove-repo` warns when the repo has extra worktrees. See "Extra worktrees" and "archive: dirty detection" in Design Notes.
- [ ] **Foreign-worktree visibility** (deferred) — filesystem walk to detect worktrees *not* on our bares; emit archive-time warning (move breaks their registration, `git worktree repair` recovers) and a `doctor` report. Visibility only — foreign handling is already correct without it.
- [ ] **Worktree guidance in generated `CLAUDE.md`** — location (`code/<repo>--<task>`), always on a branch, creator cleans up.
- [ ] **Checkout roles** — advisory `role` field per repo (`edit` default, `reference`), surfaced to agents via the generated `CLAUDE.md`. No mechanical effect. See "Checkout roles" in Design Notes.
- [ ] **Format version stamp** — `schema = 1` in root `meta.toml` and each `.aworkspace.toml` (missing = 1); refuse a schema newer than the binary with one clear message; `list` checks root metadata once instead of warning per workspace. See "Format versioning" in Design Notes.
- [ ] **Better error messages** — User-friendly errors with suggestions
  - [ ] **`add-repo` on a remote with no commits** — today fails with git's raw `error: Cannot determine remote HEAD` (from `git remote set-head origin --auto`). Detect it after the fetch (`git -C <bare> for-each-ref refs/remotes/origin` is empty) and return a clear message: `repo has no commits yet; push an initial commit, then run add-repo again`. The bare clone may stay — the retry reuses it. Not supported on purpose: a no-branch worktree needs a commit to sit at, and `git worktree add --orphan -b <branch>` would need a branch name, against the branch model. Documented workflow for a brand-new repo: create it with an initial commit (or push one), then `add-repo`.

### Out of Scope for 0.1

- A command to create/manage extra worktrees of the same repo (the convention and the archive scan are in scope — see "Extra worktrees" in Design Notes)
- Bookmarks
- `update`/`reset` commands
- `prune` command
- `doctor` command
- Submodule support
- Web dashboard

## Milestone 0.2 - Multi-Repo & Polish

- [ ] Shell completion (bash/zsh/fish)
- [ ] Submodule support (with git-worktree-tools integration)
- [ ] **Destroyer command (`purge`/`delete`)** — deferred from 0.1; separate command, not an `archive` flag. Name TBD.
- [ ] Archive index — stamp frontmatter into the record + write/regenerate `_archive/INDEX.md` ledger (`archive reindex`); `list --archived`
- [ ] File-handling conventions — generate default `.gitignore` (worktree subdir + `tmp/` + `secrets/`), archive-exclude `tmp/`/`code/`, keep-but-ignore `secrets/` with size + secrets warnings
- [ ] `aworkspace prune` for orphaned bare repos (global GC, never a side effect of `archive`) — detects unpushed commits before deleting a bare; confirm/`--force`, never a hard block. See "prune: unpushed-commit detection" above.
- [ ] `aworkspace update` — fetch; move untouched no-branch repos to the new default-branch tip. Repos on a branch are the user's to rebase (open, see "Branch model"). Before fetching, set the fetch mapping and `origin/HEAD` on each bare (the same setup `add-repo` does), so bares made by older builds heal. See "Healing old bares".
- [ ] `aworkspace reset` — reset workspace to clean state
- [ ] Bookmarks for common git hosts
- [ ] `new --from <workspace>` — seed a new workspace from a live, archived or template workspace (seeded copy, no link). See "Seeding: `new --from`" above.
- [ ] **Configurable `CLAUDE.md` generation** — Config option to disable auto-creation of `CLAUDE.md`
- [ ] **Verbosity levels** — Default (friendly short output), `--quiet` (silent), `--verbose` (include git command output)
- [ ] **Workspace templates** — User-defined templates for `new` command (scaffolding, standard files, custom CLAUDE.md). Largely covered by `new --from` (a template is a workspace you seed from); keep only what that doesn't cover.

## Milestone 0.3 - Quality of Life

- [ ] `aworkspace doctor` — environment checks
  - Bare repo health (report only): fetch mapping present, `origin/HEAD` set. Points the user at `update` to repair. See "Healing old bares".
- [ ] Better git URL parsing (support all formats)
- [ ] Homebrew formula
- [ ] **Rich `list` output** — flexible formatting like `ls -l`
  - `-l, --long` — detailed format (status, repo count, branches, last modified)
  - `-1` — single column (name only, one per line)
  - `--format <template>` — custom output format
  - Sortable fields (name, created, modified, status)
  - Filter by status or other attributes
- [ ] **Optional workspace subtitle** — One-line description field in `.aworkspace.toml` (e.g., `subtitle = "Q2 nav rewrite"`), shown in `list -l`

## Future

- Web dashboard for session management
- Agent orchestration integration
- Workspace templates
- Workspace sharing/export
- Ceiling paths for workspace discovery (à la `MISE_CEILING_PATHS`) — stop walking up at configured boundaries (e.g., network mounts)
- Workspaces outside `WorkspacesDir` — already supported by discovery, but `list` and other commands may need awareness
- **Multiple workspace roots** (e.g. work vs personal) — each root is an independent collection with its own root config and (optionally) its own VCS history. Resolution precedence: `--root` flag → `AWORKSPACE_ROOT` env var → global-config default. Shell integration could switch roots like `cd` does. Composes cleanly with the root-scoped-config design above — a root is already self-describing, so "more than one" needs no new model. Later nicety: a named-roots registry in global config (`[roots] work = "…"`) for `--root work` by name + completion. Only commands that *enumerate* (`list`) or *create* (`new`) care about the active root; commands inside a workspace still infer from cwd. Not needed for first releases.
- `aworkspace cd` with no arg — interactive picker (à la `mise use`)
- **Stale-cwd detection in shell integration** — a `precmd`/`chpwd` hook in `aworkspace init` that warns when the shell's cwd is no longer a live workspace (vanished, or moved under `_archive/`). Covers the "other shell was sitting in a workspace that got archived" gap the `archive` cwd-guard can't reach (it only inspects the invoking shell). Cheap (one stat per prompt); same mechanism direnv/mise use. Not archive-specific — general "cwd left the workspace" signal.
