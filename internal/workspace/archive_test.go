//go:build integration

package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These tests pin the archive behavior decided in ROADMAP.md
// ("Workspace Lifecycle", "archive: dirty detection & confirmation", and
// "Archive folder naming: name@date"). The destroyer (purge/delete) is deferred
// past 0.1, so it isn't covered here.
//
// Proposed library surface (rename if you prefer):
//
//	type ArchiveOptions struct {
//		Force bool // skip the dirty/confirm guard
//	}
//	func (ws *Workspace) Archive(opts ArchiveOptions, config Config) (archivePath string, err error)
//
// Contract:
//   - clean worktree            -> ok, returns the archive path
//   - dirty worktree, !Force    -> error (nothing moved)
//   - dirty worktree, Force      -> ok
//   - re-archiving a reused name -> distinct, deconflicted paths (never overwrites)

// addDirtyFile writes an untracked file into a repo's worktree so the worktree
// reads as dirty (uncommitted/untracked = "exists only here").
func addDirtyFile(t *testing.T, ws Workspace, config Config, repoName string) {
	t.Helper()
	wt := filepath.Join(ws.Path, config.WorktreeSubdir, repoName)
	if err := os.WriteFile(filepath.Join(wt, "wip.txt"), []byte("uncommitted"), 0644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}
}

func TestArchiveIntegration(t *testing.T) {

	t.Run("clean archive: moves record under name@date, detaches worktree, keeps branch", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "feature1")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, "", config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)
		origPath := ws.Path

		archived, err := ws.Archive(ArchiveOptions{}, config)
		if err != nil {
			t.Fatalf("Archive: %v", err)
		}

		// landed under _archive/ as a "<name>@<date>" folder
		if got := filepath.Dir(archived); got != filepath.Join(config.WorkspacesDir, "_archive") {
			t.Errorf("archive parent = %q, want %q", got, filepath.Join(config.WorkspacesDir, "_archive"))
		}
		if base := filepath.Base(archived); !strings.HasPrefix(base, "feature1@") {
			t.Errorf("archive folder %q: want 'feature1@<date>' form", base)
		}

		// original workspace dir is gone (moved aside)
		if _, err := os.Stat(origPath); !os.IsNotExist(err) {
			t.Errorf("expected original workspace dir %q gone, stat err = %v", origPath, err)
		}
		// record survives in the archive
		if _, err := os.Stat(filepath.Join(archived, workspaceMetaFile)); err != nil {
			t.Errorf("expected record %q in archive: %v", workspaceMetaFile, err)
		}
		// the worktree (container) is NOT archived
		archivedWorktree := filepath.Join(archived, config.WorktreeSubdir, repoName)
		if _, err := os.Stat(archivedWorktree); !os.IsNotExist(err) {
			t.Errorf("expected worktree %q absent in archive, stat err = %v", archivedWorktree, err)
		}
		// A1: branch still lives in the bare (shared content untouched)
		bare, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("ensureBareRepo: %v", err)
		}
		if err := exec.Command("git", "-C", bare, "rev-parse", "--verify", "feature1").Run(); err != nil {
			t.Errorf("expected branch 'feature1' to remain in bare after archive: %v", err)
		}
	})

	t.Run("dirty archive without --force: errors and moves nothing", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "feature2")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, "", config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		addDirtyFile(t, ws, config, filepath.Base(remoteRepo))
		origPath := ws.Path

		if _, err := ws.Archive(ArchiveOptions{}, config); err == nil {
			t.Fatalf("Archive on dirty worktree: expected error, got <nil>")
		}

		// workspace stays exactly where it was — no partial move
		if _, err := os.Stat(filepath.Join(origPath, workspaceMetaFile)); err != nil {
			t.Errorf("expected workspace to remain in place after refused archive: %v", err)
		}
		// nothing landed in _archive
		entries, _ := os.ReadDir(filepath.Join(config.WorkspacesDir, "_archive"))
		if len(entries) != 0 {
			t.Errorf("expected no archive entries after refused archive, got %d", len(entries))
		}
	})

	t.Run("dirty archive with --force: proceeds", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "feature3")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, "", config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		addDirtyFile(t, ws, config, filepath.Base(remoteRepo))

		archived, err := ws.Archive(ArchiveOptions{Force: true}, config)
		if err != nil {
			t.Fatalf("Archive(force) on dirty worktree: %v", err)
		}
		if _, err := os.Stat(filepath.Join(archived, workspaceMetaFile)); err != nil {
			t.Errorf("expected record in archive after forced archive: %v", err)
		}
	})

	t.Run("re-archiving a reused name never overwrites (name@date deconfliction)", func(t *testing.T) {
		// Archive feature1, then recreate feature1 in the SAME root (the name frees
		// up once it leaves the live workspaces dir) and archive again. The two must
		// land in distinct folders — nothing overwritten.
		ws1, config := createTestWorkspace(t, "feature1")
		first, err := ws1.Archive(ArchiveOptions{}, config)
		if err != nil {
			t.Fatalf("Archive #1: %v", err)
		}

		// recreate in the same workspaces root
		ws2, err := CreateWorkspace("feature1", config)
		if err != nil {
			t.Fatalf("recreate workspace: %v", err)
		}
		ws2, err = LoadWorkspace(ws2.Path)
		if err != nil {
			t.Fatalf("LoadWorkspace: %v", err)
		}
		second, err := ws2.Archive(ArchiveOptions{}, config)
		if err != nil {
			t.Fatalf("Archive #2: %v", err)
		}

		if first == second {
			t.Fatalf("re-archive reused the same path %q (would overwrite)", first)
		}
		for _, p := range []string{first, second} {
			if _, err := os.Stat(filepath.Join(p, workspaceMetaFile)); err != nil {
				t.Errorf("expected record in archive %q: %v", p, err)
			}
		}
	})
}
