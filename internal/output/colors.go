// Package output provides helpers for CLI output formatting.
package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const colorEnv = "CANOPY_COLOR"

var (
	// AccentStyle highlights headers and labels.
	AccentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#22D3EE")).Bold(true)
	// SuccessStyle highlights successful outcomes.
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	// WarningStyle highlights warnings and dry-run output.
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	// ErrorStyle highlights errors and failures.
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	// InfoStyle highlights informational messages.
	InfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8"))
	// MutedStyle de-emphasizes secondary text.
	MutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
)

// ColorEnabled returns true when color output should be used.
func ColorEnabled() bool {
	if val, ok := os.LookupEnv("NO_COLOR"); ok && strings.TrimSpace(val) != "" {
		return false
	}

	if val, ok := os.LookupEnv(colorEnv); ok {
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "0", "false", "no":
			return false
		default:
			return true
		}
	}

	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Colorize renders text with the provided style when color is enabled.
func Colorize(style lipgloss.Style, text string) string {
	if !ColorEnabled() {
		return text
	}

	return style.Render(text)
}

// Column renders a fixed-width column with optional styling.
func Column(text string, width int, style lipgloss.Style) string {
	if width <= 0 {
		return text
	}

	truncated := truncateText(text, width)

	if !ColorEnabled() {
		return fmt.Sprintf("%-*s", width, truncated)
	}

	return style.Inline(true).MaxWidth(width).Width(width).Render(truncated)
}

func truncateText(text string, width int) string {
	if width <= 0 {
		return text
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}

	return string(runes[:width])
}
