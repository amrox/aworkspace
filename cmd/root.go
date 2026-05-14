package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aworkspace",
	Short: "a workspace tool",
	Long: `yada
                yada
                yada`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}

func addDirFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("dir", "C", "", "run as if started in `path`")
}

func getStartDir(cmd *cobra.Command) (string, error) {
	dir, _ := cmd.Flags().GetString("dir")
	if dir != "" {
		return filepath.Abs(dir)
	}
	return os.Getwd()
}
