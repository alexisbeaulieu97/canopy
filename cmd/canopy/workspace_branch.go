package main

import (
	"context"
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
		return validateWorkspaceTargetArgs(
			cmd,
			args,
			2,
			"args",
			"workspace ID and branch name are required",
			1,
			"branch",
			"branch name is required when using --pattern or --all",
		)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		create, _ := cmd.Flags().GetBool("create")

		pattern, err := resolveWorkspaceBulkPattern(cmd)
		if err != nil {
			return err
		}

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service

		if pattern != "" {
			branchName := args[0]

			ids, err := resolveBulkWorkspaceIDs(cmd.Context(), service, pattern)
			if err != nil {
				return err
			}

			if len(ids) == 0 {
				output.Info("No matching workspaces found.")
				return nil
			}

			printMatchedWorkspaceIDs(ids)

			report := runBulkWorkspaceOperations(cmd.Context(), ids, bulkWorkspaceRunOptions[struct{}]{
				Parallelism: 1,
				OnStart: func(id string, current, total int) {
					output.Infof("Switching workspace %s (%d/%d)", id, current, total)
				},
				OnFailure: func(id string, _, _ int, err error) {
					output.Warnf("Failed to switch workspace %s: %v", id, err)
				},
			}, func(ctx context.Context, id string) (struct{}, error) {
				return struct{}{}, service.SwitchBranch(ctx, id, branchName, create)
			})

			output.Success("Bulk branch switch completed", fmt.Sprintf("%d succeeded, %d failed", len(report.SuccessIDs), len(report.FailedIDs)))

			if len(report.FailedIDs) > 0 {
				output.Warnf("Failed workspaces: %s", strings.Join(report.FailedIDs, ", "))
				return cerrors.NewCommandFailed("branch", report.FirstErr)
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
