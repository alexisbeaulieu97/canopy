package main

import (
	"fmt"

	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// workspace_git.go defines the "workspace git" subcommand.

var workspaceGitCmd = &cobra.Command{
	Use:   "git <WORKSPACE-ID> [--] <git-args...>",
	Short: "Run a git command across all repositories in a workspace",
	Long: `Execute any git command in all repositories within a workspace.

The command is run in each repository and results are displayed with clear separation.
Use -- to separate flags for the git command from canopy flags.

Examples:
  canopy workspace git my-workspace status
  canopy workspace git my-workspace -- fetch --all
  canopy workspace git my-workspace --parallel pull`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID := args[0]
		gitArgs := args[1:]

		parallel, _ := cmd.Flags().GetBool("parallel")
		continueOnError, _ := cmd.Flags().GetBool("continue-on-error")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		opts := workspaces.GitRunOptions{
			Parallel:        parallel,
			ContinueOnError: continueOnError,
		}

		results, err := app.Service.RunGitInWorkspace(cmd.Context(), workspaceID, gitArgs, opts)
		if err != nil && !continueOnError {
			// Print any results we got before the error.
			printGitResults(results)
			return err
		}

		printGitResults(results)

		// Count failures for exit code.
		var failures int
		for _, r := range results {
			if r.Error != nil || r.ExitCode != 0 {
				failures++
			}
		}

		if failures > 0 {
			output.Infof("\n%d/%d repos failed", failures, len(results))
			return cerrors.NewCommandFailed("git", fmt.Errorf("%d repos failed", failures))
		}

		output.Infof("\nAll %d repos completed successfully", len(results))

		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceGitCmd)

	workspaceGitCmd.Flags().Bool("parallel", false, "Execute git command in repos concurrently")
	workspaceGitCmd.Flags().Bool("continue-on-error", false, "Continue execution even if a repo fails")
}
