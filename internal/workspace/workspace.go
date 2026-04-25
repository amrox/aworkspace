package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

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

func CreateWorkspaceParentDir(config Config) error {
	fileinfo, err := os.Stat(config.WorkspacesDir)
	if (err != nil) {
		return err
	}
	if !fileinfo.IsDir() {

	}
	return nil
}

func CreateWorkspace(name string, config Config) (Workspace, error) {

	var err error
	wsPath := filepath.Join(config.WorkspacesDir, name)

	_, err = os.Stat(wsPath)
	if err == nil {
		return Workspace{}, fmt.Errorf("directory %q already exists: %w", wsPath, os.ErrExist)
	}
	if !os.IsNotExist(err) {
		return Workspace{}, err
	}

	err = os.MkdirAll(wsPath, 0777)
	if err != nil {
		return Workspace{}, err
	}

	var ws Workspace
	ws.path = wsPath

	return ws, nil
}