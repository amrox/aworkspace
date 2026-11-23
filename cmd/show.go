package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	addDirFlag(showCmd)
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
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

		wsCanon, err := workspace.CanonicalPath(ws.Path)
		if err != nil {
			return err
		}

		worktreeMap, err := ws.WorktreesByRepo(*cfg)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Workspace Name: %v\n", ws.Name())
		fmt.Fprintf(cmd.OutOrStdout(), "Workspace Path: %v\n", ws.Path)
		fmt.Fprintf(cmd.OutOrStdout(), "Repos:\n")
		for name, rc := range ws.Metadata.Repos {
			// TODO: sort keys
			fmt.Fprintf(cmd.OutOrStdout(), "- %s: %s\n", name, rc.URL)

			for _, wt := range worktreeMap[name] {

				relPath, err := filepath.Rel(wsCanon, wt.Path)
				if err != nil {
					relPath = wt.Path // fallback to absolute if somehow not relative
				}

				ref := wt.Branch
				if ref == "" && len(wt.HEAD) >= 7 {
					ref = wt.HEAD[:7] + " (detached))"
				}

				fmt.Fprintf(cmd.OutOrStdout(), "    - %s @ %s\n", relPath, wt.Branch)
			}
		}

		return nil
	},
}
