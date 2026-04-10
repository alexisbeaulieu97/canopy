package components

import "testing"

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
