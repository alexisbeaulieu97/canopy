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

		output.Printf("%s\n", output.Colorize(output.AccentStyle, fmt.Sprintf("=== %s ===", r.RepoName)))

		if r.Error != nil {
			output.Printf("%s\n", output.Colorize(output.ErrorStyle, fmt.Sprintf("Error: %s", r.Error)))
			continue
		}

		if r.Stdout != "" {
			output.Print(r.Stdout)
		}

		if r.Stderr != "" {
			output.Print(r.Stderr)
		}

		if r.ExitCode != 0 {
			output.Printf("%s\n", output.Colorize(output.ErrorStyle, fmt.Sprintf("Exit code: %d", r.ExitCode)))
		}
	}
}

func printWorkspaceClosePreview(preview *domain.WorkspaceClosePreview) {
	if preview == nil {
		return
	}

	output.Printf("%s Would close workspace: %s\n", output.Colorize(output.WarningStyle, "[DRY RUN]"), preview.WorkspaceID)

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
			output.Printf("  %s\n", output.Colorize(output.WarningStyle, fmt.Sprintf("⚠ %s has uncommitted changes", status.Name)))
		}

		if status.UnpushedCount > 0 {
			output.Printf("  %s\n", output.Colorize(output.WarningStyle, fmt.Sprintf("⚠ %s has %d unpushed commit(s)", status.Name, status.UnpushedCount)))
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
