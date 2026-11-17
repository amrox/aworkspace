package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	addDirFlag(rmRepoCmd)
	rmRepoCmd.Flags().BoolP("force", "f", false, "remove even if the worktree is dirty")
	rootCmd.AddCommand(rmRepoCmd)
}

var rmRepoCmd = &cobra.Command{
	Use:   "remove-repo",
	Short: "Remove a repo from this workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]

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

		wt := ws.WorktreePath(repoName)
		if startDir == wt || strings.HasPrefix(startDir, wt+string(os.PathSeparator)) {
			return fmt.Errorf("cannot remove %q from inside its worktree (cwd %s); cd out first", repoName, startDir)
		}

		force, _ := cmd.Flags().GetBool("force")

		err = ws.RemoveRepo(repoName, workspace.RemoveRepoOptions{Force: force}, *cfg)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Removed repo %s\n", repoName)

		return nil
	},
}
