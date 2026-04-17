package components

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/formatting"
)

// WorkspaceItem represents a workspace in the list.
type WorkspaceItem struct {
	Workspace         domain.Workspace
	Summary           WorkspaceSummary
	OrphanCount       int
	OrphanCheckFailed bool // true if orphan detection failed for this workspace
	LockCheckFailed   bool // true if lock detection failed for this workspace
	Err               error
	Loaded            bool
	Selected          bool
}

// WorkspaceSummary holds aggregated status info for a workspace.
type WorkspaceSummary struct {
	RepoCount     int
	DirtyRepos    int
	UnpushedRepos int
	BehindRepos   int
	ErrorRepos    int
}

// Title returns the workspace ID as the list item title.
func (i WorkspaceItem) Title() string { return i.Workspace.ID }

// Description returns an empty string (description is rendered separately).
func (i WorkspaceItem) Description() string { return "" }

// FilterValue returns the workspace ID for filtering.
func (i WorkspaceItem) FilterValue() string { return i.Workspace.ID }

// SummarizeStatus creates a WorkspaceSummary from domain.WorkspaceStatus.
// Returns an empty summary if status is nil.
func SummarizeStatus(status *domain.WorkspaceStatus) WorkspaceSummary {
	if status == nil {
		return WorkspaceSummary{}
	}

	summary := WorkspaceSummary{
		RepoCount: len(status.Repos),
	}

	for _, repo := range status.Repos {
		if repo.Error != "" {
			summary.ErrorRepos++
		}

		if repo.IsDirty {
			summary.DirtyRepos++
		}

		if repo.UnpushedCommits > 0 {
			summary.UnpushedRepos++
		}

		if repo.BehindRemote > 0 {
			summary.BehindRepos++
		}
	}

	return summary
}

// secondaryLineIndent aligns the secondary line with the title.
// Layout: cursor(1) + space(1) + selection(3) + space(1) = 6 chars
const secondaryLineIndent = "      "

// WorkspaceDelegate handles rendering of workspace items in the list.
type WorkspaceDelegate struct {
	styles         list.DefaultItemStyles
	staleThreshold int
}

// NewWorkspaceDelegate creates a new WorkspaceDelegate with the given stale threshold.
func NewWorkspaceDelegate(staleThreshold int) WorkspaceDelegate {
	styles := list.NewDefaultItemStyles()
	styles.NormalTitle = styles.NormalTitle.
		Bold(true).
		Foreground(ColorText)
	styles.SelectedTitle = styles.SelectedTitle.
		Bold(true).
		Foreground(ColorAccent)
	styles.NormalDesc = styles.NormalDesc.
		Foreground(ColorMuted)
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(lipgloss.Color("#9CA3AF"))

	return WorkspaceDelegate{
		styles:         styles,
		staleThreshold: staleThreshold,
	}
}

// Height returns the height of each list item (2 lines for compact layout).
func (d WorkspaceDelegate) Height() int { return 2 }

// Spacing returns the spacing between list items.
func (d WorkspaceDelegate) Spacing() int { return 0 }

// Update handles messages for the delegate (no-op for this delegate).
func (d WorkspaceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// buildStatusPills creates inline status indicators for a workspace item.
func (d WorkspaceDelegate) buildStatusPills(wsItem WorkspaceItem) string {
	var statusPills []string

	switch {
	case wsItem.Err != nil:
		statusPills = append(statusPills, StatusDirtyStyle.Render(IconError))
	case !wsItem.Loaded:
		statusPills = append(statusPills, StatusLoadingStyle.Render(IconLoading))
	default:
		statusPills = d.buildLoadedStatusPills(wsItem)
	}

	if len(statusPills) > 0 {
		return " " + strings.Join(statusPills, " ")
	}

	return ""
}

// buildLoadedStatusPills creates status pills for a successfully loaded workspace.
func (d WorkspaceDelegate) buildLoadedStatusPills(wsItem WorkspaceItem) []string {
	var pills []string

	if wsItem.Summary.DirtyRepos > 0 {
		pills = append(pills, StatusDirtyStyle.Render(fmt.Sprintf("%s%d", IconDirty, wsItem.Summary.DirtyRepos)))
	}

	if wsItem.Summary.UnpushedRepos > 0 {
		pills = append(pills, StatusDirtyStyle.Render(fmt.Sprintf("%s%d", IconUnpushed, wsItem.Summary.UnpushedRepos)))
	}

	if wsItem.Summary.BehindRepos > 0 {
		pills = append(pills, StatusWarnStyle.Render(fmt.Sprintf("%s%d", IconBehind, wsItem.Summary.BehindRepos)))
	}

	if wsItem.Summary.ErrorRepos > 0 {
		pills = append(pills, StatusDirtyStyle.Render(fmt.Sprintf("%s%d", IconError, wsItem.Summary.ErrorRepos)))
	}

	if wsItem.LockCheckFailed {
		pills = append(pills, StatusWarnStyle.Render("lock?"))
	}

	// Show orphan check failure warning
	if wsItem.OrphanCheckFailed {
		pills = append(pills, StatusWarnStyle.Render(fmt.Sprintf("%s?", IconWarning)))
	} else if wsItem.OrphanCount > 0 {
		pills = append(pills, StatusWarnStyle.Render(fmt.Sprintf("%s%d", IconWarning, wsItem.OrphanCount)))
	}

	if wsItem.Workspace.IsStale(d.staleThreshold) {
		pills = append(pills, StatusWarnStyle.Render("stale"))
	}

	return pills
}

// buildSecondaryLine creates the description line for a workspace item.
func buildSecondaryLine(wsItem WorkspaceItem) string {
	switch {
	case wsItem.Err != nil:
		return StatusDirtyStyle.Render("Error loading status")
	case !wsItem.Loaded:
		return StatusLoadingStyle.Render("Loading...")
	default:
		repoText := FormatCount(wsItem.Summary.RepoCount, "repo", "repos")
		diskSize := formatting.Bytes(wsItem.Workspace.DiskUsageBytes, formatting.ByteSizeOptions{
			Zero:      "0 B",
			Precision: 1,
		})
		lastUpdated := formatting.RelativeTime(wsItem.Workspace.LastModified, formatting.RelativeTimeOptions{
			Zero:          "unknown",
			Compact:       true,
			Yesterday:     true,
			UseWeeks:      true,
			AbsoluteAfter: 30 * 24 * time.Hour,
		})

		return fmt.Sprintf("%s • %s • %s", repoText, diskSize, lastUpdated)
	}
}

// Render renders a workspace item in the list using a two-line compact layout.
func (d WorkspaceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	wsItem, ok := listItem.(WorkspaceItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Build cursor indicator
	cursor := IconNoCursor
	if isSelected {
		cursor = CursorStyle.Render(IconCursor)
	}

	// Selection checkbox
	selectionIndicator := "[ ]"
	selectionStyle := SubtleTextStyle

	if wsItem.Selected {
		selectionIndicator = "[x]"
		selectionStyle = AccentTextStyle
	}

	// Choose title style based on selection
	titleStyle := d.styles.NormalTitle
	if isSelected {
		titleStyle = d.styles.SelectedTitle
	}

	// Title with inline status pills
	title := titleStyle.Render(wsItem.Workspace.ID)
	pillsStr := d.buildStatusPills(wsItem)

	// First line: cursor + selection + title + inline status pills
	line1 := fmt.Sprintf("%s %s %s%s", cursor, selectionStyle.Render(selectionIndicator), title, pillsStr)

	// Build description line (second line)
	descStyle := d.styles.NormalDesc
	if isSelected {
		descStyle = d.styles.SelectedDesc
	}

	secondary := buildSecondaryLine(wsItem)

	// Output with proper indentation (2 lines for compact layout)
	_, _ = fmt.Fprintf(w, "%s\n", line1)
	_, _ = fmt.Fprintf(w, "%s%s\n", secondaryLineIndent, descStyle.Render(secondary))
}

// Pluralize returns the singular or plural form based on count.
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}

// FormatCount formats a count with its label.
func FormatCount(count int, singular, plural string) string {
	return fmt.Sprintf("%d %s", count, Pluralize(count, singular, plural))
}
