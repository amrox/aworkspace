# aworkspace Roadmap

## Design Notes

**Branch naming:** The default branch is the **(slugified) workspace name**, the same name in every repo of the workspace — no prefix. Workspace `my-feature` → branch `my-feature`.

Guiding principle: *the branch name describes the work, not the plumbing.* The earlier `ws/<workspace>` scheme was called "plumbing," but the user sees it everywhere (push, PRs, prompts) — so it isn't plumbing, it's porcelain. Two concrete wins from dropping the prefix:

- **Local name = remote name.** No translation when you push or open a PR, no `ws/foo`→`feature/login` mental mapping. This was the single most surprising thing about the prefix scheme.
- **Bookkeeping moves to metadata, not the name.** "Which branches belong to this workspace?" is answered by `.aworkspace.toml` / the workspace dir, not by a `grep ws/`. The branch name is freed to be natural. (Cost: discoverability/cleanup now depend on the metadata store being trustworthy — `doctor` reconciles metadata vs. actual worktrees/branches.)

A `branch_prefix` config (default `""`) remains for users who deliberately want namespaced/throwaway branches.

**Collision handling** (least-surprising order — never silent):
1. Branch doesn't exist → create it from the base branch.
2. Branch exists and isn't checked out in any worktree → **reuse it** (attach the worktree). Silently forking to `my-feature-2` when `my-feature` is sitting right there is the surprising move.
3. Branch exists **and** is checked out in another worktree (git forbids sharing) → **error, don't auto-rename.**

Why error rather than auto-disambiguate (e.g. `my-feature-2`): with unique workspace names and clean teardown, this collision is *exceptional* — it almost always means a dirty state (a workspace removed without cleanup left a stale worktree holding the branch). Silently minting a `-2` branch reintroduces exactly the surprising, non-`local==remote` name we dropped `ws/` to avoid, and it produces litter. So we shout instead:

- Fail with a clear message naming the worktree that holds the branch, e.g. `branch 'my-feature' is already checked out at <path>; run 'aworkspace doctor' or pass an explicit branch`.
- The user's escape hatches are already there: `add-repo <url> <branch>` to pick a different name, or `doctor`/`prune` to clean up the stale holder.
- (An earlier draft proposed a `<workspace>+<repo>` suffix. Dropped: the suffix is constant per (workspace, repo) so it's idempotent — it can't disambiguate a *second* collision, and it can't disambiguate two worktrees of the *same* repo at all.)

**Visibility is the real least-surprise lever.** `new`/`add-repo` print exactly what they did, e.g.:

```
workspace 'fix-auth'
  api     → branch fix-auth (new, from main)
  shared  → branch fix-auth (reusing existing)
```

**Base branch:** Each repo can specify a `base_branch` — what the workspace branch forks from. Stored in `.aworkspace.toml` per-repo (metadata, **not** in the branch name):

```toml
[repos.my-service]
url = "git@github.com:org/my-service.git"
base_branch = "main"
```

- `add-repo` starts the branch from `origin/<base_branch>`: `git worktree add <dest> -b my-feature origin/main` (case 1 above).
- Normal git upstream applies — once pushed, `my-feature` tracks `origin/my-feature` (set on first push). We do **not** set upstream to `origin/<base_branch>`: that would make `git push` target the base branch — exactly the surprising indirection we're removing. Ahead/behind-of-base is computed explicitly from metadata by `status`/`update` instead.
- If omitted, defaults to the remote's HEAD (the repo's default branch) and is stored explicitly so the metadata is self-describing.
- `update` (0.2) fetches and rebases/merges from `origin/<base_branch>`.
- Useful for reference repos where you want to pin to `main` even if the repo defaults to `dev`.

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

**Checkout roles (edit vs reference):** Each repo checkout in a workspace could be tagged as either "edit" (actively being developed) or "reference" (read-only context for agents). For example, a workspace might have `my-service` as the edit target while `shared-lib` and `api-spec` are references. This distinction is purely advisory — it helps agents understand which code they should be modifying vs. just reading for context. Could be a field in `.aworkspace.toml` per-repo entry.

**Worktree subdir scoping (and "the workspaces root" as a first-class entity):** The worktree subdir name (`worktree_subdir`, e.g. `code/`) should be **scoped to the workspaces root, not to global config read at per-workspace creation time.** The current flaw: the value is resolved from transient global config when each workspace is created, so one root accumulates workspaces stamped with whatever global config said on each creation day — producing intra-root heterogeneity (`A/code/`, `B/worktrees/`). The configurability isn't the problem; the scope and timing are. Fix with three-level resolution (like git config / mise):

- **global config** (`~/.config/aworkspace/config.toml`) — default for *new roots* only
- **root config** (`~/Workspaces/.aworkspace/meta.toml`) — the value for *every* workspace in this root, written once from the global default at root init
- **`.aworkspace.toml`** — records the resolved value per workspace (self-describing; drives `.gitignore` generation; rarely hand-overridden)

At `new`: read the root value (not global) → fall back to global only to initialize the root → write the resolved value into `.aworkspace.toml`. Result: a root is homogeneous by construction (changing global config can't intermix dirs in an existing root), configurability survives (different roots can differ), and the only thing lost — different worktree dirs *within a single root* — is the footgun, not a feature.

This formalizes **the workspaces root** as a managed entity rather than "whatever `WorkspacesDir` points at." It's the thing you `git init`/`jj git init`, and the root config file is the natural neighbor of the root-level universal `.gitignore` (worktree dirs, `tmp/`, `secrets/`) — both are properties of the collection, set once at the root.

Related rule: **worktrees live in a dedicated, non-empty subdir** (name configurable, emptiness disallowed). `worktree_subdir = ''` puts worktrees at the workspace root intermixed with the record, destroying the clean ignore/layer boundary. (If `''` ever must be supported, the generated `.gitignore` would enumerate worktree dirs by repo name from `.aworkspace.toml` — simpler to just disallow empty.)

**Decided:** root config lives at `.aworkspace/meta.toml` within the workspaces root. Required (created by `CreateWorkspaceRoot`).

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
| Image | The **record**: `.aworkspace.toml`, `WORKSPACE.md`, notes/resources | No — irreplaceable | `rm --purge` |
| Base layers | Shared **bare repos** (live outside the workspace) | Yes (re-clone) | `prune` only |

Safety property: **no single-workspace command can ever delete shared state.** Bare repos are refcounted layers; only the explicit, global `prune` reclaims dangling ones.

### Command family (safe by default)

- **`rm` = archive (default, safe).** Detaches worktrees (the container), keeps the record (the image), moves the workspace dir to `_archive/`. Never touches bare repos. This is the everyday, muscle-memory verb — destruction is opt-in.
- **`rm --purge`** — also deletes the record. Irreversible.
- **`prune`** — global GC of bare repos no workspace references. A separate operation, never a side effect of `rm` (see 0.2).

Primary motivation for archive is **preserving the record/memory**. Resumability (re-create worktrees on the same branches, or `new --from` an archived workspace) is a secondary bonus that falls out of keeping `.aworkspace.toml` + final commit SHAs in the record.

### Archive substrate: plain move-aside folder

Archiving moves the workspace dir (minus dropped layers) to `_archive/` as a **plain folder of files**. Not a tarball (opaque — can't grep or `cd` into it), not git-baked (binaries bloat history; couples aworkspace to a VCS). If you want the archive in version control, that's the orthogonal VCS-tracking story below — your call, not aworkspace's.

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

### Rehydrate: archive's inverse, and the backup-loop closer

A command that makes on-disk reality match the record — recreate missing bares and worktrees from `.aworkspace.toml` (URLs + base branches + recorded SHAs). **Name is a working title; candidates: `rehydrate` / `restore` / `materialize` / `doctor --fix`.** The capability is what matters, not the name.

It's more load-bearing than a mere repair tool: **because `code/` is gitignored, a VCS backup deliberately omits the worktrees and bares.** So a fresh clone of the journal on a new machine is just the record — nothing runnable. Rehydrate is the *only* thing that turns that record back into a working setup, which means the moment VCS-backup is a goal, rehydrate stops being optional. It's a single primitive that serves three needs:

- **New machine / disaster recovery** — clone the journal, rehydrate the worktrees.
- **Resume an archived workspace** — archive detached the worktrees; rehydrate reattaches them.
- **Repair** — `code/` got deleted or corrupted; rehydrate restores it.

**Restore boundary** (maps onto the layer model — restores the *reconstructible* layer from the *record*, nothing more):

| Layer | Restored? | From |
|---|---|---|
| Record (notes, `.aworkspace.toml`) | ✅ already present | the VCS clone |
| Bares + worktrees | ✅ recreated | re-clone + worktree add at recorded base/SHA |
| `secrets/` | ❌ not restored | gitignored — user re-provides |
| Uncommitted worktree changes | ❌ gone | never in VCS/bares unless pushed |

Secrets not syncing is correct and desired, but it's a sharp edge on a fresh restore — so rehydrate should *say so*: a one-line `secrets/ is gitignored and was not restored — re-add your env files` (visibility, never enforcement, same as the big-files warning). Out of scope: multi-machine conflicts, branching, and sharing of the journal repo — that's the user's VCS workflow.

### Parked implementation TODOs

- **Archived folder naming/deconfliction** — archive-forever guarantees eventual `feature1` collisions. Leaning `name@date` (e.g. `_archive/feature1@2026-05-02/`) with a same-day `-N` fallback; friendly name preserved in frontmatter. Doesn't affect the design.
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
- [x] **`aworkspace add-repo <url> [branch]`** — Clone bare repo + create worktree
  - Handle bare clone creation
  - Create worktree with branch
  - Update `.aworkspace.toml` with repo metadata
  - Branch = workspace name by default (no prefix); optional configurable prefix; collision handling per Design Notes
  - **Support multiple different repos per workspace**
- [x] **`-C <path>` flag** — Override working directory for commands that depend on cwd (`show`, `add-repo`)
- [x] **Basic tests** — Unit tests for workspace discovery, config loading, path handling

### Nice to Have

- [x] **`aworkspace cd <name>`** — Output path for shell integration
- [x] **`aworkspace init <shell>`** — Shell integration setup (emits shell wrapper for `cd`)
- [x] **Tab completion for `aworkspace cd`** — Complete workspace names; critical for usability of `cd`
- [ ] **`aworkspace rm [workspace]`** — Safe-by-default archive (detach worktrees, keep the record, move to `_archive/`); `--purge` to destroy the record. See "Workspace Lifecycle" above.
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
- [ ] `aworkspace rm` with uncommitted change detection (warn before archiving dirty/unpushed worktrees)
- [ ] Archive index — stamp frontmatter into the record + write/regenerate `_archive/INDEX.md` ledger (`archive reindex`); `list --archived`
- [ ] File-handling conventions — generate default `.gitignore` (worktree subdir + `tmp/` + `secrets/`), archive-exclude `tmp/`/`code/`, keep-but-ignore `secrets/` with size + secrets warnings
- [ ] `aworkspace prune` for orphaned bare repos (global GC, never a side effect of `rm`)
- [ ] `aworkspace update` — fetch and rebase workspace branches
- [ ] `aworkspace reset` — reset workspace to clean state
- [ ] Bookmarks for common git hosts
- [ ] `--from` flag for cloning workspaces
- [ ] **Configurable `CLAUDE.md` generation** — Config option to disable auto-creation of `CLAUDE.md`
- [ ] **Verbosity levels** — Default (friendly short output), `--quiet` (silent), `--verbose` (include git command output)
- [ ] **Workspace templates** — User-defined templates for `new` command (scaffolding, standard files, custom CLAUDE.md)

## Milestone 0.3 - Quality of Life

- [ ] `aworkspace doctor` — environment checks
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
