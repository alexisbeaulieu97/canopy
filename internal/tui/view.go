package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// View renders the UI by delegating to the current view state.
func (m Model) View() string {
	return m.viewState.View(&m)
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
			m.ui.Spinner.View(),
			accentTextStyle.Render(m.pushTarget))
		b.WriteString(spinnerLine)
		b.WriteString("\n\n")
	}

	// Main list
	b.WriteString(m.ui.List.View())

	// Footer with shortcuts
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderListViewWithConfirm renders the list view with a confirmation dialog overlay.
func (m Model) renderListViewWithConfirm(state *ConfirmViewState) string {
	var b strings.Builder

	// Header section
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Show spinner if pushing
	if m.pushing {
		spinnerLine := fmt.Sprintf("%s Pushing %s...",
			m.ui.Spinner.View(),
			accentTextStyle.Render(m.pushTarget))
		b.WriteString(spinnerLine)
		b.WriteString("\n\n")
	}

	// Confirmation prompt
	dialog := components.ConfirmDialog{
		Active:      true,
		Action:      state.Action,
		TargetLabel: m.confirmTargetLabel(state),
	}
	b.WriteString(dialog.Render())
	b.WriteString("\n\n")

	// Main list
	b.WriteString(m.ui.List.View())

	// Footer with shortcuts
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) confirmTargetLabel(state *ConfirmViewState) string {
	if state == nil || len(state.TargetIDs) == 0 {
		return ""
	}

	if len(state.TargetIDs) == 1 {
		return fmt.Sprintf("workspace %s", accentTextStyle.Render(state.TargetIDs[0]))
	}

	return fmt.Sprintf("%s workspaces", accentTextStyle.Render(strconv.Itoa(len(state.TargetIDs))))
}

// renderHeader renders the top header bar.
func (m Model) renderHeader() string {
	var parts []string

	// Title with count
	total := len(m.workspaces.Items())
	visible := len(m.ui.List.Items())

	titleText := fmt.Sprintf("%s Workspaces (%d)", m.symbols.Workspaces(), total)
	if visible != total {
		titleText = fmt.Sprintf("%s Workspaces (%d/%d)", m.symbols.Workspaces(), visible, total)
	}

	parts = append(parts, titleStyle.Render(titleText))

	// Disk usage
	if m.workspaces.TotalDiskUsage() > 0 {
		diskInfo := fmt.Sprintf("%s %s", m.symbols.Disk(), humanizeBytes(m.workspaces.TotalDiskUsage()))
		parts = append(parts, mutedTextStyle.Render(diskInfo))
	}

	// Active filters
	var filters []string

	if m.workspaces.IsStaleFilterActive() {
		filters = append(filters, badgeWarnStyle.Render("STALE"))
	}

	if m.ui.List.FilterValue() != "" {
		searchBadge := badgeInfoStyle.Render(fmt.Sprintf("%s %s", m.symbols.Search(), m.ui.List.FilterValue()))
		filters = append(filters, searchBadge)
	}

	if len(filters) > 0 {
		parts = append(parts, strings.Join(filters, " "))
	}

	if count := m.selectionCount(); count > 0 {
		parts = append(parts, mutedTextStyle.Render(fmt.Sprintf("%d selected", count)))
	}

	header := strings.Join(parts, "  ")

	// Error message if any
	if m.err != nil {
		header += "\n" + statusDirtyStyle.Render(fmt.Sprintf("%s Error: %v", m.symbols.Warning(), m.err))
	}

	// Info message if any
	if m.infoMessage != "" {
		header += "\n" + statusCleanStyle.Render(fmt.Sprintf("%s %s", m.symbols.Check(), m.infoMessage))
	}

	return header
}

// renderFooter renders the keyboard shortcuts footer.
func (m Model) renderFooter() string {
	if m.pushing {
		return subtleTextStyle.Render(fmt.Sprintf("%s Push in progress...", m.symbols.Loading()))
	}

	if m.isConfirming() {
		return ""
	}

	// Build shortcuts using configured keybindings
	searchKey := firstKey(m.ui.Keybindings.Search)
	toggleStaleKey := firstKey(m.ui.Keybindings.ToggleStale)
	detailsKey := firstKey(m.ui.Keybindings.Details)
	openKey := firstKey(m.ui.Keybindings.OpenEditor)
	syncKey := firstKey(m.ui.Keybindings.Sync)
	pushKey := firstKey(m.ui.Keybindings.Push)
	closeKey := firstKey(m.ui.Keybindings.Close)
	selectKey := firstKey(m.ui.Keybindings.Select)
	selectAllKey := firstKey(m.ui.Keybindings.SelectAll)
	deselectAllKey := firstKey(m.ui.Keybindings.DeselectAll)
	quitKey := firstKey(m.ui.Keybindings.Quit)

	var shortcuts []string

	if count := m.selectionCount(); count > 0 {
		shortcuts = append(shortcuts, accentTextStyle.Render(fmt.Sprintf("%d selected", count)))
	}

	shortcuts = append(shortcuts,
		subtleTextStyle.Render("[↑↓] navigate"),
		subtleTextStyle.Render(fmt.Sprintf("[%s] search", searchKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] stale", toggleStaleKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] details", detailsKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] open", openKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] sync", syncKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] push", pushKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] close", closeKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] select", selectKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] all", selectAllKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] none", deselectAllKey)),
		subtleTextStyle.Render(fmt.Sprintf("[%s] quit", quitKey)),
	)

	return strings.Join(shortcuts, "  •  ")
}

// renderDetailView renders the detailed workspace view.
func (m Model) renderDetailView() string {
	var b strings.Builder

	// Loading state
	detailState := m.getDetailState()
	if detailState != nil && detailState.Loading {
		b.WriteString(fmt.Sprintf("%s Loading workspace details...", m.ui.Spinner.View()))

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
	header := fmt.Sprintf("%s %s", m.symbols.Folder(), m.selectedWS.ID)
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

	// Footer with configured keys (cancel keys always exist via WithDefaults)
	b.WriteString(helpTextStyle.Render(fmt.Sprintf("Press [%s] to return", firstKey(m.ui.Keybindings.Cancel))))

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

	branchLabel := repo.Branch

	if repo.Error != "" {
		errText := strings.ReplaceAll(string(repo.Error), "\n", " ")
		statusParts = append(statusParts,
			statusDirtyStyle.Render(fmt.Sprintf("error: %s", errText)))

		branchLabel = "error"
		if repo.Error == domain.StatusErrorTimeout {
			branchLabel = "timeout"
		}
	} else {
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
	}

	if len(statusParts) == 0 {
		statusParts = append(statusParts, statusCleanStyle.Render(fmt.Sprintf("%s clean", m.symbols.Check())))
	}

	statusStr := strings.Join(statusParts, " • ")

	// Format: icon name [branch] status
	return fmt.Sprintf("  %s %-20s %s  %s",
		m.symbols.Repo(),
		repo.Name,
		subtleTextStyle.Render(fmt.Sprintf("[%s]", branchLabel)),
		statusStr)
}

// renderDetailOrphans renders the orphaned worktrees section in the detail view.
func (m Model) renderDetailOrphans() string {
	var b strings.Builder

	b.WriteString(statusWarnStyle.Render(fmt.Sprintf("%s Orphaned Worktrees", m.symbols.Warning())))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n")

	for _, orphan := range m.wsOrphans {
		line := fmt.Sprintf("  %s %-20s %s",
			m.symbols.Warning(),
			orphan.RepoName,
			statusWarnStyle.Render(orphan.ReasonDescription()))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
