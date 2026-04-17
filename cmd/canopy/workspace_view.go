package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/formatting"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_view.go defines the "workspace view" subcommand.

const unknownDisplay = "unknown"

var workspaceViewCmd = &cobra.Command{
	Use:   "view <ID>",
	Short: "View details of a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		viewData, err := loadWorkspaceViewData(cmd.Context(), app.Service, id)
		if err != nil {
			return err
		}

		if jsonOutput {
			branchName := ""

			var repos []domain.RepoStatus

			if viewData.Status != nil {
				branchName = viewData.Status.BranchName
				repos = viewData.Status.Repos
			}

			return output.PrintJSON(map[string]interface{}{
				"workspace":        viewData.Workspace.ID,
				"branch":           branchName,
				"repos":            repos,
				"path":             viewData.Path,
				"repo_count":       len(viewData.Workspace.Repos),
				"disk_usage_bytes": viewData.Workspace.DiskUsageBytes,
				"last_modified":    viewData.Workspace.LastModified,
				"locked":           viewData.Locked,
				"orphans":          viewData.Orphans,
				"warnings":         viewData.Warnings,
			})
		}

		renderWorkspaceView(cmd.OutOrStdout(), viewData)

		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceViewCmd)

	workspaceViewCmd.Flags().Bool("json", false, "Output in JSON format")
}

type workspaceViewService interface {
	GetStatus(ctx context.Context, workspaceID string) (*domain.WorkspaceStatus, error)
	ListWorkspaces(ctx context.Context) ([]domain.Workspace, error)
	WorkspacePath(ctx context.Context, workspaceID string) (string, error)
	WorkspaceLocked(workspaceID string) (bool, error)
	DetectOrphansForWorkspace(ctx context.Context, workspaceID string) ([]domain.OrphanedWorktree, error)
}

type workspaceViewData struct {
	Workspace domain.Workspace
	Status    *domain.WorkspaceStatus
	Path      string
	Locked    bool
	Orphans   []domain.OrphanedWorktree
	Warnings  []string
}

func loadWorkspaceViewData(ctx context.Context, service workspaceViewService, id string) (*workspaceViewData, error) {
	workspaces, err := service.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	var workspace *domain.Workspace

	for i := range workspaces {
		if workspaces[i].ID == id {
			workspace = &workspaces[i]
			break
		}
	}

	if workspace == nil {
		return nil, cerrors.NewWorkspaceNotFound(id)
	}

	status, err := service.GetStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	path, err := service.WorkspacePath(ctx, id)
	if err != nil {
		return nil, err
	}

	var warnings []string

	locked, err := service.WorkspaceLocked(id)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Lock status unavailable: %v", err))
	}

	orphans, err := service.DetectOrphansForWorkspace(ctx, id)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Orphan detection unavailable: %v", err))
	}

	return &workspaceViewData{
		Workspace: *workspace,
		Status:    status,
		Path:      path,
		Locked:    locked,
		Orphans:   orphans,
		Warnings:  warnings,
	}, nil
}

func renderWorkspaceView(out io.Writer, viewData *workspaceViewData) {
	if viewData == nil {
		return
	}

	icons := output.NewIcons()
	branchName, repoStatuses := workspaceViewStatus(viewData)

	sections := []output.BoxSection{
		renderWorkspaceMetadataSection(viewData, branchName),
	}

	if warningSection, ok := renderWorkspaceWarningsSection(viewData); ok {
		sections = append(sections, warningSection)
	}

	if orphanSection, ok := renderWorkspaceOrphansSection(viewData); ok {
		sections = append(sections, orphanSection)
	}

	sections = append(sections, renderWorkspaceRepositoriesSection(viewData, repoStatuses, icons))

	box := output.NewBox("Workspace: " + viewData.Workspace.ID).WithWidth(88).WithWriter(out)
	box.RenderWithSections(sections)
}

func workspaceViewStatus(viewData *workspaceViewData) (string, []domain.RepoStatus) {
	if viewData == nil || viewData.Status == nil {
		return unknownDisplay, nil
	}

	return viewData.Status.BranchName, viewData.Status.Repos
}

func renderWorkspaceMetadataSection(viewData *workspaceViewData, branchName string) output.BoxSection {
	lines := []string{
		output.FormatKeyValue("Path", viewData.Path, 12),
		output.FormatKeyValue("Branch", branchName, 12),
		output.FormatKeyValue("Repos", fmt.Sprintf("%d", len(viewData.Workspace.Repos)), 12),
		output.FormatKeyValue("Disk", output.FormatBytes(viewData.Workspace.DiskUsageBytes), 12),
		output.FormatKeyValue("Modified", formatWorkspaceModified(viewData.Workspace.LastModified), 12),
	}

	if viewData.Locked {
		lines = append(lines, output.FormatKeyValue("Lock", "locked", 12))
	}

	return output.BoxSection{Lines: lines}
}

func renderWorkspaceWarningsSection(viewData *workspaceViewData) (output.BoxSection, bool) {
	if len(viewData.Warnings) == 0 {
		return output.BoxSection{}, false
	}

	lines := make([]string, 0, len(viewData.Warnings))
	for _, warning := range viewData.Warnings {
		lines = append(lines, "  "+warning)
	}

	return output.BoxSection{
		Title: "Warnings",
		Lines: lines,
	}, true
}

func renderWorkspaceOrphansSection(viewData *workspaceViewData) (output.BoxSection, bool) {
	if len(viewData.Orphans) == 0 {
		return output.BoxSection{}, false
	}

	lines := []string{
		fmt.Sprintf("  %d orphaned worktree(s) need attention", len(viewData.Orphans)),
	}
	for _, orphan := range viewData.Orphans {
		lines = append(lines, fmt.Sprintf("  - %s: %s", orphan.RepoName, orphan.ReasonDescription()))
	}

	return output.BoxSection{
		Title: "Orphans",
		Lines: lines,
	}, true
}

func renderWorkspaceRepositoriesSection(
	viewData *workspaceViewData,
	repoStatuses []domain.RepoStatus,
	icons output.Icons,
) output.BoxSection {
	return output.BoxSection{
		Title: fmt.Sprintf("Repositories (%d)", len(repoStatuses)),
		Lines: renderWorkspaceRepoLines(viewData, repoStatuses, icons),
	}
}

func renderWorkspaceRepoLines(
	viewData *workspaceViewData,
	repoStatuses []domain.RepoStatus,
	icons output.Icons,
) []string {
	switch {
	case viewData.Status == nil:
		return []string{"  Status unavailable for this workspace."}
	case len(repoStatuses) == 0:
		return []string{
			"  No repositories in this workspace.",
			fmt.Sprintf("  Add one with: canopy workspace repo add %s <repo>", viewData.Workspace.ID),
		}
	default:
		lines := []string{
			fmt.Sprintf("  %-14s %-20s %s", "NAME", "BRANCH", "STATUS"),
			"  " + output.HorizontalRule(72),
		}
		for _, repo := range repoStatuses {
			lines = append(lines, renderWorkspaceRepoLine(repo, icons))
		}

		return lines
	}
}

func renderWorkspaceRepoLine(repo domain.RepoStatus, icons output.Icons) string {
	name := repo.Name
	if len(name) > 12 {
		name = name[:11] + "…"
	}

	branch := repo.Branch
	if len(branch) > 18 {
		branch = branch[:17] + "…"
	}

	return fmt.Sprintf("  %-14s %-20s %s", name, branch, formatRepoViewStatus(repo, icons))
}

func formatWorkspaceModified(ts time.Time) string {
	if ts.IsZero() {
		return unknownDisplay
	}

	return formatting.RelativeTime(ts, formatting.RelativeTimeOptions{
		Zero:          unknownDisplay,
		Compact:       true,
		Yesterday:     true,
		UseWeeks:      true,
		AbsoluteAfter: 30 * 24 * time.Hour,
	})
}

func formatRepoViewStatus(r domain.RepoStatus, icons output.Icons) string {
	if r.Error != "" {
		return output.Colorize(output.ErrorStyle, icons.Error()+" status error: "+string(r.Error))
	}

	var parts []string

	if r.IsDirty {
		parts = append(parts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%s modified", icons.Dirty())))
	}

	if r.UnpushedCommits > 0 {
		parts = append(parts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%s %d unpushed", icons.Unpushed(), r.UnpushedCommits)))
	}

	if r.BehindRemote > 0 {
		parts = append(parts, output.Colorize(output.WarningStyle, fmt.Sprintf("%s %d behind", icons.Behind(), r.BehindRemote)))
	}

	if len(parts) == 0 {
		return output.Colorize(output.SuccessStyle, icons.Success()+" clean")
	}

	return joinOutputParts(parts)
}

func joinOutputParts(parts []string) string {
	return strings.Join(parts, output.Colorize(output.MutedStyle, " • "))
}
