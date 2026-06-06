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

func execGitCommand(config Config, args ...string) (string, error) {

	git := "git"
	if config.Git.Path != "" {
		git = config.Git.Path
	}

	return execCommand(git, args...)
}

func ensureBareRepo(repoURL string, config Config) (string, error) {

	repo, err := parseRepoURL(repoURL)
	if err != nil {
		return "", err
	}

	destPath := repo.bareRepoPath(config)
	if _, err := os.Stat(destPath); err == nil {
		Log(LogLevelNormal, "Using existing bare repo at: %v\n", destPath)
	} else {
		Log(LogLevelNormal, "Cloning bare repo to: %v\n", destPath)
		_, err = execGitCommand(config, "clone", "--bare", repoURL, destPath)
		if err != nil {
			return "", err
		}
	}

	// TODO: remove clone on error?

	_, err = execGitCommand(config, "-C", destPath,
		"config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	if err != nil {
		return "", err
	}

	_, err = execGitCommand(config, "-C", destPath,
		"fetch", "origin")
	if err != nil {
		return "", err
	}

	_, err = execGitCommand(config, "-C", destPath,
		"remote", "set-head", "origin", "--auto")
	if err != nil {
		return "", err
	}

	return destPath, nil
}

func getDefaultBranch(bareRepoPath string, config Config) (string, error) {

	out, err := execGitCommand(config, "-C", bareRepoPath,
		"symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func addWorktree(bareRepoPath string, worktreeDest string, branch string, config Config) error {

	Log(LogLevelNormal, "Creating worktree: %v branch: %v\n", worktreeDest, branch)

	defaultBranch, err := getDefaultBranch(bareRepoPath, config)
	if err != nil {
		return err
	}

	_, err = execGitCommand(config, "-C", bareRepoPath,
		"rev-parse", "--verify", branch)

	if err != nil {
		// branch does not exist, create with tracking
		_, err = execGitCommand(config, "-C", bareRepoPath,
			"worktree", "add", worktreeDest, "-b", branch, defaultBranch)
		return err
	}

	// branch exists, check out
	_, err = execGitCommand(config, "-C", bareRepoPath,
		"worktree", "add", worktreeDest, branch)
	return err
}
