package cmd

import (
	"fmt"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	addDirFlag(showCmd)
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use: "show",
	Short: "Show current workspace info",
	RunE: func(cmd *cobra.Command, args []string) error {

		startDir, err := getStartDir(cmd)
		if err != nil {
			return err
		}

		wsDir, err := workspace.FindWorkspaceDir(startDir)
		if err != nil {
			return err
		}

		ws, err := workspace.LoadWorkspace(wsDir)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Workspace Name: %v\n", ws.Name())
		fmt.Fprintf(cmd.OutOrStdout(), "Workspace Path: %v\n", ws.Path)
		fmt.Fprintf(cmd.OutOrStdout(), "Repos:\n")
		for name, rc := range(ws.Metadata.Repos) {
			// TODO: sort keys
			fmt.Fprintf(cmd.OutOrStdout(), "- %s: %s\n", name, rc.URL)
		}

		return nil
	},
}