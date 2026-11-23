//go:build integration

package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWorktreesByRepoIntegration(t *testing.T) {

	t.Run("returns only worktrees in this workspace", func(t *testing.T) {

		// ws1 and ws2 back onto the same shared bare. WorktreesByRepo on ws1
		// must return ws1's worktree only, never ws2's.
		ws1, config := createTestWorkspace(t, "ws-one")
		remoteRepo := createTestRepo(t)
		if err := ws1.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo ws1: %v", err)
		}

		ws2, err := CreateWorkspace("ws-two", config)
		if err != nil {
			t.Fatalf("CreateWorkspace ws2: %v", err)
		}
		if ws2, err = LoadWorkspace(ws2.Path); err != nil {
			t.Fatalf("LoadWorkspace ws2: %v", err)
		}
		if err := ws2.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo ws2: %v", err)
		}

		repoName := filepath.Base(remoteRepo)

		m, err := ws1.WorktreesByRepo(config)
		if err != nil {
			t.Fatalf("WorktreesByRepo: %v", err)
		}
		got := m[repoName]
		if len(got) != 1 {
			t.Fatalf("got %d worktrees, want 1: %+v", len(got), got)
		}
		under, err := isUnderDir(ws1.Path, got[0].Path)
		if err != nil {
			t.Fatalf("isUnderDir: %v", err)
		}
		if !under {
			t.Errorf("worktree %q not under ws1 %q", got[0].Path, ws1.Path)
		}
	})

	t.Run("captures an extra worktree on the same bare", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "extra-ws")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)

		// An agent-style extra worktree off the same bare, not in metadata.
		bare := filepath.Join(config.BaresDir, "_", remoteRepo)
		extra := filepath.Join(ws.Path, config.WorktreeSubdir, repoName+"--extra")
		if out, err := exec.Command("git", "-C", bare, "worktree", "add", "--detach", extra).CombinedOutput(); err != nil {
			t.Fatalf("worktree add extra: %v\n%s", err, out)
		}

		m, err := ws.WorktreesByRepo(config)
		if err != nil {
			t.Fatalf("WorktreesByRepo: %v", err)
		}
		if got := len(m[repoName]); got != 2 {
			t.Fatalf("got %d worktrees, want 2 (primary + extra): %+v", got, m[repoName])
		}
	})

	t.Run("skips a prunable worktree without erroring", func(t *testing.T) {

		// Regression: a prunable worktree's dir is gone, so canonicalizing its
		// path would fail. WorktreesByRepo must skip it, not error.
		ws, config := createTestWorkspace(t, "prune-ws")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)

		bare := filepath.Join(config.BaresDir, "_", remoteRepo)
		gone := filepath.Join(ws.Path, config.WorktreeSubdir, repoName+"--gone")
		if out, err := exec.Command("git", "-C", bare, "worktree", "add", "--detach", gone).CombinedOutput(); err != nil {
			t.Fatalf("worktree add gone: %v\n%s", err, out)
		}
		// Remove the dir on disk without pruning → the bare lists it prunable.
		if err := os.RemoveAll(gone); err != nil {
			t.Fatalf("RemoveAll: %v", err)
		}

		m, err := ws.WorktreesByRepo(config)
		if err != nil {
			t.Fatalf("WorktreesByRepo errored on a prunable worktree: %v", err)
		}
		if got := len(m[repoName]); got != 1 {
			t.Fatalf("got %d worktrees, want 1 (prunable skipped): %+v", got, m[repoName])
		}
	})
}
