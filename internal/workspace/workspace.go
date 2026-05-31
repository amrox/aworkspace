package workspace

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const workspaceRootMetaDirName = ".aworkspace"
const workspaceRootMetaFile = "meta.toml"
const workspaceMetaFile = ".aworkspace.toml"

//go:embed templates/CLAUDE.md
var claudeMdTemplate string

//go:embed templates/WORKSPACE.md
var workspaceMdTemplate string

type RepoConfig struct {
	URL string `toml:"url"`
}

type WorkspaceRootMetadata struct {
	WorktreeSubdir string `toml:"worktree_subdir"`
}

type WorkspaceMetadata struct {
	Repos map[string]RepoConfig  `toml:"repos"`
	Root  *WorkspaceRootMetadata `toml:"-"`
}

func FindWorkspaceDir(startDir string) (string, error) {
	dir := startDir

	for {
		_, err := os.Stat(filepath.Join(dir, workspaceMetaFile))
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

func (w *Workspace) writeMetadata() error {
	bytes, err := toml.Marshal(w.Metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(w.Path, workspaceMetaFile), bytes, 0666)
}

func loadRootMetadata(workspaceRootPath string) (WorkspaceRootMetadata, error) {
	wsRootMetaPath := filepath.Join(workspaceRootPath, workspaceRootMetaDirName, workspaceRootMetaFile)
	bytes, err := os.ReadFile(wsRootMetaPath)
	if err != nil {
		return WorkspaceRootMetadata{}, fmt.Errorf("could not find workspace root metadata: %w", err)
	}
	var wsRootMeta WorkspaceRootMetadata
	err = toml.Unmarshal(bytes, &wsRootMeta)
	if err != nil {
		return WorkspaceRootMetadata{}, err
	}
	return wsRootMeta, nil
}

func (w *Workspace) loadMetadata(wsRootMeta WorkspaceRootMetadata) error {

	bytes, err := os.ReadFile(filepath.Join(w.Path, workspaceMetaFile))
	if err != nil {
		return err
	}

	err = toml.Unmarshal(bytes, &w.Metadata)
	if err != nil {
		return err
	}

	w.Metadata.Root = &wsRootMeta

	return nil
}

func NewWorkspace(path string) Workspace {
	return Workspace{
		Path: path,
		Metadata: WorkspaceMetadata{
			Repos: make(map[string]RepoConfig),
		},
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
			_ = os.RemoveAll(wsPath)
		}
	}()

	err = os.MkdirAll(wsPath, 0777)
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

	ws := NewWorkspace(wsPath)
	err = ws.writeMetadata()
	if err != nil {
		return Workspace{}, err
	}

	success = true
	return ws, nil
}

func LoadWorkspace(wsPath string) (Workspace, error) {

	_, err := os.Stat(wsPath)
	if err != nil {
		return Workspace{}, fmt.Errorf("workspace path %q not found: %w", wsPath, err)
	}

	wsRootPath := filepath.Join(wsPath, "..")
	wsRootMeta, err := loadRootMetadata(wsRootPath)
	if err != nil {
		return Workspace{}, err
	}

	ws := NewWorkspace(wsPath)
	err = ws.loadMetadata(wsRootMeta)
	if err != nil {
		return Workspace{}, err
	}

	return ws, nil
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
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			wsPath := filepath.Join(config.WorkspacesDir, e.Name())
			ws, err := LoadWorkspace(wsPath)
			if err != nil {
				LogWarning("Could not load workspace at: %v\n", wsPath)
			} else {
				workspaces = append(workspaces, ws)
			}
		}
	}
	return workspaces, nil
}

func defaultBranchName(workspace Workspace, config Config) string {
	return config.BranchPrefix + workspace.Name()
}

func (ws *Workspace) AddRepo(repoURL string, branch string, config Config) error {

	for k := range ws.Metadata.Repos {
		if ws.Metadata.Repos[k].URL == repoURL {
			return fmt.Errorf("repo with URL %v already added to workspace", repoURL)
		}
	}

	repo, err := parseRepoURL(repoURL)
	if err != nil {
		return err
	}

	bareRepoPath, err := cloneBareRepo(repoURL, config)
	if err != nil {
		return err
	}

	if branch == "" {
		branch = defaultBranchName(*ws, config)
	}

	// TODO: handle conflicting worktree paths
	workTreeDest := filepath.Join(ws.Path, ws.Metadata.Root.WorktreeSubdir, repo.Name)
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(workTreeDest)
		}
	}()

	err = addWorktree(bareRepoPath, workTreeDest, branch, config)
	if err != nil {
		return err
	}

	ws.Metadata.Repos[repo.Name] = RepoConfig{URL: repo.Url}
	err = ws.writeMetadata()
	if err != nil {
		return err
	}

	success = true
	return nil
}

func CreateWorkspaceRoot(path string, config Config) error {

	fileInfo, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// create directory
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}

	} else if !fileInfo.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)

	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory %s is not empty", path)
		}
	}

	wsRootMetaDirPath := filepath.Join(path, workspaceRootMetaDirName)
	err = os.MkdirAll(wsRootMetaDirPath, 0755)
	if err != nil {
		return err
	}

	meta := WorkspaceRootMetadata{
		WorktreeSubdir: config.WorktreeSubdir,
	}

	bytes, err := toml.Marshal(meta)
	if err != nil {
		return err
	}

	wsRootMetaPath := filepath.Join(wsRootMetaDirPath, workspaceRootMetaFile)
	err = os.WriteFile(wsRootMetaPath, bytes, 0666)
	if err != nil {
		return err
	}
	return nil
}
