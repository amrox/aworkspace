package cmd

import (
	_ "embed"
	"text/template"

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

		t1, err := template.New("t1").Parse(initShTemplate)
		if err != nil {
			return err
		}

		data := map[string]string {
			"Shell": shell,
		}
		err = t1.Execute(cmd.OutOrStdout(), data)
		if err != nil {
			return err
		}

		return nil
	},
}