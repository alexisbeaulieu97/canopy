package main

import "github.com/spf13/cobra"

// workspace_list.go defines the "workspace list" subcommand.

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active workspaces",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runWorkspaceList(cmd)
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)

	workspaceListCmd.Flags().Bool("json", false, "Output in JSON format")
	workspaceListCmd.Flags().Bool("closed", false, "List closed workspaces")
	workspaceListCmd.Flags().Bool("status", false, "Show git status for each repository")
	workspaceListCmd.Flags().Bool("parallel-status", true, "Fetch workspace status in parallel (default)")
	workspaceListCmd.Flags().Bool("sequential-status", false, "Fetch workspace status sequentially")
	workspaceListCmd.Flags().String("timeout", "5s", "Timeout for status check per workspace (e.g. 5s, 10s)")
	workspaceListCmd.Flags().Bool("show-locks", false, "Show workspace lock status")
}
