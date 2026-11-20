//go:build integration

package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amrox/aworkspace/internal/workspace"
)

func gitInitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		full := append([]string{"-C", dir}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("-c", "user.email=a@b.c", "-c", "user.name=x", "commit", "--allow-empty", "-m", "init")
	return dir
}

// setupWSWithRepo creates a workspaces root + workspace and adds one repo
// (creating its worktree). Returns (wsPath, repoName, config).
func setupWSWithRepo(t *testing.T, wsName string) (string, string, workspace.Config) {
	t.Helper()
	workspacesDir := t.TempDir()
	config := workspace.DefaultConfig()
	config.WorkspacesDir = workspacesDir
	config.BaresDir = t.TempDir()
	if err := workspace.CreateWorkspaceRoot(workspacesDir, config); err != nil {
		t.Fatalf("CreateWorkspaceRoot: %v", err)
	}
	ws, err := workspace.CreateWorkspace(wsName, config)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	ws, err = workspace.LoadWorkspace(ws.Path)
	if err != nil {
		t.Fatalf("LoadWorkspace: %v", err)
	}
	repo := gitInitRepo(t)
	if err := ws.AddRepo(repo, config); err != nil {
		t.Fatalf("AddRepo: %v", err)
	}
	return ws.Path, filepath.Base(repo), config
}

func runRmRepo(t *testing.T, config workspace.Config, args ...string) (string, error) {
	t.Helper()
	t.Cleanup(func() { cfg = nil })
	cfg = &config
	cmd := rootCmd
	cmd.SetArgs(append([]string{"remove-repo"}, args...))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	return out.String(), err
}

func TestRmRepoCmd(t *testing.T) {

	t.Run("requires a repo name", func(t *testing.T) {
		wsPath, _, config := setupWSWithRepo(t, "feature1")
		_, err := runRmRepo(t, config, "-C", wsPath)
		if err == nil {
			t.Fatal("expected error when no repo name given, got nil")
		}
	})

	t.Run("refuses from inside the worktree", func(t *testing.T) {
		wsPath, repoName, config := setupWSWithRepo(t, "feature1")
		inside := filepath.Join(wsPath, config.WorktreeSubdir, repoName)
		_, err := runRmRepo(t, config, repoName, "-C", inside)
		if err == nil {
			t.Fatal("expected refusal from inside the worktree, got nil")
		}
		if !strings.Contains(err.Error(), "inside") {
			t.Errorf("expected 'inside' in error, got: %v", err)
		}
		if _, serr := os.Stat(inside); serr != nil {
			t.Errorf("expected worktree to remain after refused remove: %v", serr)
		}
	})

	t.Run("removes the repo from outside the worktree", func(t *testing.T) {
		wsPath, repoName, config := setupWSWithRepo(t, "feature1")
		// cwd = workspace root: not inside the worktree -> allowed
		out, err := runRmRepo(t, config, repoName, "-C", wsPath)
		if err != nil {
			t.Fatalf("rm-repo failed: %v", err)
		}
		if !strings.Contains(out, "Removed repo "+repoName) {
			t.Errorf("expected removal output, got: %q", out)
		}
		wt := filepath.Join(wsPath, config.WorktreeSubdir, repoName)
		if _, serr := os.Stat(wt); !os.IsNotExist(serr) {
			t.Errorf("expected worktree gone, stat err = %v", serr)
		}
	})
}
