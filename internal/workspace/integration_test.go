//go:build integration

package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func init() {
	CurLogLevel = LogLevelVerbose
}

func createTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	exec.Command("git", "-C", dir, "init").Run()
	exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init").Run()

	return dir
}

func createTestWorkspace(t *testing.T, name string) (Workspace, Config) {
	t.Helper()

	config := Config{
		WorkspacesDir:  t.TempDir(),
		BaresDir:       t.TempDir(),
		WorktreeSubdir: "code",
		BranchPrefix:   "ws/",
	}

	ws, err := CreateWorkspace(name, config)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	return ws, config
}

func verifyWorktree(t *testing.T, ws Workspace, config Config, repoName string) {
	t.Helper()

	worktreeGitPath := filepath.Join(ws.Path, config.WorktreeSubdir, repoName, ".git")
	fileInfo, err := os.Stat(worktreeGitPath)
	if err != nil {
		t.Fatalf("check worktree: %v", err)
	}
	if fileInfo.IsDir() {
		t.Fatalf("check worktree: expect '.git' file, found %v", fileInfo)
	}
}

func TestAddRepoIntegration(t *testing.T) {

	t.Run("happy path - no existing bare", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := createTestRepo(t)
		err := ws.AddRepo(remoteRepo, "", config)
		if err != nil {
			t.Fatalf("AddRepo: %v", err)
		}

		// verify bare now exists
		barePath := filepath.Join(config.BaresDir, "_", remoteRepo)
		t.Log(barePath)
		fileInfo, err := os.Stat(barePath)
		if err != nil {
			t.Fatalf("check bare clone: %v", err)
		}
		if !fileInfo.IsDir() {
			t.Fatalf("check bare clone: expected '%v' dir, found %v", barePath, fileInfo)
		}

		remoteRepoName := filepath.Base(remoteRepo)
		verifyWorktree(t, ws, config, remoteRepoName)
	})

	t.Run("happy path - with existing bare", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := createTestRepo(t)

		// pre-add bare clone
		_, err := cloneBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("cloneBareRepo: %v", err)
		}

		err = ws.AddRepo(remoteRepo, "", config)
		if err != nil {
			t.Fatalf("AddRepo: %v", err)
		}

		remoteRepoName := filepath.Base(remoteRepo)
		verifyWorktree(t, ws, config, remoteRepoName)
	})

	t.Run("happy path - with existing remote branch", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := createTestRepo(t)

		// create a branch matching our ws branch
		err := exec.Command("git", "-C", remoteRepo, "branch", defaultBranchName(ws, config)).Run()
		if err != nil {
			t.Fatalf("git remote branch creation failed: %v", err)
		}

		err = ws.AddRepo(remoteRepo, "", config)
		if err != nil {
			t.Fatalf("AddRepo: %v", err)
		}

		remoteRepoName := filepath.Base(remoteRepo)
		verifyWorktree(t, ws, config, remoteRepoName)
	})

	t.Run("happy path - with existing local branch", func(t *testing.T) {

		// this is a weird case - existing branch with no existing worktree - but it should succeed

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := createTestRepo(t)

		// pre-add bare clone
		localBare, err := cloneBareRepo(remoteRepo, config)

		// create a branch matching our ws branch
		err = exec.Command("git", "-C", localBare, "branch", defaultBranchName(ws, config)).Run()
		if err != nil {
			t.Fatalf("git local branch creation failed: %v", err)
		}

		err = ws.AddRepo(remoteRepo, "", config)
		if err != nil {
			t.Fatalf("AddRepo: %v", err)
		}

		remoteRepoName := filepath.Base(remoteRepo)
		verifyWorktree(t, ws, config, remoteRepoName)
	})

	t.Run("failure - bogus remote", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := t.TempDir()

		err := ws.AddRepo(remoteRepo, "", config)
		if err == nil {
			t.Fatalf("AddRepo: expected error, got <nil>")
		}
	})

	t.Run("failure - existing worktree", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := createTestRepo(t)

		// pre-add bare clone
		localBare, err := cloneBareRepo(remoteRepo, config)

		// create a worktree with branch matching our ws branch
		worktreeDest := t.TempDir()
		err = exec.Command("git", "-C", localBare, "worktree", "add", worktreeDest, "-B", defaultBranchName(ws, config)).Run()
		if err != nil {
			t.Fatalf("git worktree creation failed: %v", err)
		}

		err = ws.AddRepo(remoteRepo, "", config)
		if err == nil {
			t.Fatalf("AddRepo: expected error, got <nil>")
		}
	})
}