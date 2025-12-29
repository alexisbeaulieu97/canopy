// Package output provides helpers for CLI output formatting.
package output

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// ProgressOptions configures progress bar behavior.
type ProgressOptions struct {
	// Total is the number of items to process.
	Total int
	// Width is the progress bar width in characters.
	Width int
	// ShowPercentage displays completion percentage.
	ShowPercentage bool
	// ShowCount displays "current/total" count.
	ShowCount bool
	// Writer is the output destination (defaults to os.Stderr).
	Writer io.Writer
}

// DefaultProgressOptions returns sensible defaults for progress display.
func DefaultProgressOptions(total int) ProgressOptions {
	return ProgressOptions{
		Total:          total,
		Width:          40,
		ShowPercentage: true,
		ShowCount:      true,
		Writer:         os.Stderr,
	}
}

// Progress tracks and displays operation progress.
type Progress struct {
	mu           sync.Mutex
	opts         ProgressOptions
	current      int
	lastRendered int // track last rendered progress for non-TTY deduplication
	message      string
	bar          progress.Model
	finished     bool
	isTTY        bool
}

// NewProgress creates a progress tracker with the given options.
func NewProgress(opts ProgressOptions) *Progress {
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	if opts.Width <= 0 {
		opts.Width = 40
	}

	bar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(opts.Width),
		progress.WithoutPercentage(),
	)

	isTTY := isWriterTTY(opts.Writer)

	return &Progress{
		opts:  opts,
		bar:   bar,
		isTTY: isTTY,
	}
}

// SetMessage updates the current operation message.
func (p *Progress) SetMessage(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.message = msg
	p.render()
}

// Increment advances progress by one and optionally updates the message.
func (p *Progress) Increment(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current++
	if msg != "" {
		p.message = msg
	}

	p.render()
}

// SetCurrent sets the current progress value directly.
func (p *Progress) SetCurrent(n int, msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = n
	if msg != "" {
		p.message = msg
	}

	p.render()
}

// Current returns the current progress count.
func (p *Progress) Current() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.current
}

// Finish completes the progress bar and moves to a new line.
func (p *Progress) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}

	p.finished = true

	if p.isTTY {
		// Clear the line and move to start
		_, _ = fmt.Fprint(p.opts.Writer, "\r\033[K") //nolint:forbidigo // progress bar output
	}
}

// Cancel marks the progress as cancelled.
func (p *Progress) Cancel() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}

	p.finished = true

	if p.isTTY {
		// Show cancelled state
		_, _ = fmt.Fprint(p.opts.Writer, "\r\033[K")                            //nolint:forbidigo // progress bar output
		_, _ = fmt.Fprintln(p.opts.Writer, Colorize(WarningStyle, "Cancelled")) //nolint:forbidigo // progress bar output
	} else {
		_, _ = fmt.Fprintln(p.opts.Writer, "Cancelled") //nolint:forbidigo // progress bar output
	}
}

func (p *Progress) render() {
	if p.finished {
		return
	}

	var pct float64
	if p.opts.Total > 0 {
		pct = float64(p.current) / float64(p.opts.Total)
	}

	if p.isTTY {
		p.renderTTY(pct)
	} else {
		p.renderNonTTY()
	}
}

func (p *Progress) renderTTY(pct float64) {
	// Move to start of line and clear it
	_, _ = fmt.Fprint(p.opts.Writer, "\r\033[K") //nolint:forbidigo // progress bar output

	// Build the progress line
	var line string

	// Progress bar
	line += p.bar.ViewAs(pct)

	// Count
	if p.opts.ShowCount && p.opts.Total > 0 {
		countStr := fmt.Sprintf(" %d/%d", p.current, p.opts.Total)
		line += Colorize(MutedStyle, countStr)
	}

	// Percentage
	if p.opts.ShowPercentage {
		pctStr := fmt.Sprintf(" %3.0f%%", pct*100)
		line += Colorize(lipgloss.NewStyle(), pctStr)
	}

	// Message
	if p.message != "" {
		line += " " + p.message
	}

	_, _ = fmt.Fprint(p.opts.Writer, line) //nolint:forbidigo // progress bar output
}

func (p *Progress) renderNonTTY() {
	// For non-TTY, only print when progress actually changes (avoid duplicates from SetMessage)
	if p.current == p.lastRendered {
		return
	}

	p.lastRendered = p.current

	if p.opts.Total > 0 {
		_, _ = fmt.Fprintf(p.opts.Writer, "[%d/%d] %s\n", p.current, p.opts.Total, p.message) //nolint:forbidigo // progress bar output
	} else {
		_, _ = fmt.Fprintf(p.opts.Writer, "%s\n", p.message) //nolint:forbidigo // progress bar output
	}
}

// IsTTY returns whether progress output goes to a terminal.
func (p *Progress) IsTTY() bool {
	return p.isTTY
}

func isWriterTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}

	return false
}
