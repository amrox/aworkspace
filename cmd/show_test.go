package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amrox/aworkspace/internal/workspace"
)

func TestShowWithDirFlag(t *testing.T) {
	// Create a temp workspace
	dir := t.TempDir()
	t.Log(dir)
	config := workspace.Config{WorkspacesDir: dir, WorktreeSubdir: "code"}
	err := workspace.CreateWorkspaceRoot(dir, config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = workspace.CreateWorkspace("test-ws", config)
	if err != nil {
		t.Fatal(err)
	}

	wsPath := filepath.Join(dir, "test-ws")

	cmd := rootCmd
	cmd.SetArgs([]string{"show", "-C", wsPath})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("show -C failed: %v", err)
	}

	if !strings.Contains(out.String(), "test-ws") {
		t.Errorf("expected workspace name in output, got: %s", out.String())
	}
}
