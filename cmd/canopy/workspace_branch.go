package main

import (
	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_branch.go defines the "workspace branch" subcommand.

var workspaceBranchCmd = &cobra.Command{
	Use:   "branch <ID> <BRANCH-NAME>",
	Short: "Switch branch for all repositories in a workspace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		branchName := args[1]
		create, _ := cmd.Flags().GetBool("create")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service

		if err := service.SwitchBranch(cmd.Context(), id, branchName, create); err != nil {
			return err
		}

		output.Infof("Switched workspace %s to branch %s", id, branchName)
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceBranchCmd)

	workspaceBranchCmd.Flags().Bool("create", false, "Create branch if it doesn't exist")
}
