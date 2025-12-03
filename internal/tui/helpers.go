package tui

import (
	"fmt"
	"time"
)

// humanizeBytes formats a byte count into a human-readable string.
func humanizeBytes(size int64) string {
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

// relativeTime formats a time.Time into a human-friendly relative string.
func relativeTime(t time.Time) string { //nolint:gocyclo // time formatting has many cases
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

// pluralize returns the singular or plural form based on count.
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}

	return plural
}

// formatCount formats a count with its label.
func formatCount(count int, singular, plural string) string {
	return fmt.Sprintf("%d %s", count, pluralize(count, singular, plural))
}
