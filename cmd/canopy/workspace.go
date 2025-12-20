package main

import "github.com/spf13/cobra"

// workspace.go defines the parent workspace command.

var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"w"},
	Short:   "Manage workspaces",
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}
