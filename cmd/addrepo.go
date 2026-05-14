package cmd

import (
	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	addDirFlag(addRepoCmd)
	rootCmd.AddCommand(addRepoCmd)
}

var addRepoCmd = &cobra.Command{
	Use: "add-repo",
	Short: "Add a repo to this workspace",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]

		startDir, err := getStartDir(cmd)
		if err != nil {
			return err
		}

		config, err := workspace.LoadOrDefaultConfig("")
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

		err = ws.AddRepo(repoURL, "", config)
		if err != nil {
			return err
		}

		return nil
	},
}
