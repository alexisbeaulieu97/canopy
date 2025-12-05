package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// workspaceItem represents a workspace in the list.
type workspaceItem struct {
	workspace         domain.Workspace
	summary           workspaceSummary
	orphanCount       int
	orphanCheckFailed bool // true if orphan detection failed for this workspace
	err               error
	loaded            bool
}

// workspaceSummary holds aggregated status info for a workspace.
type workspaceSummary struct {
	repoCount     int
	dirtyRepos    int
	unpushedRepos int
	behindRepos   int
}

// list.Item interface implementation
func (i workspaceItem) Title() string       { return i.workspace.ID }
func (i workspaceItem) Description() string { return "" }
func (i workspaceItem) FilterValue() string { return i.workspace.ID }

// workspaceDelegate handles rendering of workspace items in the list.
type workspaceDelegate struct {
	styles         list.DefaultItemStyles
	staleThreshold int
}

func newWorkspaceDelegate(staleThreshold int) workspaceDelegate {
	styles := list.NewDefaultItemStyles()
	styles.NormalTitle = styles.NormalTitle.
		Bold(true).
		Foreground(lipgloss.Color("#F9FAFB"))
	styles.SelectedTitle = styles.SelectedTitle.
		Bold(true).
		Foreground(colorPrimary)
	styles.NormalDesc = styles.NormalDesc.
		Foreground(colorMuted)
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(lipgloss.Color("#9CA3AF"))

	return workspaceDelegate{
		styles:         styles,
		staleThreshold: staleThreshold,
	}
}

func (d workspaceDelegate) Height() int  { return 3 }
func (d workspaceDelegate) Spacing() int { return 1 }

func (d workspaceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d workspaceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	wsItem, ok := listItem.(workspaceItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Build cursor indicator
	cursor := iconNoCursor
	if isSelected {
		cursor = cursorStyle.Render(iconCursor)
	}

	// Get health status
	_, statusIcon, statusStyle := healthForWorkspace(wsItem, d.staleThreshold)

	// Choose title style based on selection
	titleStyle := d.styles.NormalTitle
	if isSelected {
		titleStyle = d.styles.SelectedTitle
	}

	// Build the title line with status indicator
	statusIndicator := statusStyle.Render(statusIcon)
	title := titleStyle.Render(wsItem.workspace.ID)
	badges := renderBadges(wsItem, d.staleThreshold)

	// First line: cursor + status + title + badges
	line1 := fmt.Sprintf("%s %s %s %s", cursor, statusIndicator, title, badges)

	// Build description line
	descStyle := d.styles.NormalDesc
	if isSelected {
		descStyle = d.styles.SelectedDesc
	}

	var secondary string

	switch {
	case wsItem.err != nil:
		secondary = statusDirtyStyle.Render("⚠ Error loading status")
	case !wsItem.loaded:
		secondary = statusLoadingStyle.Render("⋯ Loading status...")
	default:
		lastUpdated := relativeTime(wsItem.workspace.LastModified)
		repoText := formatCount(wsItem.summary.repoCount, "repo", "repos")
		diskSize := humanizeBytes(wsItem.workspace.DiskUsageBytes)
		secondary = fmt.Sprintf("%s  •  %s  •  %s", repoText, diskSize, lastUpdated)
	}

	// Third line: status summary (always render for consistent height)
	var statusLine string

	if wsItem.loaded && wsItem.err == nil {
		statusLine = buildStatusLine(wsItem.summary)
	}

	// Output with proper indentation (always 3 lines to match Height())
	_, _ = fmt.Fprintf(w, "%s\n", line1)

	_, _ = fmt.Fprintf(w, "    %s\n", descStyle.Render(secondary))
	_, _ = fmt.Fprintf(w, "    %s\n", statusLine)
}

// healthForWorkspace determines the health status of a workspace.
func healthForWorkspace(item workspaceItem, staleThreshold int) (string, string, lipgloss.Style) {
	switch {
	case item.err != nil:
		return "error", iconError, statusDirtyStyle
	case !item.loaded:
		return "loading", iconLoading, statusLoadingStyle
	case item.orphanCount > 0:
		return "orphaned", iconWarning, statusWarnStyle
	case item.summary.dirtyRepos > 0 || item.summary.unpushedRepos > 0:
		return "dirty", iconDirty, statusDirtyStyle
	case item.workspace.IsStale(staleThreshold) || item.summary.behindRepos > 0:
		return "attention", iconWarning, statusWarnStyle
	default:
		return "clean", iconClean, statusCleanStyle
	}
}

// buildOrphanBadge creates an orphan status badge if applicable.
func buildOrphanBadge(item workspaceItem) string {
	if item.orphanCheckFailed {
		return badgeWarnStyle.Render("⚠ orphan check failed")
	}

	if item.orphanCount > 0 {
		text := fmt.Sprintf("%d orphan", item.orphanCount)
		if item.orphanCount > 1 {
			text = fmt.Sprintf("%d orphans", item.orphanCount)
		}

		return badgeWarnStyle.Render(text)
	}

	return ""
}

// renderBadges creates status badges for the workspace item.
func renderBadges(item workspaceItem, staleThreshold int) string {
	if !item.loaded && item.err == nil {
		return ""
	}

	var badges []string

	if item.err != nil {
		badges = append(badges, badgeDirtyStyle.Render("ERROR"))
	}

	if orphanBadge := buildOrphanBadge(item); orphanBadge != "" {
		badges = append(badges, orphanBadge)
	}

	if item.summary.dirtyRepos > 0 {
		text := fmt.Sprintf("%d dirty", item.summary.dirtyRepos)
		badges = append(badges, badgeDirtyStyle.Render(text))
	}

	if item.summary.unpushedRepos > 0 {
		text := fmt.Sprintf("%d unpushed", item.summary.unpushedRepos)
		badges = append(badges, badgeDirtyStyle.Render(text))
	}

	if item.summary.behindRepos > 0 {
		text := fmt.Sprintf("%d behind", item.summary.behindRepos)
		badges = append(badges, badgeWarnStyle.Render(text))
	}

	if item.workspace.IsStale(staleThreshold) {
		badges = append(badges, badgeWarnStyle.Render("STALE"))
	}

	return strings.Join(badges, " ")
}

// buildStatusLine creates a summary line for workspace status.
func buildStatusLine(summary workspaceSummary) string {
	if summary.dirtyRepos == 0 && summary.unpushedRepos == 0 && summary.behindRepos == 0 {
		return statusCleanStyle.Render("✓ All repos clean")
	}

	var parts []string

	if summary.dirtyRepos > 0 {
		parts = append(parts, statusDirtyStyle.Render(
			fmt.Sprintf("%d dirty", summary.dirtyRepos)))
	}

	if summary.unpushedRepos > 0 {
		parts = append(parts, statusDirtyStyle.Render(
			fmt.Sprintf("%d unpushed", summary.unpushedRepos)))
	}

	if summary.behindRepos > 0 {
		parts = append(parts, statusWarnStyle.Render(
			fmt.Sprintf("%d behind", summary.behindRepos)))
	}

	return strings.Join(parts, "  •  ")
}

// summarizeStatus creates a summary from workspace status.
func summarizeStatus(status *domain.WorkspaceStatus) workspaceSummary {
	summary := workspaceSummary{
		repoCount: len(status.Repos),
	}

	for _, repo := range status.Repos {
		if repo.IsDirty {
			summary.dirtyRepos++
		}

		if repo.UnpushedCommits > 0 {
			summary.unpushedRepos++
		}

		if repo.BehindRemote > 0 {
			summary.behindRepos++
		}
	}

	return summary
}
