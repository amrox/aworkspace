package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amrox/aworkspace/internal/workspace"
	"github.com/spf13/cobra"
)

func init() {
	addDirFlag(archiveCmd)
	archiveCmd.Flags().BoolP("force", "f", false, "archive even if a worktree is dirty")
	rootCmd.AddCommand(archiveCmd)
}

var archiveCmd = &cobra.Command{
	Use:   "archive <name>",
	Short: "Archive a workspace",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		wsDir := filepath.Join(cfg.WorkspacesDir, args[0])
		ws, err := workspace.LoadWorkspace(wsDir)
		if err != nil {
			return err
		}

		startDir, err := getStartDir(cmd)
		if err != nil {
			return err
		}
		abs, err := filepath.Abs(wsDir)
		if err != nil {
			return err
		}
		if startDir == abs || strings.HasPrefix(startDir, abs+string(os.PathSeparator)) {
			return fmt.Errorf("cannot archive %q from inside it (cwd %s); cd out first", args[0], startDir)
		}

		force, _ := cmd.Flags().GetBool("force")
		archived, err := ws.Archive(workspace.ArchiveOptions{Force: force}, *cfg)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Archived to %s\n", archived)
		return nil
	},
}
