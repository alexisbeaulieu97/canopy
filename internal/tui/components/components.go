// Package components provides reusable TUI components for the Canopy workspace manager.
// These components are designed to be composable and testable, with clear interfaces
// between them and the main TUI model.
package components

import "github.com/charmbracelet/lipgloss"

// Semantic color palette - colors defined by meaning, not visual preference
var (
	// Core semantic colors
	ColorAccent  = lipgloss.Color("#8B5CF6") // Selection, primary actions
	ColorSuccess = lipgloss.Color("#22C55E") // Clean status, confirmations
	ColorWarning = lipgloss.Color("#F59E0B") // Stale, behind, needs sync
	ColorDanger  = lipgloss.Color("#EF4444") // Dirty, errors, destructive
	ColorMuted   = lipgloss.Color("#6B7280") // Secondary text, disabled
	ColorSubtle  = lipgloss.Color("#374151") // Borders, dividers
	ColorSurface = lipgloss.Color("#1F2937") // Panel backgrounds
	ColorText    = lipgloss.Color("#F9FAFB") // Primary text

	// Legacy aliases for backwards compatibility
	ColorPrimary   = ColorAccent
	ColorSecondary = lipgloss.Color("#A78BFA") // Light violet
)

// Spacing constants for consistent layout
const (
	SpacingXS = 1 // Minimal padding
	SpacingSM = 2 // Small padding
	SpacingMD = 3 // Medium padding
	SpacingLG = 4 // Large padding
)

// Panel/box styles for layout sections
var (
	// HeaderPanelStyle is used for the top header bar
	HeaderPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorSubtle).
				Padding(0, SpacingSM).
				MarginBottom(1)

	// ContentPanelStyle is used for the main content area
	ContentPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorSubtle).
				Padding(SpacingXS, SpacingSM)

	// FooterPanelStyle is used for the footer help bar
	FooterPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorSubtle).
				Padding(0, SpacingSM).
				MarginTop(1)

	// ModalPanelStyle is used for confirmation dialogs and overlays
	ModalPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(ColorWarning).
			Padding(SpacingXS, SpacingMD).
			Background(ColorSurface)

	// CardStyle is used for detail view sections
	CardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(SpacingXS, SpacingSM).
			MarginBottom(1)

	// SelectedRowStyle is used for highlighted/selected rows
	SelectedRowStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorAccent).
				BorderLeft(true).
				BorderRight(false).
				BorderTop(false).
				BorderBottom(false).
				PaddingLeft(1)

	// AlternateRowStyle is used for alternating row backgrounds
	AlternateRowStyle = lipgloss.NewStyle().
				Background(ColorSurface)
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
// Icons are chosen to be distinguishable by shape, not just color.
const (
	IconClean    = "✓" // Checkmark for clean/healthy state
	IconDirty    = "●" // Filled circle for dirty/needs attention
	IconWarning  = "⚠" // Warning triangle for stale/behind
	IconLoading  = "○" // Empty circle for loading
	IconError    = "✗" // X mark for errors
	IconCursor   = "❯" // Arrow for selection cursor
	IconNoCursor = " "
)
