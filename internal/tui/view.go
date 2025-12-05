package tui

import (
	"fmt"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// View renders the UI for the current state.
func (m Model) View() string {
	if m.detailView {
		return m.renderDetailView()
	}

	return m.renderListView()
}

// renderListView renders the main workspace list view.
func (m Model) renderListView() string {
	var b strings.Builder

	// Header section
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Show spinner if pushing
	if m.pushing {
		spinnerLine := fmt.Sprintf("%s Pushing %s...",
			m.spinner.View(),
			accentTextStyle.Render(m.pushTarget))
		b.WriteString(spinnerLine)
		b.WriteString("\n\n")
	}

	// Confirmation prompt if active
	if m.confirming {
		b.WriteString(m.renderConfirmPrompt())
		b.WriteString("\n\n")
	}

	// Main list
	b.WriteString(m.list.View())

	// Footer with shortcuts
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderHeader renders the top header bar.
func (m Model) renderHeader() string {
	var parts []string

	// Title with count
	total := len(m.allItems)
	visible := len(m.list.Items())

	titleText := fmt.Sprintf("🌲 Workspaces (%d)", total)
	if visible != total {
		titleText = fmt.Sprintf("🌲 Workspaces (%d/%d)", visible, total)
	}

	parts = append(parts, titleStyle.Render(titleText))

	// Disk usage
	if m.totalDiskUsage > 0 {
		diskInfo := fmt.Sprintf("💾 %s", humanizeBytes(m.totalDiskUsage))
		parts = append(parts, mutedTextStyle.Render(diskInfo))
	}

	// Active filters
	var filters []string

	if m.filterStale {
		filters = append(filters, badgeWarnStyle.Render("STALE"))
	}

	if m.list.FilterValue() != "" {
		searchBadge := badgeInfoStyle.Render(fmt.Sprintf("🔍 %s", m.list.FilterValue()))
		filters = append(filters, searchBadge)
	}

	if len(filters) > 0 {
		parts = append(parts, strings.Join(filters, " "))
	}

	header := strings.Join(parts, "  ")

	// Error message if any
	if m.err != nil {
		header += "\n" + statusDirtyStyle.Render(fmt.Sprintf("⚠ Error: %v", m.err))
	}

	// Info message if any
	if m.infoMessage != "" {
		header += "\n" + statusCleanStyle.Render(fmt.Sprintf("✓ %s", m.infoMessage))
	}

	return header
}

// renderConfirmPrompt renders the confirmation dialog.
func (m Model) renderConfirmPrompt() string {
	var actionDesc string

	switch m.actionToConfirm {
	case actionClose:
		actionDesc = "close (delete local files)"
	case actionPush:
		actionDesc = "push all changes in"
	default:
		actionDesc = m.actionToConfirm
	}

	prompt := fmt.Sprintf("⚠️  Confirm %s workspace %s?",
		actionDesc,
		accentTextStyle.Render(m.confirmingID))

	hint := subtleTextStyle.Render("Press [y] to confirm, [n] or [esc] to cancel")

	return confirmPromptStyle.Render(prompt) + "\n" + hint
}

// renderFooter renders the keyboard shortcuts footer.
func (m Model) renderFooter() string {
	if m.pushing {
		return subtleTextStyle.Render("⏳ Push in progress...")
	}

	if m.confirming {
		return ""
	}

	shortcuts := []string{
		"↑↓ navigate",
		"/ search",
		"s stale",
		"⏎ details",
		"o open",
		"p push",
		"c close",
		"q quit",
	}

	return subtleTextStyle.Render(strings.Join(shortcuts, "  •  "))
}

// renderDetailView renders the detailed workspace view.
func (m Model) renderDetailView() string {
	var b strings.Builder

	// Loading state
	if m.loadingDetail {
		b.WriteString(fmt.Sprintf("%s Loading workspace details...", m.spinner.View()))

		return b.String()
	}

	// No workspace selected
	if m.selectedWS == nil {
		b.WriteString(subtleTextStyle.Render("No workspace selected."))
		b.WriteString("\n\n")
		b.WriteString(helpTextStyle.Render("Press [esc] to return to the list"))

		return b.String()
	}

	// Workspace header
	header := fmt.Sprintf("📂 %s", m.selectedWS.ID)
	b.WriteString(detailHeaderStyle.Render(header))
	b.WriteString("\n\n")

	// Metadata section
	b.WriteString(m.renderDetailMetadata())
	b.WriteString("\n\n")

	// Orphans section (if any)
	if len(m.wsOrphans) > 0 {
		b.WriteString(m.renderDetailOrphans())
		b.WriteString("\n\n")
	}

	// Repos section
	b.WriteString(m.renderDetailRepos())
	b.WriteString("\n\n")

	// Footer
	b.WriteString(helpTextStyle.Render("Press [esc] or [q] to return"))

	return b.String()
}

// renderDetailMetadata renders workspace metadata in the detail view.
func (m Model) renderDetailMetadata() string {
	var rows []string

	// Branch
	row := detailLabelStyle.Render("Branch:") + " " +
		detailValueStyle.Render(m.selectedWS.BranchName)
	rows = append(rows, row)

	// Disk usage
	row = detailLabelStyle.Render("Disk Usage:") + " " +
		detailValueStyle.Render(humanizeBytes(m.selectedWS.DiskUsageBytes))
	rows = append(rows, row)

	// Last modified
	row = detailLabelStyle.Render("Last Modified:") + " " +
		detailValueStyle.Render(relativeTime(m.selectedWS.LastModified))
	rows = append(rows, row)

	// Repo count
	if m.wsStatus != nil {
		repoCount := len(m.wsStatus.Repos)
		row = detailLabelStyle.Render("Repositories:") + " " +
			detailValueStyle.Render(fmt.Sprintf("%d", repoCount))
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// renderDetailRepos renders the repository list in the detail view.
func (m Model) renderDetailRepos() string {
	var b strings.Builder

	b.WriteString(boldTextStyle.Render("Repositories"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n")

	if m.wsStatus == nil || len(m.wsStatus.Repos) == 0 {
		b.WriteString(subtleTextStyle.Render("No repositories found."))
		return b.String()
	}

	for _, repo := range m.wsStatus.Repos {
		b.WriteString(m.renderRepoLine(repo))
		b.WriteString("\n")
	}

	return b.String()
}

// renderRepoLine renders a single repository line.
func (m Model) renderRepoLine(repo domain.RepoStatus) string {
	var statusParts []string

	if repo.IsDirty {
		statusParts = append(statusParts, statusDirtyStyle.Render("dirty"))
	}

	if repo.UnpushedCommits > 0 {
		statusParts = append(statusParts,
			statusDirtyStyle.Render(fmt.Sprintf("%d unpushed", repo.UnpushedCommits)))
	}

	if repo.BehindRemote > 0 {
		statusParts = append(statusParts,
			statusWarnStyle.Render(fmt.Sprintf("%d behind", repo.BehindRemote)))
	}

	if len(statusParts) == 0 {
		statusParts = append(statusParts, statusCleanStyle.Render("✓ clean"))
	}

	statusStr := strings.Join(statusParts, " • ")

	// Format: icon name [branch] status
	return fmt.Sprintf("  📁 %-20s %s  %s",
		repo.Name,
		subtleTextStyle.Render(fmt.Sprintf("[%s]", repo.Branch)),
		statusStr)
}

// renderDetailOrphans renders the orphaned worktrees section in the detail view.
func (m Model) renderDetailOrphans() string {
	var b strings.Builder

	b.WriteString(statusWarnStyle.Render("⚠ Orphaned Worktrees"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n")

	for _, orphan := range m.wsOrphans {
		line := fmt.Sprintf("  ⚠ %-20s %s",
			orphan.RepoName,
			statusWarnStyle.Render(orphan.ReasonDescription()))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
