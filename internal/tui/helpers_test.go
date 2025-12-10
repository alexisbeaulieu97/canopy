package tui

import "testing"

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		name string
		val  int64
		want string
	}{
		{name: "zero", val: 0, want: "0 B"},
		{name: "bytes", val: 512, want: "512 B"},
		{name: "just under kb", val: 1023, want: "1023 B"},
		{name: "kilobyte", val: 1024, want: "1.0 KB"},
		{name: "mixed kb", val: 1536, want: "1.5 KB"},
		{name: "megabyte", val: 1024 * 1024, want: "1.0 MB"},
		{name: "gigabyte", val: 1024 * 1024 * 1024, want: "1.0 GB"},
		{name: "terabyte cap", val: 1024 * 1024 * 1024 * 1024 * 5, want: "5.0 TB"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := humanizeBytes(tt.val); got != tt.want {
				t.Fatalf("humanizeBytes(%d) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func TestMatchesKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		bindings []string
		want     bool
	}{
		{name: "exact match", key: "q", bindings: []string{"q"}, want: true},
		{name: "match in list", key: "q", bindings: []string{"x", "q", "ctrl+c"}, want: true},
		{name: "no match", key: "q", bindings: []string{"x", "y", "z"}, want: false},
		{name: "empty bindings", key: "q", bindings: []string{}, want: false},
		{name: "ctrl key match", key: "ctrl+c", bindings: []string{"q", "ctrl+c"}, want: true},
		{name: "case sensitive no match", key: "Q", bindings: []string{"q"}, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesKey(tt.key, tt.bindings); got != tt.want {
				t.Fatalf("matchesKey(%q, %v) = %v, want %v", tt.key, tt.bindings, got, tt.want)
			}
		})
	}
}

func TestFirstKey(t *testing.T) {
	tests := []struct {
		name     string
		bindings []string
		want     string
	}{
		{name: "single key", bindings: []string{"q"}, want: "q"},
		{name: "multiple keys", bindings: []string{"q", "ctrl+c"}, want: "q"},
		{name: "empty bindings", bindings: []string{}, want: ""},
		{name: "nil bindings", bindings: nil, want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := firstKey(tt.bindings); got != tt.want {
				t.Fatalf("firstKey(%v) = %q, want %q", tt.bindings, got, tt.want)
			}
		})
	}
}
