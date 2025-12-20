package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// workspace_sync.go defines the "workspace sync" subcommand.

var workspaceSyncCmd = &cobra.Command{
	Use:   "sync <ID>",
	Short: "Pull updates for all repositories in a workspace",
	Long: `Pull updates for all repositories in a workspace and display a summary.
Per-repository timeouts can be configured to prevent slow remotes from blocking the entire operation.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		timeoutStr, _ := cmd.Flags().GetString("timeout")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		var timeout time.Duration
		if timeoutStr != "" {
			var err error
			timeout, err = time.ParseDuration(timeoutStr)
			if err != nil {
				return cerrors.NewInvalidArgument("timeout", fmt.Sprintf("invalid duration: %v", err))
			}
		}

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		opts := workspaces.SyncOptions{
			Timeout: timeout,
		}

		result, err := app.Service.SyncWorkspace(cmd.Context(), id, opts)
		if err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(result)
		}

		output.Infof("Syncing workspace: %s", id)
		output.Println("")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		_, _ = fmt.Fprintln(w, "REPOSITORY\tSTATUS\tUPDATED\tDETAILS")

		for _, r := range result.Repos {
			status := strings.ToUpper(string(r.Status))
			updatedStr := fmt.Sprintf("%d commits", r.Updated)
			if r.Updated == 0 {
				updatedStr = "-"
			}

			// Sanitize error message to prevent breaking tabwriter layout.
			errDetail := r.Error
			errDetail = strings.ReplaceAll(errDetail, "\n", " ")
			errDetail = strings.ReplaceAll(errDetail, "\r", " ")
			errDetail = strings.ReplaceAll(errDetail, "\t", " ")
			if len(errDetail) > 100 {
				errDetail = errDetail[:97] + "..."
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, status, updatedStr, errDetail)
		}
		_ = w.Flush()

		if result.TotalErrors > 0 {
			output.Warnf("\nPartial failure: %d repositories failed to sync", result.TotalErrors)
			return cerrors.NewCommandFailed("sync", fmt.Errorf("%d repos failed", result.TotalErrors))
		}

		output.Success("\nWorkspace sync completed", fmt.Sprintf("%d total commits updated", result.TotalUpdated))
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceSyncCmd)

	workspaceSyncCmd.Flags().String("timeout", "60s", "Timeout for each repository sync (e.g. 30s, 2m)")
	workspaceSyncCmd.Flags().Bool("json", false, "Output in JSON format")
}
