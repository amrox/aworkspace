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
		if ws.Path != expectedPath {
			t.Errorf("expected ws.path %v, got: %v:", ws.Path, expectedPath)
		}

		expectedFiles := []string{"workspace.toml", "CLAUDE.md", "WORKSPACE.md"}

		for _, f := range expectedFiles {
			path := filepath.Join(ws.Path, f)
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

		if ws.Path != "" {
			t.Errorf("expected empty Workspace struct got: %v:", ws)
		}
	})
}

func TestLoadWorkspace(t *testing.T) {
	t.Run("load workspace dir does not exist", func(t *testing.T) {

		dir := t.TempDir()
		wsName := "abc"

		wsPath := filepath.Join(dir, wsName)

		_, err := LoadWorkspace(wsPath)

		if err == nil {
			t.Errorf("expected error when loading workspace at non-exist path %v", wsPath)
		}

	})

	t.Run("load workspace workspace.toml does not exist", func(t *testing.T) {

		dir := t.TempDir()
		wsName := "abc"

		wsPath := filepath.Join(dir, wsName)
		os.MkdirAll(wsPath, 0755)

		_, err := LoadWorkspace(wsPath)

		if err == nil {
			t.Errorf("expected error when loading workspace without workspace.toml at path %v", wsPath)
		}
	})

	t.Run("load workspace happy path", func(t *testing.T) {

		dir := t.TempDir()
		wsName := "abc"

		wsPath := filepath.Join(dir, wsName)
		os.MkdirAll(wsPath, 0755)
		os.WriteFile(filepath.Join(wsPath, "workspace.toml"), []byte{}, 0644)

		ws, err := LoadWorkspace(wsPath)
		if err != nil {
			t.Fatalf("LoadWorkspace returned unexpected error: %v:", err)
		}
		if ws.Path != wsPath {
			t.Errorf("expected ws.path = %v, got %v", wsPath, ws.Path)
		}
	})

	t.Run("create and load", func(t *testing.T) {

		tmpDir := t.TempDir()
		parentDir := filepath.Join(tmpDir, "Workspaces")
		config := DefaultConfig()
		config.WorkspacesDir = parentDir

		ws, err := CreateWorkspace("ws1", config)
		if err != nil {
			t.Fatalf("CreateWorkspace expected no error, got %v", err)
		}
		if ws.Metadata.Config.WorktreeSubdir != config.WorktreeSubdir {
			t.Errorf("WorktreeSubdir expected %v, got %v", config.WorktreeSubdir, ws.Metadata.Config.WorktreeSubdir)
		}

		ws2, err := LoadWorkspace(ws.Path)
		if err != nil {
			t.Fatalf("LoadWorkspace expected no error, got %v", err)
		}
		if ws.Path != ws2.Path {
			t.Errorf("expected workspace paths to be equal, ws.path: %v   ws2.path: %v", ws.Path, ws2.Path)
		}
		if ws.Metadata.Config.WorktreeSubdir != ws2.Metadata.Config.WorktreeSubdir {
			t.Errorf("expected worktreeSubdir config to be equal, ws..WorktreeSubdir: %v   ws2..WorktreeSubdir: %v", ws.Metadata.Config.WorktreeSubdir, ws2.Metadata.Config.WorktreeSubdir)
		}
	})
}

func TestListWorkspaces(t *testing.T) {
	t.Run("workspace parent does not exist", func(t *testing.T) {

		tmpDir := t.TempDir()
		parentDir := filepath.Join(tmpDir, "Workspaces")

		config := Config{WorkspacesDir: parentDir}

		list, err := ListWorkspaces(config)
		if err != nil  {
			t.Errorf("ListWorkspaces at non-existent parent path %v expect no error, got %v", parentDir, err)
		}
		if list != nil {
			t.Errorf("ListWorkspaces at non-existent parent path %v empty list, got %v", parentDir, list)
		}
	})

	t.Run("workspace parent dir exists, no children", func(t *testing.T) {

		tmpDir := t.TempDir()
		parentDir := filepath.Join(tmpDir, "Workspaces")

		os.MkdirAll(parentDir, 0755)

		config := Config{WorkspacesDir: parentDir}

		list, err := ListWorkspaces(config)
		if err != nil  {
			t.Errorf("ListWorkspaces at empty parent path %v expect no error, got %v", parentDir, err)
		}
		if list != nil {
			t.Errorf("ListWorkspaces at empty path %v empty list, got %v", parentDir, list)
		}
	})

	t.Run("good and bad workspace", func(t *testing.T) {

		tmpDir := t.TempDir()
		parentDir := filepath.Join(tmpDir, "Workspaces")
		config := Config{WorkspacesDir: parentDir}

		ws1Dir := filepath.Join(parentDir, "ws1")
		ws2Dir := filepath.Join(parentDir, "ws2")

		os.MkdirAll(ws1Dir, 0755)
		os.MkdirAll(ws2Dir, 0755)

		os.WriteFile(filepath.Join(ws1Dir, "workspace.toml"), []byte{}, 0644)

		list, err := ListWorkspaces(config)
		if err != nil  {
			t.Fatalf("ListWorkspaces expected no error, got %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 workspace entry, got %d", len(list))
		}
		if list[0].Path != ws1Dir {
			t.Errorf("expected workspace with path %v , got %v", ws1Dir, list[0].Path)
		}
	})
}
