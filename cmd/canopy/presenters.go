package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// presenters.go contains output helpers for workspace-related CLI commands.

func printGitResults(results []workspaces.RepoGitResult) {
	for i, r := range results {
		if i > 0 {
			output.Info("")
		}

		output.Printf("\033[1;36m=== %s ===\033[0m\n", r.RepoName)

		if r.Error != nil {
			output.Printf("\033[1;31mError: %s\033[0m\n", r.Error)
			continue
		}

		if r.Stdout != "" {
			output.Print(r.Stdout)
		}

		if r.Stderr != "" {
			output.Print(r.Stderr)
		}

		if r.ExitCode != 0 {
			output.Printf("\033[1;31mExit code: %d\033[0m\n", r.ExitCode)
		}
	}
}

func printWorkspaceClosePreview(preview *domain.WorkspaceClosePreview) {
	if preview == nil {
		return
	}

	output.Printf("\033[33m[DRY RUN]\033[0m Would close workspace: %s\n", preview.WorkspaceID)

	action := "Delete"
	if preview.KeepMetadata {
		action = "Archive (keep metadata)"
	}

	output.Infof("  Action: %s", action)
	output.Infof("  Remove directory: %s", preview.WorkspacePath)

	if len(preview.ReposAffected) > 0 {
		output.Infof("  Repos affected: %s", strings.Join(preview.ReposAffected, ", "))
	}

	// Show warnings for repos with uncommitted changes or unpushed commits.
	for _, status := range preview.RepoStatuses {
		if status.IsDirty {
			output.Printf("  \033[33m⚠ %s has uncommitted changes\033[0m\n", status.Name)
		}

		if status.UnpushedCount > 0 {
			output.Printf("  \033[33m⚠ %s has %d unpushed commit(s)\033[0m\n", status.Name, status.UnpushedCount)
		}
	}

	if preview.DiskUsageBytes > 0 {
		output.Infof("  Total size: %s", output.FormatBytes(preview.DiskUsageBytes))
	}
}

func printClosed(id string, closedAt *time.Time) {
	if closedAt != nil {
		output.Infof("Closed workspace %s at %s", id, closedAt.Format(time.RFC3339))
		return
	}

	output.Success("Closed workspace", id)
}

// formatRepoStatusIndicator creates a human-readable status indicator for a repo.
func formatRepoStatusIndicator(status domain.RepoStatus) string {
	if status.Error != "" {
		if status.Error == domain.StatusErrorTimeout {
			return "[timeout]"
		}

		errText := strings.ReplaceAll(string(status.Error), "\n", " ")
		return fmt.Sprintf("[error: %s]", errText)
	}

	var parts []string

	if status.IsDirty {
		parts = append(parts, "dirty")
	}

	if status.UnpushedCommits > 0 {
		parts = append(parts, fmt.Sprintf("%d ahead", status.UnpushedCommits))
	}

	if status.BehindRemote > 0 {
		parts = append(parts, fmt.Sprintf("%d behind", status.BehindRemote))
	}

	if len(parts) == 0 {
		return "[clean]"
	}

	return "[" + strings.Join(parts, ", ") + "]"
}
