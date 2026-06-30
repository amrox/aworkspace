package workspace

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Errors
var ErrNotInWorkspace = errors.New("not inside a workspace")

type GitConfig struct {
	Path string `toml:"path"`
}

type Config struct {
	BaresDir       string `toml:"bares_dir"`
	WorkspacesDir  string `toml:"workspaces_dir"`
	WorktreeSubdir string `toml:"workspace_worktree_subdir"`
	BranchPrefix   string `toml:"branch_prefix"`

	Git GitConfig `toml:"git"`

	// TODO: this is adapted from "init_submodules"
	// I can think of 3 modes: "none", "init", "worktree-init"
	// submoduleMode string
}

func homeDir() string {
	homeDirVal, err := os.UserHomeDir()
	if err != nil {
		panic("could not determine home directory: " + err.Error())
	}
	return homeDirVal
}

func DefaultConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(homeDir(), ".config")
	}
	return filepath.Join(configHome, "aworkspace", "config.toml")
}

func DefaultBaresPath() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(homeDir(), ".local", "share")
	}
	return filepath.Join(dataHome, "aworkspace", "repos")
}

func DefaultConfig() Config {
	// TODO: should we default to raw values ("~/Workspaces") and canonicalize paths later?

	return Config{
		WorkspacesDir:  filepath.Join(homeDir(), "Workspaces"),
		BaresDir:       DefaultBaresPath(),
		WorktreeSubdir: "code",
		BranchPrefix:   "",
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
	Path     string
	Metadata WorkspaceMetadata
}

func (ws Workspace) Name() string {
	if ws.Path == "" {
		return ""
	}
	return filepath.Base(ws.Path)
}

type Repo struct {
	Name      string
	Namespace string
	Host      string

	Url string
}
