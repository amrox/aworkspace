package workspace

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Errors
var ErrNotInWorkspace = errors.New("not inside a workspace")

type Config struct {
	BaresDir      string `toml:"bares_dir"`
	WorkspacesDir string `toml:"workspaces_dir"`

	BranchPrefix string `toml:"branch_prefix"`

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

func LoadConfigFromBytes(doc []byte) (Config, error) {
	config := DefaultConfig()
	err := toml.Unmarshal(doc, &config)
	return config, err
}

func LoadConfigFromPath(path string) (Config, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err == nil {
		config, err = LoadConfigFromBytes(data)
	}
	return config, err
}

func LoadOrDefaultConfig(path string) (Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	config, err := LoadConfigFromPath(path)
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


