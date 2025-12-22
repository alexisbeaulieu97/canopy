package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HealthStatus represents the health state of a workspace.
type HealthStatus string

// Health status constants.
const (
	HealthError     HealthStatus = "error"
	HealthLoading   HealthStatus = "loading"
	HealthOrphaned  HealthStatus = "orphaned"
	HealthDirty     HealthStatus = "dirty"
	HealthAttention HealthStatus = "attention"
	HealthClean     HealthStatus = "clean"
)

// StatusBadge represents a workspace status indicator with icon and style.
type StatusBadge struct {
	Status HealthStatus
	Icon   string
	Style  lipgloss.Style
}

// StatusBadgeInput contains the data needed to determine workspace health status.
type StatusBadgeInput struct {
	HasError      bool
	IsLoaded      bool
	OrphanCount   int
	DirtyRepos    int
	UnpushedRepos int
	BehindRepos   int
	ErrorRepos    int
	IsStale       bool
}

// NewStatusBadge creates a StatusBadge based on the input health indicators.
func NewStatusBadge(input StatusBadgeInput) StatusBadge {
	switch {
	case input.HasError || input.ErrorRepos > 0:
		return StatusBadge{Status: HealthError, Icon: IconError, Style: StatusDirtyStyle}
	case !input.IsLoaded:
		return StatusBadge{Status: HealthLoading, Icon: IconLoading, Style: StatusLoadingStyle}
	case input.OrphanCount > 0:
		return StatusBadge{Status: HealthOrphaned, Icon: IconWarning, Style: StatusWarnStyle}
	case input.DirtyRepos > 0 || input.UnpushedRepos > 0:
		return StatusBadge{Status: HealthDirty, Icon: IconDirty, Style: StatusDirtyStyle}
	case input.IsStale || input.BehindRepos > 0:
		return StatusBadge{Status: HealthAttention, Icon: IconWarning, Style: StatusWarnStyle}
	default:
		return StatusBadge{Status: HealthClean, Icon: IconClean, Style: StatusCleanStyle}
	}
}

// Render renders the status badge icon with appropriate styling.
func (b StatusBadge) Render() string {
	return b.Style.Render(b.Icon)
}

// BadgeSet represents a collection of badges to display for a workspace.
type BadgeSet struct {
	badges []string
}

// BadgeSetInput contains the data needed to build a set of badges.
type BadgeSetInput struct {
	HasError          bool
	IsLoaded          bool
	OrphanCount       int
	OrphanCheckFailed bool
	DirtyRepos        int
	UnpushedRepos     int
	BehindRepos       int
	ErrorRepos        int
	IsStale           bool
}

// NewBadgeSet creates a BadgeSet from the input data.
func NewBadgeSet(input BadgeSetInput) BadgeSet {
	bs := BadgeSet{}

	if !input.IsLoaded && !input.HasError {
		return bs
	}

	if input.HasError {
		bs.badges = append(bs.badges, BadgeDirtyStyle.Render("ERROR"))
	}

	if input.ErrorRepos > 0 {
		text := FormatCount(input.ErrorRepos, "error", "errors")
		bs.badges = append(bs.badges, BadgeDirtyStyle.Render(text))
	}

	if orphanBadge := buildOrphanBadge(input.OrphanCount, input.OrphanCheckFailed); orphanBadge != "" {
		bs.badges = append(bs.badges, orphanBadge)
	}

	if input.DirtyRepos > 0 {
		text := fmt.Sprintf("%d dirty", input.DirtyRepos)
		bs.badges = append(bs.badges, BadgeDirtyStyle.Render(text))
	}

	if input.UnpushedRepos > 0 {
		text := fmt.Sprintf("%d unpushed", input.UnpushedRepos)
		bs.badges = append(bs.badges, BadgeDirtyStyle.Render(text))
	}

	if input.BehindRepos > 0 {
		text := fmt.Sprintf("%d behind", input.BehindRepos)
		bs.badges = append(bs.badges, BadgeWarnStyle.Render(text))
	}

	if input.IsStale {
		bs.badges = append(bs.badges, BadgeWarnStyle.Render("STALE"))
	}

	return bs
}

// Render returns the badges as a space-separated string.
func (bs BadgeSet) Render() string {
	return strings.Join(bs.badges, " ")
}

// buildOrphanBadge creates an orphan status badge if applicable.
func buildOrphanBadge(orphanCount int, checkFailed bool) string {
	if checkFailed {
		return BadgeWarnStyle.Render("⚠ orphan check failed")
	}

	if orphanCount > 0 {
		text := fmt.Sprintf("%d orphan", orphanCount)
		if orphanCount > 1 {
			text = fmt.Sprintf("%d orphans", orphanCount)
		}

		return BadgeWarnStyle.Render(text)
	}

	return ""
}

// StatusLine represents a summary line for workspace status.
type StatusLine struct {
	parts []string
}

// StatusLineInput contains the data needed to build a status line.
type StatusLineInput struct {
	DirtyRepos    int
	UnpushedRepos int
	BehindRepos   int
	ErrorRepos    int
}

// NewStatusLine creates a StatusLine from the input data.
func NewStatusLine(input StatusLineInput) StatusLine {
	sl := StatusLine{}

	if input.DirtyRepos == 0 && input.UnpushedRepos == 0 && input.BehindRepos == 0 && input.ErrorRepos == 0 {
		sl.parts = append(sl.parts, StatusCleanStyle.Render("✓ All repos clean"))
		return sl
	}

	if input.ErrorRepos > 0 {
		sl.parts = append(sl.parts, StatusDirtyStyle.Render(
			FormatCount(input.ErrorRepos, "error", "errors")))
	}

	if input.DirtyRepos > 0 {
		sl.parts = append(sl.parts, StatusDirtyStyle.Render(
			fmt.Sprintf("%d dirty", input.DirtyRepos)))
	}

	if input.UnpushedRepos > 0 {
		sl.parts = append(sl.parts, StatusDirtyStyle.Render(
			fmt.Sprintf("%d unpushed", input.UnpushedRepos)))
	}

	if input.BehindRepos > 0 {
		sl.parts = append(sl.parts, StatusWarnStyle.Render(
			fmt.Sprintf("%d behind", input.BehindRepos)))
	}

	return sl
}

// Render returns the status line as a bullet-separated string.
func (sl StatusLine) Render() string {
	return strings.Join(sl.parts, "  •  ")
}
