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
			return output.PrintJSON(map[string]interface{}{
				"workspace":        viewData.Workspace.ID,
				"branch":           viewData.Status.BranchName,
				"repos":            viewData.Status.Repos,
				"path":             viewData.Path,
				"repo_count":       len(viewData.Workspace.Repos),
				"disk_usage_bytes": viewData.Workspace.DiskUsageBytes,
				"last_modified":    viewData.Workspace.LastModified,
				"locked":           viewData.Locked,
				"orphans":          viewData.Orphans,
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
}

func loadWorkspaceViewData(ctx context.Context, service workspaceViewService, id string) (*workspaceViewData, error) {
	status, err := service.GetStatus(ctx, id)
	if err != nil {
		return nil, err
	}

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

	path, err := service.WorkspacePath(ctx, id)
	if err != nil {
		return nil, err
	}

	locked, err := service.WorkspaceLocked(id)
	if err != nil {
		locked = false
	}

	orphans, err := service.DetectOrphansForWorkspace(ctx, id)
	if err != nil {
		orphans = nil
	}

	return &workspaceViewData{
		Workspace: *workspace,
		Status:    status,
		Path:      path,
		Locked:    locked,
		Orphans:   orphans,
	}, nil
}

func renderWorkspaceView(out io.Writer, viewData *workspaceViewData) {
	if viewData == nil || viewData.Status == nil {
		return
	}

	icons := output.NewIcons()

	var sections []output.BoxSection

	var metadataLines []string

	metadataLines = append(metadataLines, output.FormatKeyValue("Path", viewData.Path, 12))
	metadataLines = append(metadataLines, output.FormatKeyValue("Branch", viewData.Status.BranchName, 12))
	metadataLines = append(metadataLines, output.FormatKeyValue("Repos", fmt.Sprintf("%d", len(viewData.Workspace.Repos)), 12))
	metadataLines = append(metadataLines, output.FormatKeyValue("Disk", output.FormatBytes(viewData.Workspace.DiskUsageBytes), 12))
	metadataLines = append(metadataLines, output.FormatKeyValue("Modified", formatWorkspaceModified(viewData.Workspace.LastModified), 12))

	if viewData.Locked {
		metadataLines = append(metadataLines, output.FormatKeyValue("Lock", "locked", 12))
	}

	sections = append(sections, output.BoxSection{
		Title: "",
		Lines: metadataLines,
	})

	if len(viewData.Orphans) > 0 {
		orphanLines := []string{
			fmt.Sprintf("  %d orphaned worktree(s) need attention", len(viewData.Orphans)),
		}

		for _, orphan := range viewData.Orphans {
			orphanLines = append(orphanLines, fmt.Sprintf("  - %s: %s", orphan.RepoName, orphan.ReasonDescription()))
		}

		sections = append(sections, output.BoxSection{
			Title: "Orphans",
			Lines: orphanLines,
		})
	}

	var repoLines []string
	if len(viewData.Status.Repos) == 0 {
		repoLines = append(repoLines, "  No repositories in this workspace.")
		repoLines = append(repoLines, fmt.Sprintf("  Add one with: canopy workspace repo add %s <repo>", viewData.Workspace.ID))
	} else {
		header := fmt.Sprintf("  %-14s %-20s %s", "NAME", "BRANCH", "STATUS")
		repoLines = append(repoLines, header)
		repoLines = append(repoLines, "  "+output.HorizontalRule(72))

		for _, repo := range viewData.Status.Repos {
			name := repo.Name
			if len(name) > 12 {
				name = name[:11] + "…"
			}

			branch := repo.Branch
			if len(branch) > 18 {
				branch = branch[:17] + "…"
			}

			line := fmt.Sprintf("  %-14s %-20s %s", name, branch, formatRepoViewStatus(repo, icons))
			repoLines = append(repoLines, line)
		}
	}

	sections = append(sections, output.BoxSection{
		Title: fmt.Sprintf("Repositories (%d)", len(viewData.Status.Repos)),
		Lines: repoLines,
	})

	box := output.NewBox("Workspace: " + viewData.Workspace.ID).WithWidth(88).WithWriter(out)
	box.RenderWithSections(sections)
}

func formatWorkspaceModified(ts time.Time) string {
	if ts.IsZero() {
		return "unknown"
	}

	return formatting.RelativeTime(ts, formatting.RelativeTimeOptions{
		Zero:          "unknown",
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
