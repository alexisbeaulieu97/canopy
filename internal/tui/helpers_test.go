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
