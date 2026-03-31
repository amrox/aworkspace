package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommandExists(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command failed: %v", err)
	}
}

func TestListSubcommandExists(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	// We expect no error even though list does nothing yet
	if err != nil {
		t.Fatalf("list subcommand not found: %v", err)
	}
}
