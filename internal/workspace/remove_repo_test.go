//go:build integration

package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// RemoveRepo detaches one repo's worktree from a live workspace and drops it
// from the record. Contract:
//   - known repo, clean     -> worktree gone, branch kept in bare, entry removed from .aworkspace.toml
//   - unknown repo          -> error, nothing touched
//   - dirty worktree, !Force -> error, worktree + entry intact
//   - dirty worktree, Force  -> removed

func TestRemoveRepoIntegration(t *testing.T) {

	t.Run("clean: removes worktree, keeps branch, drops metadata entry", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "feature1")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, "", config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)
		worktree := ws.WorktreePath(repoName)

		if _, err := os.Stat(worktree); err != nil {
			t.Fatalf("precondition: worktree should exist: %v", err)
		}

		if err := ws.RemoveRepo(repoName, RemoveRepoOptions{}, config); err != nil {
			t.Fatalf("RemoveRepo: %v", err)
		}

		// worktree gone
		if _, err := os.Stat(worktree); !os.IsNotExist(err) {
			t.Errorf("expected worktree %q gone, stat err = %v", worktree, err)
		}
		// branch kept in bare (parity with archive: removing a worktree never drops the branch)
		bare, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("ensureBareRepo: %v", err)
		}
		if err := exec.Command("git", "-C", bare, "rev-parse", "--verify", "feature1").Run(); err != nil {
			t.Errorf("expected branch 'feature1' to remain in bare: %v", err)
		}
		// entry gone from the persisted record
		reloaded, err := LoadWorkspace(ws.Path)
		if err != nil {
			t.Fatalf("LoadWorkspace: %v", err)
		}
		if _, ok := reloaded.Metadata.Repos[repoName]; ok {
			t.Errorf("expected %q removed from metadata, still present", repoName)
		}
	})

	t.Run("unknown repo: errors, touches nothing", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "feature2")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, "", config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)

		if err := ws.RemoveRepo("nonexistent", RemoveRepoOptions{}, config); err == nil {
			t.Fatal("expected error for unknown repo, got nil")
		}
		// the real repo's worktree + entry are untouched
		if _, err := os.Stat(ws.WorktreePath(repoName)); err != nil {
			t.Errorf("expected existing worktree untouched: %v", err)
		}
		reloaded, err := LoadWorkspace(ws.Path)
		if err != nil {
			t.Fatalf("LoadWorkspace: %v", err)
		}
		if _, ok := reloaded.Metadata.Repos[repoName]; !ok {
			t.Errorf("expected %q still in metadata after unrelated error", repoName)
		}
	})

	t.Run("dirty without --force: errors, worktree and entry intact", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "feature3")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, "", config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)
		addDirtyFile(t, ws, config, repoName)

		if err := ws.RemoveRepo(repoName, RemoveRepoOptions{}, config); err == nil {
			t.Fatal("expected error removing dirty worktree without --force, got nil")
		}
		if _, err := os.Stat(ws.WorktreePath(repoName)); err != nil {
			t.Errorf("expected dirty worktree to remain: %v", err)
		}
		reloaded, err := LoadWorkspace(ws.Path)
		if err != nil {
			t.Fatalf("LoadWorkspace: %v", err)
		}
		if _, ok := reloaded.Metadata.Repos[repoName]; !ok {
			t.Errorf("expected %q still in metadata after refused remove", repoName)
		}
	})

	t.Run("dirty with --force: removes", func(t *testing.T) {
		ws, config := createTestWorkspace(t, "feature4")
		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, "", config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}
		repoName := filepath.Base(remoteRepo)
		addDirtyFile(t, ws, config, repoName)

		if err := ws.RemoveRepo(repoName, RemoveRepoOptions{Force: true}, config); err != nil {
			t.Fatalf("RemoveRepo(force): %v", err)
		}
		if _, err := os.Stat(ws.WorktreePath(repoName)); !os.IsNotExist(err) {
			t.Errorf("expected worktree gone after forced remove, stat err = %v", err)
		}
	})
}
