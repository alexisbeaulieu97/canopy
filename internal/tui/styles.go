package tui

import "github.com/charmbracelet/lipgloss"

// Action constants for confirmation dialogs.
const (
	actionClose = "close"
	actionPush  = "push"
)

// Color palette - using a modern, accessible color scheme
var (
	// Primary colors
	colorPrimary   = lipgloss.Color("#7C3AED") // Violet
	colorSecondary = lipgloss.Color("#A78BFA") // Light violet

	// Status colors - softer, more modern tones
	colorSuccess = lipgloss.Color("#10B981") // Emerald green
	colorWarning = lipgloss.Color("#F59E0B") // Amber
	colorDanger  = lipgloss.Color("#EF4444") // Red
	colorMuted   = lipgloss.Color("#6B7280") // Gray
)

// Status indicator styles
var (
	statusCleanStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	statusDirtyStyle = lipgloss.NewStyle().
				Foreground(colorDanger).
				Bold(true)

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	statusLoadingStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)
)

// Text styles
var (
	subtleTextStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	mutedTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	boldTextStyle = lipgloss.NewStyle().
			Bold(true)

	accentTextStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)
)

// Badge styles - pill-shaped status indicators
var (
	baseBadgeStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginRight(1)

	badgeDirtyStyle = baseBadgeStyle.
			Foreground(colorDanger).
			Background(lipgloss.Color("#7F1D1D"))

	badgeWarnStyle = baseBadgeStyle.
			Foreground(colorWarning).
			Background(lipgloss.Color("#78350F"))

	badgeInfoStyle = baseBadgeStyle.
			Foreground(colorSecondary).
			Background(lipgloss.Color("#312E81"))
)

// Layout styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9FAFB"))

	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				MarginBottom(1)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Width(14)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F9FAFB"))
)

// Interactive element styles
var (
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	confirmPromptStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true).
				Padding(1, 0)

	helpTextStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true).
			MarginTop(1)
)

// Status icons (using Unicode for cross-platform support)
const (
	iconClean    = "●"
	iconDirty    = "●"
	iconWarning  = "●"
	iconLoading  = "○"
	iconError    = "✗"
	iconCursor   = "❯"
	iconNoCursor = " "
)
