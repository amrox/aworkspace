package cmd

import (
	"fmt"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use: "list",
	Short: "List workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {

		config, err := workspace.LoadOrDefaultConfig("")
		if err != nil {
			return err
		}

		workspaces, err := workspace.ListWorkspaces(config)
		if err != nil {
			return err
		}

		for _, ws := range workspaces {
			fmt.Println(ws.Name())
		}

		return nil
	},
}