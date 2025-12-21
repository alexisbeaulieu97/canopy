package main

import (
	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_repo.go defines repo subcommands under workspace.

var workspaceRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories in a workspace",
}

var workspaceRepoAddCmd = &cobra.Command{
	Use:   "add <WORKSPACE-ID> <REPO-NAME>",
	Short: "Add a repository to an existing workspace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID := args[0]
		repoName := args[1]

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service

		if err := service.AddRepoToWorkspace(cmd.Context(), workspaceID, repoName); err != nil {
			return err
		}

		output.Infof("Added repository %s to workspace %s", repoName, workspaceID)
		return nil
	},
}

var workspaceRepoRemoveCmd = &cobra.Command{
	Use:   "remove <WORKSPACE-ID> <REPO-NAME>",
	Short: "Remove a repository from an existing workspace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID := args[0]
		repoName := args[1]

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service

		if err := service.RemoveRepoFromWorkspace(cmd.Context(), workspaceID, repoName); err != nil {
			return err
		}

		output.Infof("Removed repository %s from workspace %s", repoName, workspaceID)
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceRepoCmd)
	workspaceRepoCmd.AddCommand(workspaceRepoAddCmd)
	workspaceRepoCmd.AddCommand(workspaceRepoRemoveCmd)
}
