package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

var cfg *workspace.Config

var rootCmd = &cobra.Command{
	Use:   "aworkspace",
	Short: "a workspace tool",
	Long: `yada
                yada
                yada`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cfg != nil {
			return nil
		}
		c, err := workspace.LoadOrDefaultConfig("")
		if err != nil {
			return err
		}
		cfg = &c
		if v, _ := cmd.Flags().GetBool("verbose"); v {
			workspace.CurLogLevel = workspace.LogLevelVerbose
		}
		return nil
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
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
