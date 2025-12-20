package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_rename.go defines the "workspace rename" subcommand.

var workspaceRenameCmd = &cobra.Command{
	Use:   "rename <OLD-ID> <NEW-ID>",
	Short: "Rename a workspace",
	Long:  `Rename a workspace to a new ID. Optionally renames branches in all repos if they match the old ID.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldID := args[0]
		newID := args[1]

		// Fast-fail if IDs are the same.
		if oldID == newID {
			return fmt.Errorf("old and new workspace IDs are the same")
		}

		renameBranch, _ := cmd.Flags().GetBool("rename-branch")
		force, _ := cmd.Flags().GetBool("force")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		if err := app.Service.RenameWorkspace(cmd.Context(), oldID, newID, renameBranch, force); err != nil {
			return err
		}

		if renameBranch {
			output.Infof("Renamed workspace %s to %s (branches also renamed)", oldID, newID)
		} else {
			output.Infof("Renamed workspace %s to %s", oldID, newID)
		}
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceRenameCmd)

	workspaceRenameCmd.Flags().Bool("rename-branch", true, "Rename branches in repos if they match the old workspace ID")
	workspaceRenameCmd.Flags().Bool("force", false, "Overwrite if target workspace already exists")
}
