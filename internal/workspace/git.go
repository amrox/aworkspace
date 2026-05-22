package workspace

import (
	"os"
	"path/filepath"
	"strings"

	giturls "github.com/whilp/git-urls"
)

func parseRepoURL(rawURL string) (Repo, error) {
	u, err := giturls.Parse(rawURL)
	if err != nil {
		return Repo{}, err
	}

	trimmed := strings.TrimSuffix(u.Path, ".git")
	trimmed = strings.TrimPrefix(trimmed, "/")

	split := strings.Split(trimmed, "/")

	name := split[len(split)-1]
	namespace := strings.Join(split[0:len(split)-1], "/")

	var host string
	if u.Scheme == "file" {
		host = "_"
	} else {
		host = u.Host
	}

	repo := Repo{
		Host:      host,
		Name:      name,
		Namespace: namespace,
		Url:       rawURL,
	}

	return repo, nil
}

func (r Repo) bareRepoSubpath() string {

	path := filepath.Join(r.Host, r.Namespace, r.Name)
	return path
}

func (r Repo) bareRepoPath(config Config) string {

	subPath := r.bareRepoSubpath()
	return filepath.Join(config.BaresDir, subPath)
}

func execGitCommand(config Config, args ...string) error {

	git := "git"
	if config.Git.Path != "" {
		git = config.Git.Path
	}

	return execCommand(git, args...)
}

func cloneBareRepo(repoURL string, config Config) (string, error) {

	repo, err := parseRepoURL(repoURL)
	if err != nil {
		return "", err
	}

	destPath := repo.bareRepoPath(config)
	if _, err := os.Stat(destPath); err == nil {
		Log(LogLevelNormal, "Using existing bare repo at: %v\n", destPath)
		return destPath, nil
	}

	Log(LogLevelNormal, "Cloning bare repo to: %v\n", destPath)
	err = execGitCommand(config, "clone", "--bare", repoURL, destPath)
	if err != nil {
		return "", err
	}
	return destPath, nil
}

func addWorktree(bareRepoPath string, worktreeDest string, branch string, config Config) error {

	Log(LogLevelNormal, "Creating worktree: %v branch: %v\n", worktreeDest, branch)

	return execGitCommand(config, "-C", bareRepoPath, "worktree", "add", worktreeDest, "-B", branch)
}
