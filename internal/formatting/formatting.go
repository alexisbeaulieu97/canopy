// Package formatting provides presentation-neutral formatting helpers.
package formatting

import (
	"fmt"
	"math"
	"time"
)

// ByteSizeOptions controls byte-size string rendering.
type ByteSizeOptions struct {
	Zero      string
	Precision int
}

// RelativeTimeOptions controls relative-time string rendering.
type RelativeTimeOptions struct {
	Zero          string
	Compact       bool
	Yesterday     bool
	UseWeeks      bool
	AbsoluteAfter time.Duration
}

// Bytes formats a byte count as a human-readable string.
func Bytes(size int64, opts ByteSizeOptions) string {
	if size == 0 && opts.Zero != "" {
		return opts.Zero
	}

	const unit = 1024
	if absInt64(size) < unit {
		return fmt.Sprintf("%d B", size)
	}

	precision := opts.Precision
	if precision < 0 {
		precision = 0
	}

	value := absInt64(size)
	divisor := int64(unit)
	unitIndex := 0
	units := []string{"KB", "MB", "GB", "TB"}

	for next := value / unit; next >= unit && unitIndex < len(units)-1; next /= unit {
		divisor *= unit
		unitIndex++
	}

	formattedValue := fmt.Sprintf("%.*f", precision, math.Abs(float64(size))/float64(divisor))
	if size < 0 {
		formattedValue = "-" + formattedValue
	}

	return fmt.Sprintf("%s %s", formattedValue, units[unitIndex])
}

// RelativeTime formats a time relative to the current time.
func RelativeTime(t time.Time, opts RelativeTimeOptions) string {
	return RelativeTimeAt(time.Now(), t, opts)
}

// RelativeTimeAt formats a time relative to a fixed reference time.
func RelativeTimeAt(now, t time.Time, opts RelativeTimeOptions) string {
	if t.IsZero() {
		return opts.Zero
	}

	diff := now.Sub(t)
	if diff < 0 {
		diff = 0
	}

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return formatCount(minutes, "min", "mins", opts.Compact)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return formatCount(hours, "hour", "hours", opts.Compact)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 && opts.Yesterday {
			return "yesterday"
		}

		return formatCount(days, "day", "days", opts.Compact)
	case opts.UseWeeks && diff < opts.absoluteAfter():
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}

		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func (o RelativeTimeOptions) absoluteAfter() time.Duration {
	if o.AbsoluteAfter > 0 {
		return o.AbsoluteAfter
	}

	return 7 * 24 * time.Hour
}

func formatCount(count int, singular, plural string, compact bool) string {
	if count == 1 {
		if compact {
			return fmt.Sprintf("1%s ago", singular[:1])
		}

		return fmt.Sprintf("1 %s ago", singular)
	}

	if compact {
		return fmt.Sprintf("%d%s ago", count, singular[:1])
	}

	return fmt.Sprintf("%d %s ago", count, plural)
}

func absInt64(v int64) int64 {
	if v < 0 {
		return -v
	}

	return v
}
