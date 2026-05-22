package cmd

import (
	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use: "new <name>",
	Short: "Create new workspace",
	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		ws, err := workspace.CreateWorkspace(name, *cfg)
		if err != nil {
			return err
		}
		workspace.LogInfo("Created workspace: %s\n", ws.Name())
		return nil
	},
}