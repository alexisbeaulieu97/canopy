package formatting

import (
	"math"
	"testing"
	"time"
)

func TestBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		size int64
		opts ByteSizeOptions
		want string
	}{
		{name: "cli zero", size: 0, opts: ByteSizeOptions{Zero: "0 B", Precision: 2}, want: "0 B"},
		{name: "cli kilobytes", size: 1536, opts: ByteSizeOptions{Zero: "0 B", Precision: 2}, want: "1.50 KB"},
		{name: "tui kilobytes", size: 1536, opts: ByteSizeOptions{Zero: "0 B", Precision: 1}, want: "1.5 KB"},
		{name: "bytes", size: 512, opts: ByteSizeOptions{Zero: "0 B", Precision: 1}, want: "512 B"},
		{name: "min int64", size: math.MinInt64, opts: ByteSizeOptions{Zero: "0 B", Precision: 2}, want: "-8388608.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Bytes(tt.size, tt.opts); got != tt.want {
				t.Fatalf("Bytes(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestRelativeTimeAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	cliOpts := RelativeTimeOptions{
		Zero:          "-",
		AbsoluteAfter: 7 * 24 * time.Hour,
	}
	tuiOpts := RelativeTimeOptions{
		Zero:          "unknown",
		Compact:       true,
		Yesterday:     true,
		UseWeeks:      true,
		AbsoluteAfter: 30 * 24 * time.Hour,
	}

	tests := []struct {
		name string
		now  time.Time
		when time.Time
		opts RelativeTimeOptions
		want string
	}{
		{name: "cli zero", when: time.Time{}, opts: cliOpts, want: "-"},
		{name: "cli minutes", when: now.Add(-2 * time.Minute), opts: cliOpts, want: "2 mins ago"},
		{name: "cli days", when: now.Add(-48 * time.Hour), opts: cliOpts, want: "2 days ago"},
		{name: "tui zero", when: time.Time{}, opts: tuiOpts, want: "unknown"},
		{name: "tui minutes", when: now.Add(-2 * time.Minute), opts: tuiOpts, want: "2m ago"},
		{name: "tui yesterday", when: now.Add(-24 * time.Hour), opts: tuiOpts, want: "yesterday"},
		{name: "tui weeks", when: now.Add(-14 * 24 * time.Hour), opts: tuiOpts, want: "2 weeks ago"},
		{
			name: "calendar day arithmetic avoids near-midnight off by one",
			now:  time.Date(2026, 4, 10, 0, 30, 0, 0, time.UTC),
			when: time.Date(2026, 4, 8, 23, 30, 0, 0, time.UTC),
			opts: RelativeTimeOptions{
				Zero:     "unknown",
				UseWeeks: true,
			},
			want: "2 days ago",
		},
		{
			name: "default week cutoff is above one week",
			when: now.Add(-14 * 24 * time.Hour),
			opts: RelativeTimeOptions{
				Zero:     "unknown",
				UseWeeks: true,
			},
			want: "2 weeks ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reference := now
			if !tt.now.IsZero() {
				reference = tt.now
			}

			if got := RelativeTimeAt(reference, tt.when, tt.opts); got != tt.want {
				t.Fatalf("RelativeTimeAt() = %q, want %q", got, tt.want)
			}
		})
	}
}
