package main

import (
	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_reopen.go defines the "workspace reopen" subcommand.

var workspaceReopenCmd = &cobra.Command{
	Use:     "reopen <ID>",
	Aliases: []string{"open"},
	Short:   "Reopen a closed workspace",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		if err := app.Service.RestoreWorkspace(cmd.Context(), id, force); err != nil {
			return err
		}

		output.Success("Restored workspace", id)
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceReopenCmd)

	workspaceReopenCmd.Flags().Bool("force", false, "Close and restore the workspace, replacing an active workspace with the same ID if it exists")
}
