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
		err := ws.AddRepo(remoteRepo, config)
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

		err = ws.AddRepo(remoteRepo, config)
		if err != nil {
			t.Fatalf("AddRepo: %v", err)
		}

		remoteRepoName := filepath.Base(remoteRepo)
		verifyWorktree(t, ws, config, remoteRepoName)
	})

	t.Run("worktree is detached at the default branch", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := createTestRepo(t)
		if err := ws.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo: %v", err)
		}

		repoName := filepath.Base(remoteRepo)
		verifyWorktree(t, ws, config, repoName)

		wt := ws.WorktreePath(repoName)

		// HEAD is detached — no branch is created under the no-branch model
		if err := exec.Command("git", "-C", wt, "symbolic-ref", "-q", "HEAD").Run(); err == nil {
			t.Errorf("expected detached HEAD in worktree, but HEAD is on a branch")
		}

		// ...and it points at the bare's default branch commit
		bare := filepath.Join(config.BaresDir, "_", remoteRepo)
		defaultRef, err := getDefaultBranch(bare, config)
		if err != nil {
			t.Fatalf("getDefaultBranch: %v", err)
		}
		wantOut, err := exec.Command("git", "-C", bare, "rev-parse", defaultRef).Output()
		if err != nil {
			t.Fatalf("rev-parse default: %v", err)
		}
		gotOut, err := exec.Command("git", "-C", wt, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatalf("rev-parse HEAD: %v", err)
		}
		if got, want := strings.TrimSpace(string(gotOut)), strings.TrimSpace(string(wantOut)); got != want {
			t.Errorf("worktree HEAD = %q, want default branch %q", got, want)
		}
	})

	t.Run("two workspaces can add the same repo", func(t *testing.T) {

		// No branch per worktree means git never refuses a shared branch — two
		// workspaces back onto the same bare, each detached at the default branch.

		ws1, config := createTestWorkspace(t, "ws-one")
		remoteRepo := createTestRepo(t)
		if err := ws1.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo ws1: %v", err)
		}

		ws2, err := CreateWorkspace("ws-two", config)
		if err != nil {
			t.Fatalf("CreateWorkspace ws2: %v", err)
		}
		ws2, err = LoadWorkspace(ws2.Path)
		if err != nil {
			t.Fatalf("LoadWorkspace ws2: %v", err)
		}
		if err := ws2.AddRepo(remoteRepo, config); err != nil {
			t.Fatalf("AddRepo ws2: %v", err)
		}

		repoName := filepath.Base(remoteRepo)
		verifyWorktree(t, ws1, config, repoName)
		verifyWorktree(t, ws2, config, repoName)
	})

	t.Run("failure - bogus remote", func(t *testing.T) {

		ws, config := createTestWorkspace(t, "test-ws")

		remoteRepo := t.TempDir()

		err := ws.AddRepo(remoteRepo, config)
		if err == nil {
			t.Fatalf("AddRepo: expected error, got <nil>")
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
