package workspace

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
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

func (ws *Workspace) WorktreePath(repoName string) string {
	return filepath.Join(ws.Path, ws.Metadata.Root.WorktreeSubdir, repoName)
}

func (ws *Workspace) AddRepo(repoURL string, config Config) error {

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

	// TODO: handle conflicting worktree paths
	workTreeDest := ws.WorktreePath(repo.Name)
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(workTreeDest)
		}
	}()

	err = addWorktree(bareRepoPath, workTreeDest, config)
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

type RemoveRepoOptions struct {
	Force bool
}

func (ws *Workspace) RemoveRepo(repoName string, removeOpts RemoveRepoOptions, config Config) error {

	rc, ok := ws.Metadata.Repos[repoName]
	if !ok {
		return fmt.Errorf("repo %q not in workspace %q", repoName, ws.Name())
	}

	repo, err := parseRepoURL(rc.URL)
	if err != nil {
		return err
	}

	if !removeOpts.Force {
		wtPath := ws.WorktreePath(repoName)
		hasOrphans, err := hasOrphanCommits(wtPath, config)
		if err != nil {
			return err
		}

		if hasOrphans {
			return fmt.Errorf("worktree %q has commits on no branch (will be lost); run 'git -C \"%s\" branch <name>' to keep them, or pass --force to discard", wtPath, wtPath)
		}
	}

	err = removeWorktree(repo.bareRepoPath(config), ws.WorktreePath(repoName), removeOpts.Force, config)
	if err != nil {
		return err
	}

	delete(ws.Metadata.Repos, repoName)
	return ws.writeMetadata()
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

type Worktree struct {
	Path     string
	HEAD     string
	Branch   string // "" if detached
	Bare     bool
	Prunable bool
}

func parseWorktreeList(out string) []Worktree {

	// EXAMPLE:
	//
	// ❯ git -C /Users/amrox/.local/share/aworkspace/repos/github.com/amrox/dotfiles/worktrees/dotfiles1 worktree list --porcelain
	// worktree /Users/amrox/.local/share/aworkspace/repos/github.com/amrox/dotfiles
	// bare
	//
	// worktree /Users/amrox/Workspaces/hi/code/dotfiles
	// HEAD fbb7487ee3d69ee86cfe75e83febca1858f6d3b4
	// branch refs/heads/ws/hi
	// prunable gitdir file points to non-existent location
	//
	// worktree /Users/amrox/Workspaces/test1/code/dotfiles
	// HEAD 019a06151d0e0fdcc8870dacb967623493400712
	// branch refs/heads/ws/test1

	var wts []Worktree
	var cur *Worktree
	flush := func() {
		if cur != nil {
			wts = append(wts, *cur)
			cur = nil
		}
	}

	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			flush()
			continue
		}
		key, rest, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			flush()
			cur = &Worktree{Path: rest}
		case "HEAD":
			cur.HEAD = rest
		case "branch":
			cur.Branch = strings.TrimPrefix(rest, "refs/heads/")
		case "bare":
			cur.Bare = true
		case "prunable":
			cur.Prunable = true
		}
	}
	flush()
	return wts
}

func (ws *Workspace) WorktreesByRepo(config Config) (map[string][]Worktree, error) {

	worktreeMap := make(map[string][]Worktree)

	for name, rc := range ws.Metadata.Repos {

		repo, err := parseRepoURL(rc.URL)
		if err != nil {
			return nil, err
		}

		repoPath := repo.bareRepoPath(config)

		if _, err := os.Stat(repoPath); errors.Is(err, os.ErrNotExist) {
			worktreeMap[name] = nil // repo in metadata, no bare on disk -> no worktrees here
			continue
		} else if err != nil {
			return nil, err
		}

		out, err := execGitCommand(config, "-C", repoPath, "worktree", "list", "--porcelain")
		if err != nil {
			return nil, err
		}

		allWorktrees := parseWorktreeList(out)
		var myWorktrees []Worktree

		for _, wtree := range allWorktrees {
			if wtree.Bare || wtree.Prunable {
				continue
			}

			under, err := isUnderDir(ws.Path, wtree.Path)
			if err != nil {
				return nil, err
			}
			if under {
				myWorktrees = append(myWorktrees, wtree)
			}
		}

		worktreeMap[name] = myWorktrees
	}

	return worktreeMap, nil
}

type ArchiveOptions struct {
	Force bool
}

func (ws *Workspace) Archive(archiveOptions ArchiveOptions, config Config) (string, error) {

	// -- Step 0: Gates

	if !archiveOptions.Force {

		blocked := false
		for name := range ws.Metadata.Repos {

			wtPath := ws.WorktreePath(name)

			dirty, err := isWorktreeDirty(wtPath, config)
			if err != nil {
				return "", err
			}

			if dirty {
				LogWarning("worktree %q is dirty; commit or discard changes", wtPath)
			}

			hasOrphans, err := hasOrphanCommits(wtPath, config)
			if err != nil {
				return "", err
			}

			if hasOrphans {
				LogWarning("worktree %q has commits on no branch (will be lost); run 'git -C \"%s\" branch <name>' to keep them", wtPath, wtPath)
			}

			blocked = blocked || dirty || hasOrphans
		}

		if blocked {
			return "", fmt.Errorf("one or more worktrees not safe to archive; resolve issues or pass --force")
		}
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
