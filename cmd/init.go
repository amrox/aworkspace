package cmd

import (
	_ "embed"
	"fmt"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

//go:embed templates/init.sh
var initShTemplate string

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Init shell hooks",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		shell := args[0]

		switch shell {
		case "bash", "zsh":
			// supported
		default:
			workspace.Log(workspace.LogLevelNormal, "WARNING: shell %v is unsupported\n", shell)
		}

		fmt.Fprint(cmd.OutOrStdout(), initShTemplate)

		return nil
	},
}