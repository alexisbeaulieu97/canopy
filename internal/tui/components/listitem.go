package components

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// WorkspaceItem represents a workspace in the list.
type WorkspaceItem struct {
	Workspace         domain.Workspace
	Summary           WorkspaceSummary
	OrphanCount       int
	OrphanCheckFailed bool // true if orphan detection failed for this workspace
	Err               error
	Loaded            bool
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
		Foreground(lipgloss.Color("#F9FAFB"))
	styles.SelectedTitle = styles.SelectedTitle.
		Bold(true).
		Foreground(ColorPrimary)
	styles.NormalDesc = styles.NormalDesc.
		Foreground(ColorMuted)
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(lipgloss.Color("#9CA3AF"))

	return WorkspaceDelegate{
		styles:         styles,
		staleThreshold: staleThreshold,
	}
}

// Height returns the height of each list item.
func (d WorkspaceDelegate) Height() int { return 3 }

// Spacing returns the spacing between list items.
func (d WorkspaceDelegate) Spacing() int { return 1 }

// Update handles messages for the delegate (no-op for this delegate).
func (d WorkspaceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render renders a workspace item in the list.
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

	// Get health status badge
	badge := NewStatusBadge(StatusBadgeInput{
		HasError:      wsItem.Err != nil,
		IsLoaded:      wsItem.Loaded,
		OrphanCount:   wsItem.OrphanCount,
		DirtyRepos:    wsItem.Summary.DirtyRepos,
		UnpushedRepos: wsItem.Summary.UnpushedRepos,
		BehindRepos:   wsItem.Summary.BehindRepos,
		ErrorRepos:    wsItem.Summary.ErrorRepos,
		IsStale:       wsItem.Workspace.IsStale(d.staleThreshold),
	})

	// Choose title style based on selection
	titleStyle := d.styles.NormalTitle
	if isSelected {
		titleStyle = d.styles.SelectedTitle
	}

	// Build the title line with status indicator
	statusIndicator := badge.Render()
	title := titleStyle.Render(wsItem.Workspace.ID)
	badges := NewBadgeSet(BadgeSetInput{
		HasError:          wsItem.Err != nil,
		IsLoaded:          wsItem.Loaded,
		OrphanCount:       wsItem.OrphanCount,
		OrphanCheckFailed: wsItem.OrphanCheckFailed,
		DirtyRepos:        wsItem.Summary.DirtyRepos,
		UnpushedRepos:     wsItem.Summary.UnpushedRepos,
		BehindRepos:       wsItem.Summary.BehindRepos,
		ErrorRepos:        wsItem.Summary.ErrorRepos,
		IsStale:           wsItem.Workspace.IsStale(d.staleThreshold),
	}).Render()

	// First line: cursor + status + title + badges
	line1 := fmt.Sprintf("%s %s %s %s", cursor, statusIndicator, title, badges)

	// Build description line
	descStyle := d.styles.NormalDesc
	if isSelected {
		descStyle = d.styles.SelectedDesc
	}

	var secondary string

	switch {
	case wsItem.Err != nil:
		secondary = StatusDirtyStyle.Render("⚠ Error loading status")
	case !wsItem.Loaded:
		secondary = StatusLoadingStyle.Render("⋯ Loading status...")
	default:
		lastUpdated := RelativeTime(wsItem.Workspace.LastModified)
		repoText := FormatCount(wsItem.Summary.RepoCount, "repo", "repos")
		diskSize := HumanizeBytes(wsItem.Workspace.DiskUsageBytes)
		secondary = fmt.Sprintf("%s  •  %s  •  %s", repoText, diskSize, lastUpdated)
	}

	// Third line: status summary (always render for consistent height)
	var statusLine string

	if wsItem.Loaded && wsItem.Err == nil {
		statusLine = NewStatusLine(StatusLineInput{
			DirtyRepos:    wsItem.Summary.DirtyRepos,
			UnpushedRepos: wsItem.Summary.UnpushedRepos,
			BehindRepos:   wsItem.Summary.BehindRepos,
			ErrorRepos:    wsItem.Summary.ErrorRepos,
		}).Render()
	}

	// Output with proper indentation (always 3 lines to match Height())
	_, _ = fmt.Fprintf(w, "%s\n", line1)

	_, _ = fmt.Fprintf(w, "    %s\n", descStyle.Render(secondary))
	_, _ = fmt.Fprintf(w, "    %s\n", statusLine)
}

// HumanizeBytes formats a byte count into a human-readable string.
func HumanizeBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	value := float64(size) / float64(div)

	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", value, units[exp])
}

// RelativeTime formats a time.Time into a human-friendly relative string.
func RelativeTime(t time.Time) string { //nolint:gocyclo // time formatting has many cases
	if t.IsZero() {
		return "unknown"
	}

	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}

		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}

		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours()) / 24
		if days == 1 {
			return "yesterday"
		}

		return fmt.Sprintf("%dd ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours()) / (24 * 7)
		if weeks == 1 {
			return "1 week ago"
		}

		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 2, 2006")
	}
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
