package cmd

import (
	"fmt"

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

		config, err := workspace.LoadOrDefaultConfig("")
		if err != nil {
			return err
		}

		ws, err := workspace.CreateWorkspace(name, config)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Created workspace:", ws.Name())
		return nil
	},
}