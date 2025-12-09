package tui

import (
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// humanizeBytes formats a byte count into a human-readable string.
// Delegates to components.HumanizeBytes for consistency.
func humanizeBytes(size int64) string {
	return components.HumanizeBytes(size)
}

// relativeTime formats a time.Time into a human-friendly relative string.
// Delegates to components.RelativeTime for consistency.
func relativeTime(t time.Time) string {
	return components.RelativeTime(t)
}
