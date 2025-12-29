package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     ProgressOptions
		wantTTY  bool
		wantOpts ProgressOptions
	}{
		{
			name: "default options with buffer",
			opts: ProgressOptions{
				Total: 10,
			},
			wantTTY: false,
			wantOpts: ProgressOptions{
				Total: 10,
				Width: 40, // default width
			},
		},
		{
			name: "custom width",
			opts: ProgressOptions{
				Total: 5,
				Width: 60,
			},
			wantTTY: false,
			wantOpts: ProgressOptions{
				Total: 5,
				Width: 60,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			tt.opts.Writer = buf
			p := NewProgress(tt.opts)

			if p.IsTTY() != tt.wantTTY {
				t.Errorf("IsTTY() = %v, want %v", p.IsTTY(), tt.wantTTY)
			}

			if p.opts.Width != tt.wantOpts.Width {
				t.Errorf("Width = %v, want %v", p.opts.Width, tt.wantOpts.Width)
			}

			if p.opts.Total != tt.wantOpts.Total {
				t.Errorf("Total = %v, want %v", p.opts.Total, tt.wantOpts.Total)
			}
		})
	}
}

func TestProgress_Increment(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	opts := ProgressOptions{
		Total:     3,
		ShowCount: true,
		Writer:    buf,
	}
	p := NewProgress(opts)

	p.Increment("first")

	if p.Current() != 1 {
		t.Errorf("Current() = %v, want 1", p.Current())
	}

	p.Increment("second")

	if p.Current() != 2 {
		t.Errorf("Current() = %v, want 2", p.Current())
	}

	p.Increment("third")

	if p.Current() != 3 {
		t.Errorf("Current() = %v, want 3", p.Current())
	}

	output := buf.String()
	if !strings.Contains(output, "[1/3]") {
		t.Errorf("output should contain [1/3], got: %s", output)
	}

	if !strings.Contains(output, "[2/3]") {
		t.Errorf("output should contain [2/3], got: %s", output)
	}

	if !strings.Contains(output, "[3/3]") {
		t.Errorf("output should contain [3/3], got: %s", output)
	}
}

func TestProgress_SetCurrent(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	opts := ProgressOptions{
		Total:  10,
		Writer: buf,
	}
	p := NewProgress(opts)

	p.SetCurrent(5, "halfway")

	if p.Current() != 5 {
		t.Errorf("Current() = %v, want 5", p.Current())
	}

	p.SetCurrent(10, "done")

	if p.Current() != 10 {
		t.Errorf("Current() = %v, want 10", p.Current())
	}
}

func TestProgress_SetMessage(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	opts := ProgressOptions{
		Total:  5,
		Writer: buf,
	}
	p := NewProgress(opts)

	// SetMessage stores the message but doesn't render in non-TTY mode
	// (to avoid duplicate lines when Increment follows)
	p.SetMessage("processing item")

	// Message appears when we increment
	p.Increment("")

	output := buf.String()

	if !strings.Contains(output, "processing item") {
		t.Errorf("output should contain message, got: %s", output)
	}
}

func TestProgress_Finish(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	opts := ProgressOptions{
		Total:  3,
		Writer: buf,
	}
	p := NewProgress(opts)

	p.Increment("one")
	p.Increment("two")
	p.Finish()

	// After finish, increment should not add more output
	outputLen := buf.Len()

	p.Increment("should not appear")

	if buf.Len() != outputLen {
		t.Errorf("Finish() should prevent further output")
	}

	// Multiple Finish calls should be safe
	p.Finish()
}

func TestProgress_Cancel(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	opts := ProgressOptions{
		Total:  5,
		Writer: buf,
	}
	p := NewProgress(opts)

	p.Increment("one")
	p.Cancel()

	output := buf.String()
	if !strings.Contains(output, "Cancelled") {
		t.Errorf("Cancel() should print Cancelled, got: %s", output)
	}

	// After cancel, further operations should be no-ops
	outputLen := buf.Len()

	p.Increment("should not appear")

	if buf.Len() != outputLen {
		t.Errorf("Cancel() should prevent further output")
	}
}

func TestDefaultProgressOptions(t *testing.T) {
	t.Parallel()

	opts := DefaultProgressOptions(10)

	if opts.Total != 10 {
		t.Errorf("Total = %v, want 10", opts.Total)
	}

	if opts.Width != 40 {
		t.Errorf("Width = %v, want 40", opts.Width)
	}

	if !opts.ShowPercentage {
		t.Error("ShowPercentage should be true")
	}

	if !opts.ShowCount {
		t.Error("ShowCount should be true")
	}
}

func TestProgress_NonTTYOutput(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	opts := ProgressOptions{
		Total:     3,
		ShowCount: true,
		Writer:    buf,
	}
	p := NewProgress(opts)

	// Non-TTY should output simple line-based progress
	p.Increment("first item")
	p.Increment("second item")
	p.Increment("third item")
	p.Finish()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Errorf("expected 3 lines of output, got %d: %s", len(lines), output)
	}

	if !strings.Contains(lines[0], "[1/3]") {
		t.Errorf("line 0 should contain [1/3], got: %s", lines[0])
	}

	if !strings.Contains(lines[0], "first item") {
		t.Errorf("line 0 should contain message, got: %s", lines[0])
	}
}
