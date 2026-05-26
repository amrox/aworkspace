package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cdCmd)
}

var cdCmd = &cobra.Command{
	Use:   "cd <name>",
	Short: "Change to workspace dir (requires shell hooks) (alias: switch)",
	Aliases: []string{"switch"},
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		workspaces, err := workspace.ListWorkspaces(*cfg)
		if err != nil {
		    return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var names []string
		for _, ws := range workspaces {
			names = append(names, ws.Name())
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		shellenv := os.Getenv("AWORKSPACE_SHELL")
		if shellenv == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "Hint: run 'eval \"$(aworkspace init <shell>)\"' to enable cd support")
			return nil
		}

		name := args[0]

		path := filepath.Join(cfg.WorkspacesDir, name)
		_, err := os.Stat(path)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s", path)

		return nil
	},
}
