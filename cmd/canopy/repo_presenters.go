package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

func printRepoRemovePreview(preview *domain.RepoRemovePreview) {
	if preview == nil {
		return
	}

	output.Printf("%s Would remove repository: %s\n", output.Colorize(output.WarningStyle, "[DRY RUN]"), preview.RepoName)
	output.Infof("  Remove directory: %s", preview.RepoPath)

	if len(preview.WorkspacesAffected) > 0 {
		output.Infof("  Used by workspaces: %s (will become orphaned)", strings.Join(preview.WorkspacesAffected, ", "))
	}

	if preview.DiskUsageBytes > 0 {
		output.Infof("  Size: %s", output.FormatBytes(preview.DiskUsageBytes))
	}
}

func printSingleRepoStatus(status *domain.CanonicalRepoStatus) {
	if status == nil {
		return
	}

	output.Infof("Repository:    %s", status.Name)
	output.Infof("Path:          %s", status.Path)
	output.Infof("Size:          %s", output.FormatBytes(status.DiskUsageBytes))

	if status.LastFetchTime != nil {
		output.Infof("Last Fetch:    %s", status.LastFetchTime.Format("2006-01-02 15:04:05"))
	} else {
		output.Infof("Last Fetch:    never")
	}

	if status.UsedByCount > 0 {
		output.Infof("Workspaces:    %d (%s)", status.UsedByCount, strings.Join(status.UsedBy, ", "))
	} else {
		output.Infof("Workspaces:    0 (orphaned)")
	}
}

func printRepoStatusesTable(statuses []domain.CanonicalRepoStatus) {
	if len(statuses) == 0 {
		output.Infof("No canonical repositories found.")
		return
	}

	output.Printf("%s %s %s %s\n",
		output.Column("NAME", output.RepoNameWidth, output.AccentStyle),
		output.Column("SIZE", output.RepoSizeWidth, output.MutedStyle),
		output.Column("LAST FETCH", output.RepoLastFetchWidth, output.MutedStyle),
		output.Column("WORKSPACES", output.RepoWorkspacesWidth, output.MutedStyle),
	)
	output.Printf("%s %s %s %s\n",
		output.HorizontalRule(output.RepoNameWidth),
		output.HorizontalRule(output.RepoSizeWidth),
		output.HorizontalRule(output.RepoLastFetchWidth),
		output.HorizontalRule(output.RepoWorkspacesWidth),
	)

	for _, s := range statuses {
		fetchTime := "never"
		if s.LastFetchTime != nil {
			fetchTime = s.LastFetchTime.Format("2006-01-02 15:04")
		}

		output.Printf("%s %s %s %s\n",
			output.Column(s.Name, output.RepoNameWidth, lipgloss.NewStyle()),
			output.Column(output.FormatBytes(s.DiskUsageBytes), output.RepoSizeWidth, lipgloss.NewStyle()),
			output.Column(fetchTime, output.RepoLastFetchWidth, lipgloss.NewStyle()),
			output.Column(fmt.Sprintf("%d", s.UsedByCount), output.RepoWorkspacesWidth, lipgloss.NewStyle()),
		)
	}

	output.Infof("\n%d repositories", len(statuses))
}
