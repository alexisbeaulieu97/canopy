package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
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
		jsonOutput, _ := cmd.Flags().GetBool("json")

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

		status, err := app.Service.GetStatus(cmd.Context(), workspaceID)
		if err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(map[string]interface{}{
				"workspace": status.ID,
				"branch":    status.BranchName,
				"repos":     status.Repos,
			})
		}

		output.Infof("Workspace: %s", status.ID)
		for _, r := range status.Repos {
			statusStr := "Clean"
			if r.IsDirty {
				statusStr = "Dirty"
			}
			output.Infof("- %s: %s (Branch: %s, Unpushed: %d)", r.Name, statusStr, r.Branch, r.UnpushedCommits)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("json", false, "Output in JSON format")
}
