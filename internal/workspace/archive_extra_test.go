//go:build integration

package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// These pin the extra-worktree-safety behavior: archive owns every worktree on
// a bare it created (primary + agent-made extras, wherever they sit) and never
// touches a worktree on a foreign bare. See ROADMAP "archive: dirty detection"
// and "Extra worktrees".

// addExtraWorktree registers an extra worktree of repoName's bare at code/<suffix>.
func addExtraWorktree(t *testing.T, ws Workspace, config Config, remoteRepo, suffix string) string {
	t.Helper()
	bare := filepath.Join(config.BaresDir, "_", remoteRepo)
	path := filepath.Join(ws.Path, config.WorktreeSubdir, suffix)
	if out, err := exec.Command("git", "-C", bare, "worktree", "add", "--detach", path).CombinedOutput(); err != nil {
		t.Fatalf("worktree add %s: %v\n%s", suffix, err, out)
	}
	return path
}

func TestArchiveExtraWorktreeIntegration(t *testing.T) {

	t.Run("extra worktree on our bare is detached, not left in archive", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "extra1")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)
		addExtraWorktree(t, ws, config, remoteRepo, repoName+"--extra")

		archived, err := ws.Archive(ArchiveOptions{}, config)
		if err != nil {
			t.Fatalf("Archive: %v", err)
		}

		// extra detached like the primary — absent from the archive
		archivedExtra := filepath.Join(archived, config.WorktreeSubdir, repoName+"--extra")
		if _, err := os.Stat(archivedExtra); !os.IsNotExist(err) {
			t.Errorf("expected extra worktree %q absent in archive, stat err = %v", archivedExtra, err)
		}
		// registration cleared in the bare
		bare := filepath.Join(config.BaresDir, "_", remoteRepo)
		out, err := exec.Command("git", "-C", bare, "worktree", "list", "--porcelain").Output()
		if err != nil {
			t.Fatalf("worktree list: %v", err)
		}
		for _, wt := range parseWorktreeList(string(out)) {
			if !wt.Bare {
				t.Errorf("bare still registers a worktree after archive: %q", wt.Path)
			}
		}
	})

	t.Run("dirty extra blocks archive without --force", func(t *testing.T) {
		// The gate must scan extras, not just metadata primaries.
		ws, config := createTestWorkspace(t, "extra2")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)
		extra := addExtraWorktree(t, ws, config, remoteRepo, repoName+"--extra")
		// primary is clean; only the extra is dirty
		if err := os.WriteFile(filepath.Join(extra, "wip.txt"), []byte("uncommitted"), 0644); err != nil {
			t.Fatalf("write dirty file: %v", err)
		}
		origPath := ws.Path

		if _, err := ws.Archive(ArchiveOptions{}, config); err == nil {
			t.Fatalf("Archive with dirty extra: expected error, got <nil>")
		}
		// nothing moved
		if _, err := os.Stat(filepath.Join(origPath, workspaceMetaFile)); err != nil {
			t.Errorf("expected workspace to remain in place: %v", err)
		}
		entries, _ := os.ReadDir(filepath.Join(config.WorkspacesDir, "_archive"))
		if len(entries) != 0 {
			t.Errorf("expected no archive entries after refused archive, got %d", len(entries))
		}
	})

	t.Run("foreign worktree rides into archive untouched", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "foreign1")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}

		// a worktree from a repo aworkspace never created — not on our bare
		foreignRepo := createTestRepo(t)
		foreignWt := filepath.Join(ws.Path, config.WorktreeSubdir, "foreign")
		if out, err := exec.Command("git", "-C", foreignRepo, "worktree", "add", "--detach", foreignWt).CombinedOutput(); err != nil {
			t.Fatalf("foreign worktree add: %v\n%s", err, out)
		}

		archived, err := ws.Archive(ArchiveOptions{}, config)
		if err != nil {
			t.Fatalf("Archive with foreign worktree: %v", err)
		}
		// foreign worktree not discovered, not detached → preserved in the archive
		archivedForeign := filepath.Join(archived, config.WorktreeSubdir, "foreign")
		if _, err := os.Stat(archivedForeign); err != nil {
			t.Errorf("expected foreign worktree preserved in archive at %q: %v", archivedForeign, err)
		}
	})
}
