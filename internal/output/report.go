// Package output provides helpers for CLI output formatting.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// WriteIndentedJSON writes the value as indented JSON to the provided writer.
func WriteIndentedJSON(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(v)
}

// WriteReportHeader writes a report title and divider.
func WriteReportHeader(w io.Writer, title string, width int) {
	_, _ = fmt.Fprintln(w, title)
	_, _ = fmt.Fprintln(w, HorizontalRule(width))
	_, _ = fmt.Fprintln(w)
}

// WriteReportSummary writes a report summary with a divider.
func WriteReportSummary(w io.Writer, summary string, width int) {
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, HorizontalRule(width))
	_, _ = fmt.Fprintf(w, "Summary: %s\n", summary)
}

// FormatStatusRow formats a single icon/name/status row for CLI output.
func FormatStatusRow(name string, nameWidth int, status string, style lipgloss.Style, icon string) string {
	return fmt.Sprintf("%s %-*s %s", Colorize(style, icon), nameWidth, name, status)
}

// FormatStatusSummary formats a summary line with a leading status icon.
func FormatStatusSummary(label string, parts []string, style lipgloss.Style, icon string) string {
	return fmt.Sprintf("%s %s: %s", Colorize(style, icon), label, strings.Join(parts, ", "))
}

// SanitizeInlineMessage collapses whitespace and truncates long inline messages.
func SanitizeInlineMessage(text string, maxRunes int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	if maxRunes <= 0 {
		return text
	}

	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}

	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}

	return string(runes[:maxRunes-3]) + "..."
}
