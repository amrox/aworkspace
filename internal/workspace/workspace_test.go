package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindWorkspaceDir(t *testing.T) {
	t.Run("finds workspace in current dir", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "workspace.toml"), []byte{}, 0644)

		got, err := FindWorkspaceDir(dir)

		if err != nil {
			t.Fatalf("FindWorkspaceDir returned unexpected error: %v:", err)
		}

		if got != dir {
			t.Errorf("FindWorkspacesDir = %v, want %v", got, dir)
		}
	})

	t.Run("finds workspace in parent dir", func(t *testing.T) {
		root := t.TempDir()
		os.WriteFile(filepath.Join(root, "workspace.toml"), []byte{}, 0644)
		child := filepath.Join(root, "code", "some-repo")
		os.MkdirAll(child, 0755)

		got, err := FindWorkspaceDir(child)

		if err != nil {
			t.Fatalf("FindWorkspaceDir returned unexpected error: %v:", err)
		}

		if got != root {
			t.Errorf("FindWorkspacesDir = %v, want %v", got, root)
		}
	})

	t.Run("returns error when not in workspace", func(t *testing.T) {

		dir := t.TempDir() // assume no workspace.toml anywhere above

		_, err := FindWorkspaceDir(dir)

		if !errors.Is(err, ErrNotInWorkspace) {
			t.Errorf("expected ErrNotInWorkspace, got: %v:", err)
		}
	})
}

func TestCreateWorkspace(t *testing.T) {
	t.Run("create a new workspace no errors", func(t *testing.T) {

		dir := t.TempDir()
		wsName := "abc"

		var config Config
		config.WorkspacesDir = dir

		ws, err := CreateWorkspace(wsName, config)

		if err != nil {
			t.Fatalf("CreateWorkspace returned unexpected error: %v:", err)
		}

		expectedPath := filepath.Join(config.WorkspacesDir, wsName)
		if ws.path != expectedPath {
			t.Errorf("expected ws.path %v, got: %v:", ws.path, expectedPath)
		}

		expectedFiles := []string{"workspace.toml", "CLAUDE.md", "WORKSPACE.md"}

		for _, f := range expectedFiles {
			path := filepath.Join(ws.path, f)
			_, err = os.Stat(path)
			if err != nil {
				t.Errorf("CreateWorkspace %v not found: %v", f, err)
			}
		}
	})

	t.Run("create workspace when destination exists", func(t *testing.T) {

		dir := t.TempDir()
		wsName := "abc"

		var config Config
		config.WorkspacesDir = dir

		err := os.MkdirAll(filepath.Join(dir, wsName), 0777)
		if err != nil {
			t.Fatalf("os.MkdirAll returned unexpected error: %v:", err)
		}

		var ws Workspace
		ws, err = CreateWorkspace(wsName, config)
		if !errors.Is(err, os.ErrExist) {
			t.Errorf("expected os.ErrExists, got: %v:", err)
		}

		if ws != (Workspace{}) {
			t.Errorf("expected empty Workspace struct got: %v:", ws)
		}
	})
}
