package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/formatting"
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
		workspacesWithStatus := make([]workspaceWithStatusData, 0, len(list))

		for _, w := range list {
			workspacesWithStatus = append(workspacesWithStatus, workspaceWithStatusData{Workspace: w})
		}

		if showStatus {
			setRepoStatusError := func(ws *workspaceWithStatusData, statusErr error) {
				if statusErr == nil {
					return
				}

				errorValue := domain.StatusError(statusErr.Error())
				if errors.Is(statusErr, context.DeadlineExceeded) {
					errorValue = domain.StatusErrorTimeout
				}

				ws.RepoStatuses = ws.RepoStatuses[:0]
				for _, repo := range ws.Repos {
					ws.RepoStatuses = append(ws.RepoStatuses, domain.RepoStatus{
						Name:  repo.Name,
						Error: errorValue,
					})
				}
			}

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

					if result.Err != nil {
						setRepoStatusError(ws, result.Err)

						if !errors.Is(result.Err, context.DeadlineExceeded) {
							output.Warnf("Failed to get status for %s: %v", ws.ID, result.Err)
						}
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
					} else if statusErr != nil {
						setRepoStatusError(ws, statusErr)

						if !errors.Is(statusErr, context.DeadlineExceeded) {
							output.Warnf("Failed to get status for %s: %v", w.ID, statusErr)
						}
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
			// For JSON output, convert to a simpler structure
			type workspaceJSONOutput struct {
				domain.Workspace
				RepoStatuses []domain.RepoStatus `json:"repo_statuses,omitempty"`
			}

			if showStatus || showLocks {
				jsonWorkspaces := make([]workspaceJSONOutput, len(workspacesWithStatus))
				for i, ws := range workspacesWithStatus {
					jsonWorkspaces[i] = workspaceJSONOutput(ws)
				}

				return output.PrintJSON(map[string]interface{}{
					"workspaces": jsonWorkspaces,
				})
			}

			return output.PrintJSON(map[string]interface{}{
				"workspaces": list,
			})
		}

		// Render workspaces in a formatted table
		if len(workspacesWithStatus) == 0 {
			output.Println("No workspaces found.")
			output.Println("")
			output.Println(output.Colorize(output.MutedStyle, "Create one with: canopy workspace new <name> --repos <repo1,repo2>"))

			return nil
		}

		renderWorkspaceListTable(workspacesWithStatus, showStatus, showLocks)

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

// workspaceWithStatus combines workspace with optional status info.
type workspaceWithStatusData struct {
	domain.Workspace
	RepoStatuses []domain.RepoStatus `json:"repo_statuses,omitempty"`
}

//nolint:gocyclo // UI rendering function with multiple format paths
func renderWorkspaceListTable(workspaces []workspaceWithStatusData, showStatus, showLocks bool) {
	icons := output.NewIcons()

	// Build the box
	box := output.NewBox("Workspaces").WithWidth(70)

	var lines []string

	// Header row
	header := fmt.Sprintf("  %-16s %-6s %-10s %-12s %s",
		"WORKSPACE", "REPOS", "SIZE", "MODIFIED", "STATUS")
	lines = append(lines, header)

	// Separator
	lines = append(lines, "  "+output.HorizontalRule(64))

	// Calculate totals
	var (
		totalSize                  int64
		dirtyCount, needsSyncCount int
	)

	for _, ws := range workspaces {
		line, dirty, needsSync := formatWorkspaceRow(ws, showLocks, icons)
		lines = append(lines, line)
		totalSize += ws.DiskUsageBytes
		dirtyCount += dirty
		needsSyncCount += needsSync
	}

	box.Render(lines)

	// Summary footer
	summaryParts := []string{
		fmt.Sprintf("%d workspaces", len(workspaces)),
		output.FormatBytes(totalSize) + " total",
	}

	if showStatus && dirtyCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d dirty", dirtyCount))
	}

	if showStatus && needsSyncCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d needs sync", needsSyncCount))
	}

	_, _ = fmt.Fprintln(os.Stdout)                                  //nolint:forbidigo // user-facing CLI output
	_, _ = fmt.Fprintln(os.Stdout, output.Summary(summaryParts...)) //nolint:forbidigo // user-facing CLI output
}

func formatWorkspaceRow(ws workspaceWithStatusData, showLocks bool, icons output.Icons) (line string, dirty, needsSync int) {
	name := ws.ID
	if len(name) > 14 {
		name = name[:13] + "…"
	}

	repoCount := fmt.Sprintf("%d", len(ws.Repos))
	size := output.FormatBytes(ws.DiskUsageBytes)
	modified := formatting.RelativeTime(ws.LastModified, formatting.RelativeTimeOptions{
		Zero:          "-",
		AbsoluteAfter: 7 * 24 * time.Hour,
	})

	status := formatWorkspaceStatus(ws.RepoStatuses, icons)
	if showLocks && ws.Locked {
		lockIndicator := output.Colorize(output.WarningStyle, " [locked]")
		status += lockIndicator
	}

	for _, rs := range ws.RepoStatuses {
		if rs.IsDirty {
			dirty++
		}

		if rs.BehindRemote > 0 {
			needsSync++
		}
	}

	line = fmt.Sprintf("  %-16s %-6s %-10s %-12s %s",
		name, repoCount, size, modified, status)

	return line, dirty, needsSync
}

//nolint:gocyclo // UI formatting function with multiple status paths
func formatWorkspaceStatus(statuses []domain.RepoStatus, icons output.Icons) string {
	if len(statuses) == 0 {
		return output.Colorize(output.MutedStyle, "-")
	}

	dirty, unpushed, behind, hasError := countStatusValues(statuses)
	if hasError {
		return output.Colorize(output.ErrorStyle, icons.Error()+" error")
	}

	if dirty == 0 && unpushed == 0 && behind == 0 {
		return output.Colorize(output.SuccessStyle, icons.Success()+" clean")
	}

	return formatFirstStatusIssue(dirty, unpushed, behind, icons)
}

func countStatusValues(statuses []domain.RepoStatus) (dirty, unpushed, behind int, hasError bool) {
	for _, s := range statuses {
		if s.Error != "" {
			return 0, 0, 0, true
		}

		if s.IsDirty {
			dirty++
		}

		if s.UnpushedCommits > 0 {
			unpushed += s.UnpushedCommits
		}

		if s.BehindRemote > 0 {
			behind += s.BehindRemote
		}
	}

	return dirty, unpushed, behind, false
}

func formatFirstStatusIssue(dirty, unpushed, behind int, icons output.Icons) string {
	if dirty > 0 {
		return output.Colorize(output.ErrorStyle, fmt.Sprintf("%s %d dirty", icons.Dirty(), dirty))
	}

	if unpushed > 0 {
		return output.Colorize(output.ErrorStyle, fmt.Sprintf("%s %d unpushed", icons.Unpushed(), unpushed))
	}

	if behind > 0 {
		return output.Colorize(output.WarningStyle, fmt.Sprintf("%s %d behind", icons.Behind(), behind))
	}

	return output.Colorize(output.SuccessStyle, icons.Success()+" clean")
}
