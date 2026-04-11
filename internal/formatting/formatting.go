// Package formatting provides presentation-neutral formatting helpers.
package formatting

import (
	"fmt"
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

	value := absInt64(size)
	if value < unit {
		return fmt.Sprintf("%d B", size)
	}

	precision := opts.Precision
	if precision < 0 {
		precision = 0
	}

	divisor := uint64(unit)
	unitIndex := 0
	units := []string{"KB", "MB", "GB", "TB"}

	for next := value / unit; next >= unit && unitIndex < len(units)-1; next /= unit {
		divisor *= unit
		unitIndex++
	}

	formattedValue := fmt.Sprintf("%.*f", precision, float64(value)/float64(divisor))
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

	if formatted, ok := formatSubdayRelativeTime(diff, opts); ok {
		return formatted
	}

	return formatCalendarRelativeTime(now, t, opts)
}

func formatSubdayRelativeTime(diff time.Duration, opts RelativeTimeOptions) (string, bool) {
	switch {
	case diff < time.Minute:
		return "just now", true
	case diff < time.Hour:
		return formatCount(int(diff.Minutes()), "min", "mins", opts.Compact), true
	case diff < 24*time.Hour:
		return formatCount(int(diff.Hours()), "hour", "hours", opts.Compact), true
	default:
		return "", false
	}
}

func formatCalendarRelativeTime(now, t time.Time, opts RelativeTimeOptions) string {
	days := calendarDaysBetween(now, t)
	switch {
	case days < 7:
		if days == 1 && opts.Yesterday {
			return "yesterday"
		}

		return formatCount(days, "day", "days", opts.Compact)
	case opts.UseWeeks && days < opts.absoluteAfterDays():
		weeks := days / 7
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

	return 30 * 24 * time.Hour
}

func (o RelativeTimeOptions) absoluteAfterDays() int {
	return durationDays(o.absoluteAfter())
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

func durationDays(d time.Duration) int {
	if d <= 0 {
		return 0
	}

	return int(d / (24 * time.Hour))
}

func calendarDaysBetween(now, then time.Time) int {
	if now.Before(then) {
		return 0
	}

	location := now.Location()
	nowDay := calendarDayNumber(now.In(location))

	thenDay := calendarDayNumber(then.In(location))
	if nowDay <= thenDay {
		return 0
	}

	return int(nowDay - thenDay)
}

func calendarDayNumber(t time.Time) int64 {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix() / (24 * 60 * 60)
}

func absInt64(v int64) uint64 {
	const (
		minInt64          = -1 << 63
		minInt64Magnitude = uint64(1) << 63
	)

	if v >= 0 {
		return uint64(v)
	}

	if v == minInt64 {
		return minInt64Magnitude
	}

	return uint64(-v)
}
