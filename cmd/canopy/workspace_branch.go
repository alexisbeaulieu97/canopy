package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_branch.go defines the "workspace branch" subcommand.

var workspaceBranchCmd = &cobra.Command{
	Use:   "branch [ID] <BRANCH-NAME>",
	Short: "Switch branch for all repositories in a workspace",
	Args: func(cmd *cobra.Command, args []string) error {
		pattern, _ := cmd.Flags().GetString("pattern")
		all, _ := cmd.Flags().GetBool("all")
		if all && pattern != "" {
			return cerrors.NewInvalidArgument("pattern", "cannot use --pattern with --all")
		}

		if pattern != "" || all {
			if len(args) != 1 {
				return cerrors.NewInvalidArgument("branch", "branch name is required when using --pattern or --all")
			}

			return nil
		}

		if len(args) != 2 {
			return cerrors.NewInvalidArgument("args", "workspace ID and branch name are required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		create, _ := cmd.Flags().GetBool("create")
		pattern, _ := cmd.Flags().GetString("pattern")
		all, _ := cmd.Flags().GetBool("all")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service

		if all {
			pattern = ".*"
		}

		if pattern != "" {
			branchName := args[0]
			matched, err := service.ListWorkspacesMatching(cmd.Context(), pattern)
			if err != nil {
				return err
			}

			if len(matched) == 0 {
				output.Info("No matching workspaces found.")
				return nil
			}

			ids := make([]string, len(matched))
			for i, ws := range matched {
				ids[i] = ws.ID
			}

			output.Infof("Matched %d workspaces:", len(ids))
			for _, id := range ids {
				output.Infof("  - %s", id)
			}

			var (
				successIDs []string
				failedIDs  []string
				firstErr   error
			)

			for i, id := range ids {
				output.Infof("Switching workspace %s (%d/%d)", id, i+1, len(ids))
				if err := service.SwitchBranch(cmd.Context(), id, branchName, create); err != nil {
					if firstErr == nil {
						firstErr = err
					}
					failedIDs = append(failedIDs, id)
					output.Warnf("Failed to switch workspace %s: %v", id, err)
					continue
				}

				successIDs = append(successIDs, id)
			}

			output.Success("Bulk branch switch completed", fmt.Sprintf("%d succeeded, %d failed", len(successIDs), len(failedIDs)))
			if len(failedIDs) > 0 {
				output.Warnf("Failed workspaces: %s", strings.Join(failedIDs, ", "))
				return cerrors.NewCommandFailed("branch", firstErr)
			}

			return nil
		}

		id := args[0]
		branchName := args[1]

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
	workspaceBranchCmd.Flags().String("pattern", "", "Switch branches for workspaces matching a regex pattern")
	workspaceBranchCmd.Flags().Bool("all", false, "Switch branches for all workspaces (equivalent to --pattern \".*\")")
}
