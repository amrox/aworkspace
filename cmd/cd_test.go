package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/amrox/aworkspace/internal/workspace"
)

func TestCd(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		t.Cleanup(func() {
			cfg = nil
		})

		workspacesDir := t.TempDir()
		config := workspace.DefaultConfig()
		config.WorkspacesDir = workspacesDir

		wsName := "test12"

		ws, err := workspace.CreateWorkspace(wsName, config)
		if err != nil {
			t.Fatalf("CreateWorkspace: %v", err)
		}

		t.Setenv("AWORKSPACE_SHELL", "1")

		cmd := rootCmd
		cmd.SetArgs([]string{"cd", wsName})
		cfg = &config

		var out bytes.Buffer
		cmd.SetOut(&out)

		err = cmd.Execute()
		if err != nil {
			t.Fatalf("cd failed: %v", err)
		}

		if !strings.Contains(out.String(), ws.Path) {
			t.Errorf("expected workspace path in output, got: %s", out.String())
		}
	})

	t.Run("workspace does not exist", func(t *testing.T) {
		t.Cleanup(func() {
			cfg = nil
		})

		workspacesDir := t.TempDir()
		config := workspace.DefaultConfig()
		config.WorkspacesDir = workspacesDir

		wsName := "test12"

		t.Setenv("AWORKSPACE_SHELL", "1")

		cmd := rootCmd
		cmd.SetArgs([]string{"cd", wsName})
		cfg = &config

		var out bytes.Buffer
		cmd.SetOut(&out)

		err := cmd.Execute()
		if err == nil {
			t.Errorf("expected error, not nil")
		}
	})

	t.Run("no AWORKSPACE_SHELL env", func(t *testing.T) {
		t.Cleanup(func() {
			cfg = nil
		})

		workspacesDir := t.TempDir()
		config := workspace.DefaultConfig()
		config.WorkspacesDir = workspacesDir

		wsName := "test12"

		cmd := rootCmd
		cmd.SetArgs([]string{"cd", wsName})
		cfg = &config

		var outBuf, errBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetErr(&errBuf)

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("cd failed: %v", err)
		}

		if outBuf.String() != "" {
			t.Errorf("expected empty output, got: %s", outBuf.String())
		}

		errExpected := "aworkspace init"
		if !strings.Contains(errBuf.String(), errExpected) {
			t.Errorf("expected stderr output containing %v, got: %s", errExpected, errBuf.String())
		}
	})
}
