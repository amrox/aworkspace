package workspace

import (
	"fmt"
	"os"
	"os/exec"
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
		Host: host,
		Name: name,
		Namespace: namespace,
		Url: rawURL,
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

func cloneBareRepo(repoURL string, config Config) (string, error) {

	repo, err := parseRepoURL(repoURL)
	if err != nil {
		return "", err
	}

	destPath := repo.bareRepoPath(config)

	if _, err := os.Stat(destPath); err == nil {
    	return destPath, nil
	}

	cmd := exec.Command("git", "clone", "--bare", repoURL, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %q: %w", output, err)
	}
	return destPath, nil
}

func addWorktree(bareRepoPath string, worktreeDest string, branch string) error {

	cmd := exec.Command("git", "-C", bareRepoPath, "worktree", "add", worktreeDest, "-B", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %s: %w", output, err)
	}
	return nil
}


