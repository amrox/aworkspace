package workspace

import (
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	BaresDir      string
	WorkspacesDir string

	BranchPrefix string

	// TODO: this is adapted from "init_submodules"
	// I can think of 3 modes: "none", "init", "worktree-init"
	// submoduleMode string

}

func DefaultConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "aworkspace", "config.toml")
}

func DefaultConfig() Config {
	// TODO: should we default to raw values ("~/Workspaces") and canonicalize paths later?
	home, _ := os.UserHomeDir()
	return Config{
		WorkspacesDir: filepath.Join(home, "Workspaces"),
		BaresDir:      filepath.Join(home, "Repos"),
		BranchPrefix:  "ws/",
	}
}

func LoadConfig(path string) (Config, error) {
	var config Config
	err := errors.New("not yet implemented")
	return config, err
}

func LoadOrDefaultConfig(path string) (Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	config, err := LoadConfig(path)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	return config, err
}

type Workspace struct {
	path string
}

type Repo struct {
	path string
}
