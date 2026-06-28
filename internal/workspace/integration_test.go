//go:build integration

package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	CurLogLevel = LogLevelVerbose
}

func createTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	exec.Command("git", "-C", dir, "init", "-b", "main").Run()
	exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init").Run()

	return dir
}

func createTestWorkspace(t *testing.T, name string) (Workspace, Config) {
	t.Helper()

	wsDir := t.TempDir()

	config := Config{
		WorkspacesDir:  wsDir,
		BaresDir:       t.TempDir(),
		WorktreeSubdir: "code",
		BranchPrefix:   "", // docs: default is no prefix (branch name = workspace name)
	}

	err := CreateWorkspaceRoot(wsDir, config)
	if err != nil {
		t.Fatalf("CreateWorkspaceRoot: %v", err)
	}

	ws, err := CreateWorkspace(name, config)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Load to pick up root metadata
	ws, err = LoadWorkspace(ws.Path)
	if err != nil {
		t.Fatalf("LoadWorkspace: %v", err)
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
		_, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("ensureBareRepo: %v", err)
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
		localBare, err := ensureBareRepo(remoteRepo, config)

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

	t.Run("errors when branch is checked out in another worktree", func(t *testing.T) {

		// Per docs: when the workspace branch already exists AND is checked out in
		// another worktree (git forbids sharing a branch across worktrees), AddRepo
		// does NOT auto-rename — it errors. This is an exceptional, usually-dirty
		// state; aworkspace shouts rather than silently inventing a new branch name.

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := createTestRepo(t)

		// pre-add bare clone
		localBare, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("ensureBareRepo: %v", err)
		}

		// occupy the workspace branch with a worktree elsewhere so git refuses to reuse it
		branch := defaultBranchName(ws, config)
		occupied := t.TempDir()
		err = exec.Command("git", "-C", localBare, "worktree", "add", occupied, "-B", branch).Run()
		if err != nil {
			t.Fatalf("git worktree creation failed: %v", err)
		}

		err = ws.AddRepo(remoteRepo, "", config)
		if err == nil {
			t.Fatalf("AddRepo: expected error (branch %q checked out elsewhere), got <nil>", branch)
		}

		// and it must not leave a half-created worktree behind
		repoName := filepath.Base(remoteRepo)
		worktreePath := filepath.Join(ws.Path, config.WorktreeSubdir, repoName)
		if _, statErr := os.Stat(worktreePath); statErr == nil {
			t.Errorf("AddRepo left a worktree at %q after erroring", worktreePath)
		}
	})

	t.Run("ensureBareRepo set refspec", func(t *testing.T) {
		remoteRepo := createTestRepo(t)
		config := Config{
			BaresDir: t.TempDir(),
		}

		barePath, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("ensureBareRepo: %v", err)
		}

		out, err := exec.Command("git", "-C", barePath, "config", "--get", "remote.origin.fetch").Output()
		if err != nil {
			t.Fatalf("git config: %v", err)
		}
		got := strings.TrimSpace(string(out))
		want := "+refs/heads/*:refs/remotes/origin/*"
		if got != want {
			t.Fatalf("refspace = %q, want %q", got, want)
		}
	})

	t.Run("ensureBareRepo idempotent", func(t *testing.T) {
		remoteRepo := createTestRepo(t)
		config := Config{
			BaresDir: t.TempDir(),
		}

		barePath1, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("first call: %v", err)
		}

		barePath2, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("second call: %v", err)
		}

		if barePath1 != barePath2 {
			t.Fatalf("paths differ: %q vs %q", barePath1, barePath2)
		}
	})

	t.Run("ensureBareRepo fixes wrong refspec", func(t *testing.T) {
		remoteRepo := createTestRepo(t)
		config := Config{
			BaresDir: t.TempDir(),
		}

		barePath, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("ensureBareRepo: %v", err)
		}

		// sabotage the refspec
		exec.Command("git", "-C", barePath, "config", "remote.origin.fetch", "+refs/heads/*:refs/heads/*").Run()

		// re-ensure should fix it
		_, err = ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("re-ensure: %v", err)
		}

		out, _ := exec.Command("git", "-C", barePath, "config", "--get", "remote.origin.fetch").Output()
		got := strings.TrimSpace(string(out))
		want := "+refs/heads/*:refs/remotes/origin/*"
		if got != want {
			t.Fatalf("refspec = %q, want %q", got, want)
		}
	})

	t.Run("fetch after refspec fix creates remote tracking refs", func(t *testing.T) {
		remoteRepo := createTestRepo(t)
		config := Config{
			BaresDir: t.TempDir(),
		}

		barePath, err := ensureBareRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("ensureBareRepo: %v", err)
		}

		err = exec.Command("git", "-C", barePath, "fetch", "origin").Run()
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}

		// verify origin/master or origin/main exists
		err = exec.Command("git", "-C", barePath, "rev-parse", "--verify", "refs/remotes/origin/main").Run()
		if err != nil {
			t.Fatalf("remote tracking ref not found: %v", err)
		}
	})
}
