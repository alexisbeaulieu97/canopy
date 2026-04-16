package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/formatting"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

type workspaceListOptions struct {
	jsonOutput     bool
	closedOnly     bool
	showStatus     bool
	showLocks      bool
	parallelStatus bool
	timeout        time.Duration
}

// workspaceWithStatus combines workspace with optional status info.
type workspaceWithStatusData struct {
	domain.Workspace
	RepoStatuses []domain.RepoStatus `json:"repo_statuses,omitempty"`
}

func runWorkspaceList(cmd *cobra.Command) error {
	app, err := getApp(cmd)
	if err != nil {
		return err
	}

	opts, err := resolveWorkspaceListOptions(cmd)
	if err != nil {
		return err
	}

	if opts.closedOnly {
		return runClosedWorkspaceList(cmd.Context(), app.Service, opts.jsonOutput)
	}

	workspacesWithStatus, list, err := loadWorkspaceListData(cmd.Context(), app.Service, opts)
	if err != nil {
		return err
	}

	return renderWorkspaceListOutput(list, workspacesWithStatus, opts)
}

func loadWorkspaceListData(ctx context.Context, service *workspaces.Service, opts workspaceListOptions) ([]workspaceWithStatusData, []domain.Workspace, error) {
	list, err := service.ListWorkspaces(ctx)
	if err != nil {
		return nil, nil, err
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})

	workspacesWithStatus := makeWorkspaceListData(list)

	if opts.showStatus {
		if err := populateWorkspaceStatuses(ctx, service, workspacesWithStatus, opts); err != nil {
			return nil, nil, err
		}
	}

	if opts.showLocks {
		populateWorkspaceLocks(service, workspacesWithStatus)
	}

	return workspacesWithStatus, list, nil
}

func renderWorkspaceListOutput(list []domain.Workspace, workspacesWithStatus []workspaceWithStatusData, opts workspaceListOptions) error {
	if opts.jsonOutput {
		return output.PrintJSON(workspaceListJSONPayload(list, workspacesWithStatus, opts.showStatus || opts.showLocks))
	}

	if len(workspacesWithStatus) == 0 {
		output.Println("No workspaces found.")
		output.Println("")
		output.Println(output.Colorize(output.MutedStyle, "Create one with: canopy workspace new <name> --repos <repo1,repo2>"))

		return nil
	}

	renderWorkspaceListTable(workspacesWithStatus, opts.showStatus, opts.showLocks)

	return nil
}

func resolveWorkspaceListOptions(cmd *cobra.Command) (workspaceListOptions, error) {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	closedOnly, _ := cmd.Flags().GetBool("closed")
	showStatus, _ := cmd.Flags().GetBool("status")
	showLocks, _ := cmd.Flags().GetBool("show-locks")
	parallelStatus, _ := cmd.Flags().GetBool("parallel-status")
	sequentialStatus, _ := cmd.Flags().GetBool("sequential-status")
	timeoutStr, _ := cmd.Flags().GetString("timeout")

	timeout := 5 * time.Second

	if timeoutStr != "" {
		var err error

		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return workspaceListOptions{}, cerrors.NewInvalidArgument("timeout", fmt.Sprintf("invalid duration: %v", err))
		}
	}

	if sequentialStatus && parallelStatus && cmd.Flags().Changed("parallel-status") {
		return workspaceListOptions{}, cerrors.NewInvalidArgument("flags", "cannot use --parallel-status with --sequential-status")
	}

	if sequentialStatus {
		parallelStatus = false
	}

	return workspaceListOptions{
		jsonOutput:     jsonOutput,
		closedOnly:     closedOnly,
		showStatus:     showStatus,
		showLocks:      showLocks,
		parallelStatus: parallelStatus,
		timeout:        timeout,
	}, nil
}

func runClosedWorkspaceList(ctx context.Context, service *workspaces.Service, jsonOutput bool) error {
	archives, err := service.ListClosedWorkspaces(ctx)
	if err != nil {
		return err
	}

	if jsonOutput {
		payload := make([]domain.Workspace, 0, len(archives))
		for _, archive := range archives {
			payload = append(payload, archive.Metadata)
		}

		return output.PrintJSON(map[string]interface{}{
			"workspaces": payload,
		})
	}

	for _, archive := range archives {
		closedDate := unknownDisplay
		if archive.Metadata.ClosedAt != nil {
			closedDate = archive.Metadata.ClosedAt.Format(time.RFC3339)
		}

		output.Infof("%s (Closed: %s)", archive.Metadata.ID, closedDate)

		for _, repo := range archive.Metadata.Repos {
			output.Infof("  - %s (%s)", repo.Name, repo.URL)
		}
	}

	return nil
}

func makeWorkspaceListData(list []domain.Workspace) []workspaceWithStatusData {
	workspacesWithStatus := make([]workspaceWithStatusData, 0, len(list))
	for _, workspace := range list {
		workspacesWithStatus = append(workspacesWithStatus, workspaceWithStatusData{Workspace: workspace})
	}

	return workspacesWithStatus
}

func populateWorkspaceStatuses(ctx context.Context, service *workspaces.Service, workspacesWithStatus []workspaceWithStatusData, opts workspaceListOptions) error {
	if opts.parallelStatus {
		return populateWorkspaceStatusesParallel(ctx, service, workspacesWithStatus, opts.timeout)
	}

	populateWorkspaceStatusesSequential(ctx, service, workspacesWithStatus, opts.timeout)

	return nil
}

func populateWorkspaceStatusesParallel(ctx context.Context, service *workspaces.Service, workspacesWithStatus []workspaceWithStatusData, timeout time.Duration) error {
	workspaceIDs := make([]string, 0, len(workspacesWithStatus))
	for _, workspace := range workspacesWithStatus {
		workspaceIDs = append(workspaceIDs, workspace.ID)
	}

	results, err := service.GetWorkspaceStatusBatch(ctx, workspaceIDs, timeout)
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
			setWorkspaceRepoStatusError(ws, result.Err)
		}
	}

	return nil
}

func populateWorkspaceStatusesSequential(ctx context.Context, service *workspaces.Service, workspacesWithStatus []workspaceWithStatusData, timeout time.Duration) {
	for i, workspace := range workspacesWithStatus {
		statusCtx, cancel := context.WithTimeout(ctx, timeout)
		status, statusErr := service.GetStatus(statusCtx, workspace.ID)

		cancel()

		ws := &workspacesWithStatus[i]
		if statusErr == nil && status != nil {
			ws.RepoStatuses = status.Repos
			continue
		}

		if statusErr != nil {
			setWorkspaceRepoStatusError(ws, statusErr)
		}
	}
}

func setWorkspaceRepoStatusError(ws *workspaceWithStatusData, statusErr error) {
	if statusErr == nil {
		return
	}

	errorValue := domain.StatusError(statusErr.Error())
	if errors.Is(statusErr, context.DeadlineExceeded) {
		errorValue = domain.StatusErrorTimeout
	}

	// Replace any existing repo statuses rather than merging them.
	// This error path intentionally reuses the backing slice when present.
	ws.RepoStatuses = ws.RepoStatuses[:0]
	for _, repo := range ws.Repos {
		ws.RepoStatuses = append(ws.RepoStatuses, domain.RepoStatus{
			Name:  repo.Name,
			Error: errorValue,
		})
	}
}

func populateWorkspaceLocks(service *workspaces.Service, workspacesWithStatus []workspaceWithStatusData) {
	for i := range workspacesWithStatus {
		ws := &workspacesWithStatus[i]

		locked, err := service.WorkspaceLocked(ws.ID)
		if err != nil {
			output.Warnf("Failed to check lock status for %s: %v", ws.ID, err)
			continue
		}

		ws.Locked = locked
	}
}

func workspaceListJSONPayload(list []domain.Workspace, workspacesWithStatus []workspaceWithStatusData, includeDerivedFields bool) map[string]interface{} {
	if !includeDerivedFields {
		return map[string]interface{}{
			"workspaces": list,
		}
	}

	type workspaceJSONOutput struct {
		domain.Workspace
		RepoStatuses []domain.RepoStatus `json:"repo_statuses,omitempty"`
	}

	jsonWorkspaces := make([]workspaceJSONOutput, len(workspacesWithStatus))
	for i, workspace := range workspacesWithStatus {
		jsonWorkspaces[i] = workspaceJSONOutput(workspace)
	}

	return map[string]interface{}{
		"workspaces": jsonWorkspaces,
	}
}

//nolint:gocyclo // UI rendering function with multiple format paths
func renderWorkspaceListTable(workspaces []workspaceWithStatusData, showStatus, showLocks bool) {
	icons := output.NewIcons()
	box := output.NewBox("Workspaces").WithWidth(92)

	lines := []string{
		fmt.Sprintf("  %-16s %-6s %-10s %-12s %s", "WORKSPACE", "REPOS", "SIZE", "MODIFIED", "STATUS"),
		"  " + output.HorizontalRule(84),
	}

	var (
		totalSize                              int64
		dirtyCount, needsSyncCount, errorCount int
	)

	for _, workspace := range workspaces {
		line, dirty, needsSync, rowErrors := formatWorkspaceRow(workspace, showStatus, showLocks, icons)
		lines = append(lines, line)
		totalSize += workspace.DiskUsageBytes
		dirtyCount += dirty
		needsSyncCount += needsSync
		errorCount += rowErrors
	}

	box.Render(lines)

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

	if showStatus && errorCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d status errors", errorCount))
	}

	_, _ = fmt.Fprintln(os.Stdout)                                  //nolint:forbidigo // user-facing CLI output
	_, _ = fmt.Fprintln(os.Stdout, output.Summary(summaryParts...)) //nolint:forbidigo // user-facing CLI output
}

func formatWorkspaceRow(ws workspaceWithStatusData, showStatus, showLocks bool, icons output.Icons) (line string, dirty, needsSync, rowErrors int) {
	name := ws.ID
	if len(name) > 14 {
		name = name[:13] + "…"
	}

	status, rowErrors := formatWorkspaceStatus(ws.RepoStatuses, showStatus, icons)
	if showLocks && ws.Locked {
		status = joinStatusParts(status, output.Colorize(output.WarningStyle, "locked"))
	}

	for _, repoStatus := range ws.RepoStatuses {
		if repoStatus.IsDirty {
			dirty++
		}

		if repoStatus.BehindRemote > 0 {
			needsSync++
		}
	}

	line = fmt.Sprintf("  %-16s %-6s %-10s %-12s %s",
		name,
		fmt.Sprintf("%d", len(ws.Repos)),
		output.FormatBytes(ws.DiskUsageBytes),
		formatting.RelativeTime(ws.LastModified, formatting.RelativeTimeOptions{
			Zero:          "-",
			AbsoluteAfter: 7 * 24 * time.Hour,
		}),
		status,
	)

	return line, dirty, needsSync, rowErrors
}

//nolint:gocyclo // UI formatting function with multiple status paths
func formatWorkspaceStatus(statuses []domain.RepoStatus, showStatus bool, icons output.Icons) (string, int) {
	if len(statuses) == 0 {
		if !showStatus {
			return output.Colorize(output.MutedStyle, "-"), 0
		}

		return output.Colorize(output.MutedStyle, "no repos"), 0
	}

	dirty, unpushed, behind, errorCount, timeoutCount := countStatusValues(statuses)
	if dirty == 0 && unpushed == 0 && behind == 0 && errorCount == 0 {
		return output.Colorize(output.SuccessStyle, icons.Success()+" clean"), 0
	}

	var parts []string

	if errorCount > 0 {
		label := fmt.Sprintf("%d status errors", errorCount)
		if timeoutCount == errorCount {
			label = fmt.Sprintf("%d timeouts", timeoutCount)
		}

		parts = append(parts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%s %s", icons.Error(), label)))
	}

	if dirty > 0 {
		parts = append(parts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%d dirty", dirty)))
	}

	if unpushed > 0 {
		parts = append(parts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%d unpushed", unpushed)))
	}

	if behind > 0 {
		parts = append(parts, output.Colorize(output.WarningStyle, fmt.Sprintf("%d behind", behind)))
	}

	return strings.Join(parts, output.Colorize(output.MutedStyle, " • ")), errorCount
}

func countStatusValues(statuses []domain.RepoStatus) (dirty, unpushed, behind, errorCount, timeoutCount int) {
	for _, status := range statuses {
		if status.Error != "" {
			errorCount++

			if status.Error == domain.StatusErrorTimeout {
				timeoutCount++
			}

			continue
		}

		if status.IsDirty {
			dirty++
		}

		if status.UnpushedCommits > 0 {
			unpushed += status.UnpushedCommits
		}

		if status.BehindRemote > 0 {
			behind += status.BehindRemote
		}
	}

	return dirty, unpushed, behind, errorCount, timeoutCount
}

func joinStatusParts(base, extra string) string {
	if base == "" {
		return extra
	}

	return base + output.Colorize(output.MutedStyle, " • ") + extra
}
