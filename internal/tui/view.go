package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/formatting"
	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// Layout constants for consistent rendering.
const (
	separatorWidth = 40 // Width of horizontal dividers
)

// formatKeyAction formats a keyboard shortcut as key:action.
func formatKeyAction(key, action string) string {
	return components.SubtleTextStyle.Render(fmt.Sprintf("%s:%s", components.AccentTextStyle.Render(key), action))
}

// View renders the UI by delegating to the current view state.
func (m Model) View() string {
	return m.viewState.View(&m)
}

// renderListView renders the main workspace list view.
func (m Model) renderListView() string {
	var b strings.Builder

	// Header panel
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Show spinner if pushing
	if m.pushing {
		spinnerLine := fmt.Sprintf("  %s Pushing %s...",
			m.ui.Spinner.View(),
			components.AccentTextStyle.Render(m.pushTarget))
		b.WriteString(spinnerLine)
		b.WriteString("\n")
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
			components.AccentTextStyle.Render(m.pushTarget))
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

func (m Model) renderDetailViewWithConfirm(state *ConfirmViewState) string {
	var b strings.Builder

	dialog := components.ConfirmDialog{
		Active:      true,
		Action:      state.Action,
		TargetLabel: m.confirmTargetLabel(state),
	}
	b.WriteString(dialog.Render())
	b.WriteString("\n\n")
	b.WriteString(m.renderDetailView())

	return b.String()
}

func (m Model) renderConfirmView(state *ConfirmViewState) string {
	if state == nil {
		return m.renderListView()
	}

	if _, ok := state.Parent.(*DetailViewState); ok {
		return m.renderDetailViewWithConfirm(state)
	}

	return m.renderListViewWithConfirm(state)
}

func (m Model) confirmTargetLabel(state *ConfirmViewState) string {
	if state == nil || len(state.TargetIDs) == 0 {
		return ""
	}

	if len(state.TargetIDs) == 1 {
		return fmt.Sprintf("workspace %s", components.AccentTextStyle.Render(state.TargetIDs[0]))
	}

	return fmt.Sprintf("%s workspaces", components.AccentTextStyle.Render(strconv.Itoa(len(state.TargetIDs))))
}

// renderHeader renders the top header bar with logo, stats, and breadcrumb.
func (m Model) renderHeader() string {
	var b strings.Builder

	// Title line with logo and stats
	total := len(m.workspaces.Items())
	visible := len(m.ui.List.Items())

	// Logo and title
	logo := components.AccentTextStyle.Render(m.symbols.Workspaces())
	title := components.TitleStyle.Render("Workspaces")

	// Count badge
	var countText string
	if visible != total {
		countText = components.SubtleTextStyle.Render(fmt.Sprintf("(%d/%d)", visible, total))
	} else {
		countText = components.SubtleTextStyle.Render(fmt.Sprintf("(%d)", total))
	}

	fmt.Fprintf(&b, "%s %s %s", logo, title, countText)

	// Disk usage stat
	if m.workspaces.TotalDiskUsage() > 0 {
		diskInfo := fmt.Sprintf("  %s %s", m.symbols.Disk(), formatting.Bytes(m.workspaces.TotalDiskUsage(), formatting.ByteSizeOptions{
			Zero:      "0 B",
			Precision: 1,
		}))
		b.WriteString(components.MutedTextStyle.Render(diskInfo))
	}

	// Active filters and selection on same line
	var indicators []string

	if m.workspaces.IsStaleFilterActive() {
		indicators = append(indicators, components.BadgeWarnStyle.Render(fmt.Sprintf("%s STALE", m.symbols.Stale())))
	}

	if m.ui.List.FilterValue() != "" {
		searchBadge := components.BadgeInfoStyle.Render(fmt.Sprintf("%s %s", m.symbols.Search(), m.ui.List.FilterValue()))
		indicators = append(indicators, searchBadge)
	}

	if count := m.selectionCount(); count > 0 {
		indicators = append(indicators, components.AccentTextStyle.Render(fmt.Sprintf("%d selected", count)))
	}

	if len(indicators) > 0 {
		b.WriteString("  ")
		b.WriteString(strings.Join(indicators, " "))
	}

	// Error message on new line if any
	if m.err != nil {
		b.WriteString("\n")

		errMsg := fmt.Sprintf("%s Error: %v", m.symbols.Error(), m.err)
		b.WriteString(components.StatusDirtyStyle.Render(errMsg))
	}

	// Info message on new line if any
	if m.infoMessage != "" {
		b.WriteString("\n")

		infoMsg := fmt.Sprintf("%s %s", m.symbols.Check(), m.infoMessage)
		b.WriteString(components.StatusCleanStyle.Render(infoMsg))
	}

	return b.String()
}

// renderFooter renders the keyboard shortcuts footer.
func (m Model) renderFooter() string {
	if m.pushing {
		return components.SubtleTextStyle.Render(fmt.Sprintf("  %s Push in progress...", m.symbols.Loading()))
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

	shortcuts = append(shortcuts,
		formatKeyAction("↑↓", "nav"),
		formatKeyAction(searchKey, "search"),
		formatKeyAction(toggleStaleKey, "stale"),
		formatKeyAction(detailsKey, "details"),
		formatKeyAction(openKey, "open"),
		formatKeyAction(syncKey, "sync"),
		formatKeyAction(pushKey, "push"),
		formatKeyAction(closeKey, "close"),
		formatKeyAction(selectKey, "sel"),
		formatKeyAction(selectAllKey, "all"),
		formatKeyAction(deselectAllKey, "none"),
		formatKeyAction(quitKey, "quit"),
	)

	return strings.Join(shortcuts, "  ")
}

// renderDetailView renders the detailed workspace view.
func (m Model) renderDetailView() string {
	var b strings.Builder

	// Loading state
	detailState := m.getDetailState()
	if detailState != nil && detailState.Loading {
		fmt.Fprintf(&b, "  %s Loading workspace details...", m.ui.Spinner.View())

		return b.String()
	}

	// No workspace selected
	if m.selectedWS == nil {
		b.WriteString(components.SubtleTextStyle.Render("No workspace selected."))
		b.WriteString("\n\n")
		b.WriteString(components.HelpTextStyle.Render("Press [esc] to return to the list"))

		return b.String()
	}

	// Breadcrumb: Workspaces > WorkspaceID
	breadcrumb := fmt.Sprintf("%s %s %s %s",
		m.symbols.Workspaces(),
		components.SubtleTextStyle.Render("Workspaces"),
		components.SubtleTextStyle.Render(">"),
		components.AccentTextStyle.Render(m.selectedWS.ID))
	b.WriteString(breadcrumb)
	b.WriteString("\n\n")

	// Metadata card
	b.WriteString(m.renderDetailMetadata())
	b.WriteString("\n")

	// Orphans section (if any)
	if len(m.wsOrphans) > 0 {
		b.WriteString(m.renderDetailOrphans())
		b.WriteString("\n")
	}

	// Repos section
	b.WriteString(m.renderDetailRepos())
	b.WriteString("\n")

	// Footer with configured keys
	cancelKey := firstKey(m.ui.Keybindings.Cancel)
	openKey := firstKey(m.ui.Keybindings.OpenEditor)
	syncKey := firstKey(m.ui.Keybindings.Sync)
	pushKey := firstKey(m.ui.Keybindings.Push)
	closeKey := firstKey(m.ui.Keybindings.Close)

	shortcuts := []string{
		formatKeyAction(cancelKey, "back"),
		formatKeyAction(openKey, "open"),
		formatKeyAction(syncKey, "sync"),
		formatKeyAction(pushKey, "push"),
		formatKeyAction(closeKey, "close"),
	}

	b.WriteString(strings.Join(shortcuts, "  "))

	return b.String()
}

// renderDetailMetadata renders workspace metadata in the detail view.
func (m Model) renderDetailMetadata() string {
	var rows []string

	// Section header
	rows = append(rows, components.BoldTextStyle.Render("Workspace Info"))
	rows = append(rows, strings.Repeat("─", separatorWidth))

	// Branch
	row := fmt.Sprintf("  %s %-12s %s",
		m.symbols.Branch(),
		components.DetailLabelStyle.Render("Branch"),
		components.DetailValueStyle.Render(m.selectedWS.BranchName))
	rows = append(rows, row)

	// Disk usage
	row = fmt.Sprintf("  %s %-12s %s",
		m.symbols.Disk(),
		components.DetailLabelStyle.Render("Disk"),
		components.DetailValueStyle.Render(formatting.Bytes(m.selectedWS.DiskUsageBytes, formatting.ByteSizeOptions{
			Zero:      "0 B",
			Precision: 1,
		})))
	rows = append(rows, row)

	// Last modified
	row = fmt.Sprintf("  %s %-12s %s",
		m.symbols.Time(),
		components.DetailLabelStyle.Render("Modified"),
		components.DetailValueStyle.Render(formatting.RelativeTime(m.selectedWS.LastModified, formatting.RelativeTimeOptions{
			Zero:          "unknown",
			Compact:       true,
			Yesterday:     true,
			UseWeeks:      true,
			AbsoluteAfter: 30 * 24 * time.Hour,
		})))
	rows = append(rows, row)

	// Repo count
	if m.wsStatus != nil {
		repoCount := len(m.wsStatus.Repos)
		row = fmt.Sprintf("  %s %-12s %s",
			m.symbols.Repo(),
			components.DetailLabelStyle.Render("Repos"),
			components.DetailValueStyle.Render(fmt.Sprintf("%d", repoCount)))
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// renderDetailRepos renders the repository list in the detail view.
func (m Model) renderDetailRepos() string {
	var b strings.Builder

	b.WriteString(components.BoldTextStyle.Render("Repositories"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", separatorWidth))
	b.WriteString("\n")

	if m.wsStatus == nil || len(m.wsStatus.Repos) == 0 {
		b.WriteString(components.SubtleTextStyle.Render("  No repositories in this workspace."))
		b.WriteString("\n")
		b.WriteString(components.SubtleTextStyle.Render(fmt.Sprintf("  Add one with: canopy workspace repo add %s <repo>", m.selectedWS.ID)))

		return b.String()
	}

	for _, repo := range m.wsStatus.Repos {
		b.WriteString(m.renderRepoLine(repo))
		b.WriteString("\n")
	}

	return b.String()
}

// renderRepoLine renders a single repository line with status columns.
func (m Model) renderRepoLine(repo domain.RepoStatus) string {
	var statusParts []string

	branchLabel := repo.Branch

	if repo.Error != "" {
		errText := strings.ReplaceAll(string(repo.Error), "\n", " ")
		statusParts = append(statusParts,
			components.StatusDirtyStyle.Render(fmt.Sprintf("%s %s", m.symbols.Error(), errText)))

		branchLabel = "error"
		if repo.Error == domain.StatusErrorTimeout {
			branchLabel = "timeout"
		}
	} else {
		if repo.IsDirty {
			statusParts = append(statusParts, components.StatusDirtyStyle.Render(fmt.Sprintf("%s dirty", m.symbols.Dirty())))
		}

		if repo.UnpushedCommits > 0 {
			statusParts = append(statusParts,
				components.StatusDirtyStyle.Render(fmt.Sprintf("%s %d", m.symbols.Unpushed(), repo.UnpushedCommits)))
		}

		if repo.BehindRemote > 0 {
			statusParts = append(statusParts,
				components.StatusWarnStyle.Render(fmt.Sprintf("%s %d", m.symbols.Behind(), repo.BehindRemote)))
		}
	}

	if len(statusParts) == 0 {
		statusParts = append(statusParts, components.StatusCleanStyle.Render(m.symbols.Check()))
	}

	statusStr := strings.Join(statusParts, " ")

	// Format: icon name [branch] status
	return fmt.Sprintf("  %s %-18s %s  %s",
		m.symbols.Repo(),
		repo.Name,
		components.SubtleTextStyle.Render(fmt.Sprintf("[%s]", branchLabel)),
		statusStr)
}

// renderDetailOrphans renders the orphaned worktrees section in the detail view.
func (m Model) renderDetailOrphans() string {
	var b strings.Builder

	// Warning banner header
	b.WriteString(components.StatusWarnStyle.Render(fmt.Sprintf("%s Orphaned Worktrees (%d)", m.symbols.Warning(), len(m.wsOrphans))))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", separatorWidth))
	b.WriteString("\n")

	for _, orphan := range m.wsOrphans {
		line := fmt.Sprintf("  %s %-18s %s",
			m.symbols.Warning(),
			orphan.RepoName,
			components.StatusWarnStyle.Render(orphan.ReasonDescription()))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
