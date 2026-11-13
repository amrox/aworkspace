package workspace

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const workspaceRootMetaDirName = ".aworkspace"
const workspaceRootMetaFile = "meta.toml"
const workspaceMetaFile = ".aworkspace.toml"
const workspaceMetaComment = "# managed by aworkspace — do not edit directly\n\n"
const archiveDirName = "_archive"

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

func archiveDir(config Config) string {
	return filepath.Join(config.WorkspacesDir, archiveDirName)
}

func ensureArchiveDir(config Config) (string, error) {
	d := archiveDir(config)
	if err := os.MkdirAll(d, 0755); err != nil {
		return "", err
	}
	return d, nil
}

func writeMetadataFile(path string, data any) error {

	var buf bytes.Buffer
	buf.WriteString(workspaceMetaComment)

	err := toml.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func (w *Workspace) writeMetadata() error {

	return writeMetadataFile(filepath.Join(w.Path, workspaceMetaFile), w.Metadata)
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

var workspaceNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

func validateWorkspaceName(name string) error {

	if !workspaceNameRe.MatchString(name) {
		return fmt.Errorf("invalid workspace name - must begin with alphanumeric, and contain only alphanumeric, - and _")
	}
	return nil
}

func CreateWorkspace(name string, config Config) (Workspace, error) {

	err := validateWorkspaceName(name)
	if err != nil {
		return Workspace{}, err
	}

	wsPath := filepath.Join(config.WorkspacesDir, name)

	_, err = os.Stat(wsPath)
	if err == nil {
		return Workspace{}, fmt.Errorf("could not create directory %q: %w", wsPath, os.ErrExist)
	}
	if !os.IsNotExist(err) {
		return Workspace{}, err
	}

	allExisting, err := existingWorkspaceNames(config)
	if err != nil {
		return Workspace{}, err
	}
	for _, existing := range allExisting {
		if strings.EqualFold(name, existing) {
			return Workspace{}, fmt.Errorf("could not create workspace %q: workspace name collision with %q", name, existing)
		}
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

func existingWorkspaceNames(config Config) ([]string, error) {

	entries, err := os.ReadDir(config.WorkspacesDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func ListWorkspaces(config Config) ([]Workspace, error) {

	names, err := existingWorkspaceNames(config)
	if err != nil {
		return []Workspace{}, err
	}

	var workspaces []Workspace
	for _, e := range names {
		wsPath := filepath.Join(config.WorkspacesDir, e)
		ws, err := LoadWorkspace(wsPath)
		if err != nil {
			LogWarning("Could not load workspace at %v\n-> %v\n", wsPath, err)
		} else {
			workspaces = append(workspaces, ws)
		}
	}
	return workspaces, nil
}

func defaultBranchName(workspace Workspace, config Config) string {
	return config.BranchPrefix + workspace.Name()
}

func (ws *Workspace) worktreePath(repoName string) string {
	return filepath.Join(ws.Path, ws.Metadata.Root.WorktreeSubdir, repoName)
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

	bareRepoPath, err := ensureBareRepo(repoURL, config)
	if err != nil {
		return err
	}

	if branch == "" {
		branch = defaultBranchName(*ws, config)
	}

	// TODO: handle conflicting worktree paths
	workTreeDest := ws.worktreePath(repo.Name)
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

	wsRootMetaPath := filepath.Join(wsRootMetaDirPath, workspaceRootMetaFile)
	return writeMetadataFile(wsRootMetaPath, meta)
}

type ArchiveOptions struct {
	Force bool
}

func (ws *Workspace) Archive(archiveOptions ArchiveOptions, config Config) (string, error) {

	// -- Step 0: Dirty Gate

	var dirtyRepos []string
	for name := range ws.Metadata.Repos {
		dirty, err := isWorktreeDirty(ws.worktreePath(name), config)
		if err != nil {
			return "", err
		}

		if dirty {
			dirtyRepos = append(dirtyRepos, name)
		}
	}

	if len(dirtyRepos) > 0 && !archiveOptions.Force {
		return "", fmt.Errorf("dirty worktrees (use --force): %s", strings.Join(dirtyRepos, ", "))
	}

	// -- Step 1: Capture state

	worktreeSubdir := ws.Metadata.Root.WorktreeSubdir

	// -- Step 2: Move workspace to _archive dir

	archiveDir, err := ensureArchiveDir(config)
	if err != nil {
		return "", err
	}
	date := time.Now().Format("2006-01-02")
	dest := filepath.Join(archiveDir, fmt.Sprintf("%s@%s", ws.Name(), date))

	for i := 2; ; i++ { // collision handling
		_, err := os.Stat(dest)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return "", err
		}
		dest = filepath.Join(archiveDir, fmt.Sprintf("%s@%s-%d", ws.Name(), date, i))
	}

	if err := os.Rename(ws.Path, dest); err != nil {
		return "", err
	}

	// -- Step 3: drop worktrees from the archive, clear bare bookkeeping
	if err := os.RemoveAll(filepath.Join(dest, worktreeSubdir)); err != nil {
		return "", err
	}

	for _, rc := range ws.Metadata.Repos {
		repo, err := parseRepoURL(rc.URL)
		if err != nil {
			return "", err
		}
		if _, err := execGitCommand(config, "-C", repo.bareRepoPath(config), "worktree", "prune"); err != nil {
			return "", err
		}
	}

	return dest, nil
}
