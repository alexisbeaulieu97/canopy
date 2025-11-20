package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexisbeaulieu97/yard/internal/config"
	"github.com/alexisbeaulieu97/yard/internal/gitx"
	"github.com/alexisbeaulieu97/yard/internal/workspace"
	"github.com/alexisbeaulieu97/yard/internal/workspaces"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of current workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

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

		gitEngine := gitx.New(cfg.ProjectsRoot)
		wsEngine := workspace.New(cfg.WorkspacesRoot)
		service := workspaces.NewService(cfg, gitEngine, wsEngine, logger)

		status, err := service.GetStatus(workspaceID)
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
