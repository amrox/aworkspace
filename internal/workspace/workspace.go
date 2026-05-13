package workspace

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed templates/CLAUDE.md
var claudeMdTemplate string

//go:embed templates/WORKSPACE.md
var workspaceMdTemplate string

func FindWorkspaceDir(startDir string) (string, error) {
	dir := startDir

	for {
		_, err := os.Stat(filepath.Join(dir, "workspace.toml"))
		if err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		nextDir := filepath.Dir(dir)
		if dir == nextDir {
			return "", ErrNotInWorkspace
		}
		dir = nextDir
	}
}

func CreateWorkspace(name string, config Config) (Workspace, error) {

	var err error
	wsPath := filepath.Join(config.WorkspacesDir, name)

	_, err = os.Stat(wsPath)
	if err == nil {
		return Workspace{}, fmt.Errorf("could not create directory %q: %w", wsPath, os.ErrExist)
	}
	if !os.IsNotExist(err) {
		return Workspace{}, err
	}

	success := false
	defer func() {
		if !success {
			os.RemoveAll(wsPath)
		}
	}()

	err = os.MkdirAll(wsPath, 0777)
	if err != nil {
		return Workspace{}, err
	}

	err = os.WriteFile(filepath.Join(wsPath, "workspace.toml"), []byte{}, 0666)
	if err != nil {
		return Workspace{}, err
	}

	err = os.WriteFile(filepath.Join(wsPath, "CLAUDE.md"), []byte(claudeMdTemplate), 0666)
	if err != nil {
		return Workspace{}, err
	}

	err = os.WriteFile(filepath.Join(wsPath, "WORKSPACE.md"), []byte(workspaceMdTemplate), 0666)
	if err != nil {
		return Workspace{}, err
	}

	success = true
	return Workspace{Path: wsPath}, nil
}

func LoadWorkspace(wsPath string) (Workspace, error) {

	_, err := os.Stat(wsPath)
	if err != nil {
		return Workspace{}, fmt.Errorf("workspace path %q not found: %w", wsPath, err)
	}

	wsTomlPath := filepath.Join(wsPath, "workspace.toml")
	_, err = os.Stat(wsTomlPath)
	if err != nil {
		 return Workspace{}, fmt.Errorf("%q is not a workspace (missing workspace.toml): %w", wsPath, err)
	}

	return Workspace{Path: wsPath}, nil
}

func ListWorkspaces(config Config) ([]Workspace, error) {

	entries, err := os.ReadDir(config.WorkspacesDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var workspaces []Workspace

	for _, e := range entries {
		if e.IsDir() {
		wsPath := filepath.Join(config.WorkspacesDir, e.Name())
			ws, err := LoadWorkspace(wsPath)
			if err != nil {
				// TODO: logging
			} else {
				workspaces = append(workspaces, ws)
			}
		}
	}
	return workspaces, nil
}

func DefaultBranchName(workspace Workspace, config Config) string {
	return config.BranchPrefix + workspace.Name()
}

func AddRepo(ws Workspace, repoURL string, branch string, config Config) error {

	repo, err := parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	bareRepoPath, err := cloneBareRepo(repoURL, config)
	if err != nil {
		return err
	}

	if branch == "" {
		branch = DefaultBranchName(ws, config)
	}

	// TODO: handle conflicting worktree paths
	workTreeDest := filepath.Join(ws.Path, config.WorktreeSubdir, repo.Name)

	err = addWorktree(bareRepoPath, workTreeDest, branch)
	if err != nil {
		return err
	}

	// TODO: add to workspace.toml

	return nil
}