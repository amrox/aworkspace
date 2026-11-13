package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amrox/aworkspace/internal/workspace"
)

// setupArchiveWS creates a workspaces root with one empty workspace (no repos,
// so archive needs no git) and returns (workspacesDir, wsPath, config).
func setupArchiveWS(t *testing.T, name string) (string, string, workspace.Config) {
	t.Helper()
	workspacesDir := t.TempDir()
	config := workspace.DefaultConfig()
	config.WorkspacesDir = workspacesDir
	if err := workspace.CreateWorkspaceRoot(workspacesDir, config); err != nil {
		t.Fatalf("CreateWorkspaceRoot: %v", err)
	}
	ws, err := workspace.CreateWorkspace(name, config)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	return workspacesDir, ws.Path, config
}

// runArchive invokes `archive` with the given args against config. The -C flag
// stands in for the invoking shell's cwd (getStartDir reads it).
func runArchive(t *testing.T, config workspace.Config, args ...string) (string, error) {
	t.Helper()
	t.Cleanup(func() { cfg = nil })
	cfg = &config
	cmd := rootCmd
	cmd.SetArgs(append([]string{"archive"}, args...))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	return out.String(), err
}

func TestArchiveCmd(t *testing.T) {

	t.Run("requires a workspace name (never infers)", func(t *testing.T) {
		workspacesDir, _, config := setupArchiveWS(t, "feature1")
		_, err := runArchive(t, config, "-C", workspacesDir)
		if err == nil {
			t.Fatal("expected error when no workspace name given, got nil")
		}
	})

	t.Run("refuses when cwd is the workspace dir", func(t *testing.T) {
		_, wsPath, config := setupArchiveWS(t, "feature1")
		_, err := runArchive(t, config, "feature1", "-C", wsPath)
		if err == nil {
			t.Fatal("expected refusal when run from the workspace dir, got nil")
		}
		if !strings.Contains(err.Error(), "inside") {
			t.Errorf("expected 'inside' in refusal error, got: %v", err)
		}
	})

	t.Run("refuses when cwd is under the workspace (a worktree subdir)", func(t *testing.T) {
		_, wsPath, config := setupArchiveWS(t, "feature1")
		inside := filepath.Join(wsPath, config.WorktreeSubdir, "somerepo")
		_, err := runArchive(t, config, "feature1", "-C", inside)
		if err == nil {
			t.Fatal("expected refusal when run from inside a worktree, got nil")
		}
	})

	t.Run("archives from outside the workspace", func(t *testing.T) {
		workspacesDir, wsPath, config := setupArchiveWS(t, "feature1")
		// cwd = the workspaces root: a parent of the workspace, not under it -> allowed
		_, err := runArchive(t, config, "feature1", "-C", workspacesDir)
		if err != nil {
			t.Fatalf("archive from outside failed: %v", err)
		}
		if _, err := os.Stat(wsPath); !os.IsNotExist(err) {
			t.Errorf("expected original workspace dir gone, stat err = %v", err)
		}
		entries, _ := os.ReadDir(filepath.Join(workspacesDir, "_archive"))
		found := false
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "feature1@") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected _archive/feature1@<date>, entries = %v", entries)
		}
	})
}
