package cmd

import (
	"os"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addRepoCmd)
}

var addRepoCmd = &cobra.Command{
	Use: "add-repo",
	Short: "Add a repo to this workspace",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]

		config, err := workspace.LoadOrDefaultConfig("")
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		wsDir, err := workspace.FindWorkspaceDir(cwd)
		if err != nil {
			return err
		}

		ws, err := workspace.LoadWorkspace(wsDir)
		if err != nil {
			return err
		}

		err = ws.AddRepo(repoURL, "", config)
		if err != nil {
			return err
		}

		return nil
	},
}
