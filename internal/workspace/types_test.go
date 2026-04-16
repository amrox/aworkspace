package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	tests := []struct {
		name           string
		xdgConfigHome  string
		home           string
		expectedSuffix string
	}{
		{
			name:           "uses XDG_CONFIG_HOME when set",
			xdgConfigHome:  "/custom/config",
			home:           "/home/user",
			expectedSuffix: "/custom/config/aworkspace/config.toml",
		},
		{
			name:           "falls back to HOME/.config when XDG_CONFIG_HOME not set",
			xdgConfigHome:  "",
			home:           "/home/user",
			expectedSuffix: "/home/user/.config/aworkspace/config.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origXDG := os.Getenv("XDG_CONFIG_HOME")
			origHome := os.Getenv("HOME")
			defer func() {
				os.Setenv("XDG_CONFIG_HOME", origXDG)
				os.Setenv("HOME", origHome)
			}()

			// Set test env vars
			if tt.xdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
			os.Setenv("HOME", tt.home)

			got := DefaultConfigPath()
			if got != tt.expectedSuffix {
				t.Errorf("DefaultConfigPath() = %v, want %v", got, tt.expectedSuffix)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Set test HOME
	testHome := "/home/testuser"
	os.Setenv("HOME", testHome)

	config := DefaultConfig()

	// Check WorkspacesDir
	expectedWorkspacesDir := filepath.Join(testHome, "Workspaces")
	if config.WorkspacesDir != expectedWorkspacesDir {
		t.Errorf("WorkspacesDir = %v, want %v", config.WorkspacesDir, expectedWorkspacesDir)
	}

	// Check ReposDir
	expectedReposDir := filepath.Join(testHome, "Repos")
	if config.BaresDir != expectedReposDir {
		t.Errorf("BaresDir = %v, want %v", config.BaresDir, expectedReposDir)
	}

	// Check BranchPrefix
	if config.BranchPrefix != "ws/" {
		t.Errorf("BranchPrefix = %v, want 'ws/'", config.BranchPrefix)
	}
}

func TestLoadConfigFromBytes(t *testing.T) {

	t.Run("empty input returns defaults", func(t *testing.T) {
		var empty []byte
		defaultConfig := DefaultConfig()

		config, err := LoadConfigFromBytes(empty)
		if err != nil {
			t.Fatalf("LoadConfigFromBytes returned unexpected error: %v:", err)
		}

		if config != defaultConfig {
			t.Errorf("got = %+v, want %+v", config, defaultConfig)
		}
	})

	t.Run("partial input merges with defaults", func(t *testing.T) {
		customPrefix := "qqq/"
		doc := fmt.Sprintf(`
			branch_prefix = "%s"
		`, customPrefix)
		defaultConfig := DefaultConfig()

		config, err := LoadConfigFromBytes([]byte(doc))
		if err != nil {
			t.Fatalf("LoadConfigFromBytes returned unexpected error: %v:", err)
		}

		if config.BranchPrefix != customPrefix {
			t.Errorf("BranchPrefix = %v, want %v", config.BranchPrefix, customPrefix)
		}

		if config.BaresDir != defaultConfig.BaresDir {
			t.Errorf("BaresDir = %v, want %v", config.BaresDir, defaultConfig.BaresDir)
		}

		if config.WorkspacesDir != defaultConfig.WorkspacesDir {
			t.Errorf("WorkspacesDir = %v, want %v", config.WorkspacesDir, defaultConfig.WorkspacesDir)
		}
	})
}
