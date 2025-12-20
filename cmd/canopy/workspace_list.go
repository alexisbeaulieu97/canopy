package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_list.go defines the "workspace list" subcommand.

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active workspaces",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service

		jsonOutput, _ := cmd.Flags().GetBool("json")
		closedOnly, _ := cmd.Flags().GetBool("closed")
		showStatus, _ := cmd.Flags().GetBool("status")
		showLocks, _ := cmd.Flags().GetBool("show-locks")
		timeoutStr, _ := cmd.Flags().GetString("timeout")

		// Parse timeout duration.
		timeout := 5 * time.Second
		if timeoutStr != "" {
			var parseErr error
			timeout, parseErr = time.ParseDuration(timeoutStr)
			if parseErr != nil {
				return cerrors.NewInvalidArgument("timeout", fmt.Sprintf("invalid duration: %v", parseErr))
			}
		}

		if closedOnly {
			archives, err := service.ListClosedWorkspaces(cmd.Context())
			if err != nil {
				return err
			}

			if jsonOutput {
				var payload []domain.Workspace

				for _, a := range archives {
					payload = append(payload, a.Metadata)
				}

				return output.PrintJSON(map[string]interface{}{
					"workspaces": payload,
				})
			}

			for _, a := range archives {
				closedDate := "unknown"
				if a.Metadata.ClosedAt != nil {
					closedDate = a.Metadata.ClosedAt.Format(time.RFC3339)
				}

				output.Infof("%s (Closed: %s)", a.Metadata.ID, closedDate)
				for _, r := range a.Metadata.Repos {
					output.Infof("  - %s (%s)", r.Name, r.URL)
				}
			}

			return nil
		}

		list, err := service.ListWorkspaces(cmd.Context())
		if err != nil {
			return err
		}

		// Collect status for each workspace if --status flag is set.
		type workspaceWithStatus struct {
			domain.Workspace
			RepoStatuses []domain.RepoStatus `json:"repo_statuses,omitempty"`
		}

		var workspacesWithStatus []workspaceWithStatus

		for _, w := range list {
			ws := workspaceWithStatus{Workspace: w}

			if showStatus {
				ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
				status, statusErr := service.GetStatus(ctx, w.ID)
				cancel()

				if statusErr == nil && status != nil {
					ws.RepoStatuses = status.Repos
				} else if errors.Is(statusErr, context.DeadlineExceeded) {
					// Timeout - add placeholder status.
					for range w.Repos {
						ws.RepoStatuses = append(ws.RepoStatuses, domain.RepoStatus{
							Branch: "timeout",
						})
					}
				}
			}

			if showLocks {
				locked, lockErr := service.WorkspaceLocked(w.ID)
				if lockErr != nil {
					output.Warnf("Failed to check lock status for %s: %v", w.ID, lockErr)
				} else {
					ws.Locked = locked
				}
			}

			workspacesWithStatus = append(workspacesWithStatus, ws)
		}

		if jsonOutput {
			if showStatus {
				return output.PrintJSON(map[string]interface{}{
					"workspaces": workspacesWithStatus,
				})
			}

			return output.PrintJSON(map[string]interface{}{
				"workspaces": list,
			})
		}

		for _, ws := range workspacesWithStatus {
			lockSuffix := ""
			if showLocks && ws.Locked {
				lockSuffix = " [locked]"
			}

			output.Infof("%s (Branch: %s)%s", ws.ID, ws.BranchName, lockSuffix)
			for i, r := range ws.Repos {
				if showStatus && i < len(ws.RepoStatuses) {
					status := ws.RepoStatuses[i]
					statusStr := formatRepoStatusIndicator(status)
					output.Infof("  - %s (%s) %s", r.Name, r.URL, statusStr)
				} else {
					output.Infof("  - %s (%s)", r.Name, r.URL)
				}
			}
		}
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)

	workspaceListCmd.Flags().Bool("json", false, "Output in JSON format")
	workspaceListCmd.Flags().Bool("closed", false, "List closed workspaces")
	workspaceListCmd.Flags().Bool("status", false, "Show git status for each repository")
	workspaceListCmd.Flags().String("timeout", "5s", "Timeout for status check per workspace (e.g. 5s, 10s)")
	workspaceListCmd.Flags().Bool("show-locks", false, "Show workspace lock status")
}
