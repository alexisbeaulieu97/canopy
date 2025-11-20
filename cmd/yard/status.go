package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of current workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		cfg := app.Config

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Check if we are inside a workspace
		relPath, err := filepath.Rel(cfg.WorkspacesRoot, cwd)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("not inside a workspace")
		}

		// Extract workspace ID from path
		parts := strings.Split(relPath, string(os.PathSeparator))
		if len(parts) == 0 {
			return fmt.Errorf("unable to determine workspace from path")
		}
		workspaceID := parts[0]

		status, err := app.Service.GetStatus(workspaceID)
		if err != nil {
			return err
		}

		fmt.Printf("Workspace: %s\n", status.ID)
		for _, r := range status.Repos {
			statusStr := "Clean"
			if r.IsDirty {
				statusStr = "Dirty"
			}
			fmt.Printf("- %s: %s (Branch: %s, Unpushed: %d)\n", r.Name, statusStr, r.Branch, r.UnpushedCommits)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
