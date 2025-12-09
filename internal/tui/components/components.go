// Package components provides reusable TUI components for the Canopy workspace manager.
// These components are designed to be composable and testable, with clear interfaces
// between them and the main TUI model.
package components

import "github.com/charmbracelet/lipgloss"

// Color palette - using a modern, accessible color scheme
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // Violet
	ColorSecondary = lipgloss.Color("#A78BFA") // Light violet

	// Status colors - softer, more modern tones
	ColorSuccess = lipgloss.Color("#10B981") // Emerald green
	ColorWarning = lipgloss.Color("#F59E0B") // Amber
	ColorDanger  = lipgloss.Color("#EF4444") // Red
	ColorMuted   = lipgloss.Color("#6B7280") // Gray
)

// Status indicator styles
var (
	StatusCleanStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StatusDirtyStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Bold(true)

	StatusWarnStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	StatusLoadingStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Italic(true)
)

// Text styles
var (
	SubtleTextStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	MutedTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	BoldTextStyle = lipgloss.NewStyle().
			Bold(true)

	AccentTextStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)

// Badge styles - pill-shaped status indicators
var (
	BaseBadgeStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1)

	BadgeDirtyStyle = BaseBadgeStyle.
			Foreground(ColorDanger).
			Background(lipgloss.Color("#7F1D1D"))

	BadgeWarnStyle = BaseBadgeStyle.
			Foreground(ColorWarning).
			Background(lipgloss.Color("#78350F"))

	BadgeInfoStyle = BaseBadgeStyle.
			Foreground(ColorSecondary).
			Background(lipgloss.Color("#312E81"))
)

// Layout styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9FAFB"))

	DetailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				MarginBottom(1)

	DetailLabelStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Width(14)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F9FAFB"))
)

// Interactive element styles
var (
	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	ConfirmPromptStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true).
				Padding(1, 0)

	HelpTextStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true).
			MarginTop(1)
)

// Status icons (using Unicode for cross-platform support)
const (
	IconClean    = "●"
	IconDirty    = "●"
	IconWarning  = "●"
	IconLoading  = "○"
	IconError    = "✗"
	IconCursor   = "❯"
	IconNoCursor = " "
)
