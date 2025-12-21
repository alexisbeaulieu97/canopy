package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
		parallelStatus, _ := cmd.Flags().GetBool("parallel-status")
		sequentialStatus, _ := cmd.Flags().GetBool("sequential-status")
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
				payload := make([]domain.Workspace, 0, len(archives))

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

		sort.Slice(list, func(i, j int) bool {
			return list[i].ID < list[j].ID
		})

		// Collect status for each workspace if --status flag is set.
		type workspaceWithStatus struct {
			domain.Workspace
			RepoStatuses []domain.RepoStatus `json:"repo_statuses,omitempty"`
		}

		workspacesWithStatus := make([]workspaceWithStatus, 0, len(list))

		for _, w := range list {
			workspacesWithStatus = append(workspacesWithStatus, workspaceWithStatus{Workspace: w})
		}

		if showStatus {
			if sequentialStatus && parallelStatus && cmd.Flags().Changed("parallel-status") {
				return cerrors.NewInvalidArgument("flags", "cannot use --parallel-status with --sequential-status")
			}

			if sequentialStatus {
				parallelStatus = false
			}

			if parallelStatus {
				workspaceIDs := make([]string, 0, len(list))
				for _, w := range list {
					workspaceIDs = append(workspaceIDs, w.ID)
				}

				results, err := service.GetWorkspaceStatusBatch(cmd.Context(), workspaceIDs, timeout)
				if err != nil {
					return err
				}

				for i, result := range results {
					ws := &workspacesWithStatus[i]
					if result.Err == nil && result.Status != nil {
						ws.RepoStatuses = result.Status.Repos
						continue
					}

					if errors.Is(result.Err, context.DeadlineExceeded) {
						for _, repo := range ws.Repos {
							ws.RepoStatuses = append(ws.RepoStatuses, domain.RepoStatus{
								Name:   repo.Name,
								Branch: "timeout",
							})
						}
						continue
					}

					if result.Err != nil {
						output.Warnf("Failed to get status for %s: %v", ws.ID, result.Err)
					}
				}
			} else {
				for i, w := range list {
					ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
					status, statusErr := service.GetStatus(ctx, w.ID)
					cancel()

					ws := &workspacesWithStatus[i]
					if statusErr == nil && status != nil {
						ws.RepoStatuses = status.Repos
					} else if errors.Is(statusErr, context.DeadlineExceeded) {
						// Timeout - add placeholder status.
						for _, repo := range w.Repos {
							ws.RepoStatuses = append(ws.RepoStatuses, domain.RepoStatus{
								Name:   repo.Name,
								Branch: "timeout",
							})
						}
					} else if statusErr != nil {
						output.Warnf("Failed to get status for %s: %v", w.ID, statusErr)
					}
				}
			}
		}

		if showLocks {
			for i := range workspacesWithStatus {
				ws := &workspacesWithStatus[i]
				locked, lockErr := service.WorkspaceLocked(ws.ID)
				if lockErr != nil {
					output.Warnf("Failed to check lock status for %s: %v", ws.ID, lockErr)
				} else {
					ws.Locked = locked
				}
			}
		}

		if jsonOutput {
			if showStatus || showLocks {
				return output.PrintJSON(map[string]interface{}{
					"workspaces": workspacesWithStatus,
				})
			}

			return output.PrintJSON(map[string]interface{}{
				"workspaces": list,
			})
		}

		for _, ws := range workspacesWithStatus {
			statusByRepo := make(map[string]domain.RepoStatus, len(ws.RepoStatuses))
			for _, status := range ws.RepoStatuses {
				if status.Name == "" {
					continue
				}
				statusByRepo[status.Name] = status
			}

			lockSuffix := ""
			if showLocks && ws.Locked {
				lockSuffix = " [locked]"
			}

			output.Infof("%s (Branch: %s)%s", ws.ID, ws.BranchName, lockSuffix)
			for _, r := range ws.Repos {
				if showStatus {
					status, ok := statusByRepo[r.Name]
					if ok {
						statusStr := formatRepoStatusIndicator(status)
						output.Infof("  - %s (%s) %s", r.Name, r.URL, statusStr)
						continue
					}
				}

				output.Infof("  - %s (%s)", r.Name, r.URL)
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
	workspaceListCmd.Flags().Bool("parallel-status", true, "Fetch workspace status in parallel (default)")
	workspaceListCmd.Flags().Bool("sequential-status", false, "Fetch workspace status sequentially")
	workspaceListCmd.Flags().String("timeout", "5s", "Timeout for status check per workspace (e.g. 5s, 10s)")
	workspaceListCmd.Flags().Bool("show-locks", false, "Show workspace lock status")
}
