package main

import (
	"github.com/spf13/cobra"
)

// workspace_close.go defines the "workspace close" subcommand.

var workspaceCloseCmd = &cobra.Command{
	Use:   "close [ID]",
	Short: "Close a workspace (keep metadata or delete)",
	Args: func(cmd *cobra.Command, args []string) error {
		return validateWorkspaceTargetArgs(
			cmd,
			args,
			1,
			"id",
			"workspace ID is required",
			0,
			"id",
			"cannot provide workspace ID with --pattern or --all",
		)
	},
	RunE: runWorkspaceClose,
}

func init() {
	workspaceCmd.AddCommand(workspaceCloseCmd)

	workspaceCloseCmd.Flags().Bool("force", false, "Force close even if there are uncommitted changes")
	workspaceCloseCmd.Flags().Bool("keep", false, "Keep metadata (close without deleting)")
	workspaceCloseCmd.Flags().Bool("delete", false, "Delete without keeping metadata")
	workspaceCloseCmd.Flags().Bool("dry-run", false, "Preview what would be deleted without actually deleting")
	workspaceCloseCmd.Flags().Bool("json", false, "Output in JSON format (use with --dry-run)")
	workspaceCloseCmd.Flags().Bool("no-hooks", false, "Skip pre_close hooks")
	workspaceCloseCmd.Flags().Bool("hooks-only", false, "Run pre_close hooks without closing the workspace")
	workspaceCloseCmd.Flags().Bool("dry-run-hooks", false, "Preview pre_close hooks without executing them")
	workspaceCloseCmd.Flags().String("pattern", "", "Close workspaces matching a regex pattern")
	workspaceCloseCmd.Flags().Bool("all", false, "Close all workspaces (equivalent to --pattern \".*\")")
	workspaceCloseCmd.Flags().Bool("no-progress", false, "Disable progress indicators")
}
