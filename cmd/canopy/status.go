package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of current workspace",
	RunE: func(cmd *cobra.Command, _ []string) error {
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
		relPath, err := filepath.Rel(cfg.GetWorkspacesRoot(), cwd)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return cerrors.NewNotInWorkspace(cwd)
		}

		// Extract workspace ID from path
		parts := strings.Split(relPath, string(os.PathSeparator))
		if len(parts) == 0 {
			return cerrors.NewNotInWorkspace(cwd)
		}
		workspaceID := parts[0]

		status, err := app.Service.GetStatus(workspaceID)
		if err != nil {
			return err
		}

		fmt.Printf("Workspace: %s\n", status.ID) //nolint:forbidigo // user-facing CLI output
		for _, r := range status.Repos {
			statusStr := "Clean"
			if r.IsDirty {
				statusStr = "Dirty"
			}
			fmt.Printf("- %s: %s (Branch: %s, Unpushed: %d)\n", r.Name, statusStr, r.Branch, r.UnpushedCommits) //nolint:forbidigo // user-facing CLI output
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
