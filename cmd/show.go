package cmd

import (
	"fmt"
	"os"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use: "show",
	Short: "Show current workspace info",
	RunE: func(cmd *cobra.Command, args []string) error {

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

		fmt.Printf("Workspace Name: %v\n", ws.Name())
		fmt.Printf("Workspace Path: %v\n", ws.Path)

		return nil
	},
}