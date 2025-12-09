package components

import (
	"strings"
	"testing"
	"time"
)

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"small bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1024 * 1024, "1.0 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GB"},
		{"fractional MB", 1536 * 1024, "1.5 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HumanizeBytes(tt.size)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		contains string
	}{
		{"zero time", time.Time{}, "unknown"},
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"minutes ago", now.Add(-5 * time.Minute), "m ago"},
		{"hours ago", now.Add(-3 * time.Hour), "h ago"},
		{"yesterday", now.Add(-25 * time.Hour), "yesterday"},
		{"days ago", now.Add(-4 * 24 * time.Hour), "d ago"},
		{"weeks ago", now.Add(-14 * 24 * time.Hour), "weeks ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RelativeTime(tt.time)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %s", tt.contains, result)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count    int
		singular string
		plural   string
		expected string
	}{
		{0, "item", "items", "items"},
		{1, "item", "items", "item"},
		{2, "item", "items", "items"},
		{1, "repo", "repos", "repo"},
		{5, "repo", "repos", "repos"},
	}

	for _, tt := range tests {
		result := Pluralize(tt.count, tt.singular, tt.plural)
		if result != tt.expected {
			t.Errorf("Pluralize(%d, %s, %s) = %s, want %s",
				tt.count, tt.singular, tt.plural, result, tt.expected)
		}
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		count    int
		singular string
		plural   string
		expected string
	}{
		{0, "item", "items", "0 items"},
		{1, "item", "items", "1 item"},
		{2, "item", "items", "2 items"},
		{5, "repo", "repos", "5 repos"},
	}

	for _, tt := range tests {
		result := FormatCount(tt.count, tt.singular, tt.plural)
		if result != tt.expected {
			t.Errorf("FormatCount(%d, %s, %s) = %s, want %s",
				tt.count, tt.singular, tt.plural, result, tt.expected)
		}
	}
}
