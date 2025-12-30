// Package output provides helpers for CLI output formatting.
package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
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

	var bar progress.Model
	if ColorEnabled() {
		bar = progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(opts.Width),
			progress.WithoutPercentage(),
		)
	} else {
		// Use solid fill without colors when color is disabled
		bar = progress.New(
			progress.WithSolidFill("#"),
			progress.WithWidth(opts.Width),
			progress.WithoutPercentage(),
		)
	}

	isTTY := isWriterTTY(opts.Writer)

	return &Progress{
		opts:         opts,
		bar:          bar,
		isTTY:        isTTY,
		lastRendered: -1, // sentinel so first update at 0 is rendered
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
		line += pctStr
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

// Spinner provides an animated loading indicator for long-running operations.
type Spinner struct {
	mu       sync.Mutex
	message  string
	writer   io.Writer
	icons    Icons
	frame    int
	running  bool
	done     chan struct{}
	finished bool
	isTTY    bool
}

// NewSpinner creates a new spinner with the given initial message.
func NewSpinner(message string) *Spinner {
	writer := os.Stderr

	return &Spinner{
		message: message,
		writer:  writer,
		icons:   NewIcons(),
		done:    make(chan struct{}),
		isTTY:   isWriterTTY(writer),
	}
}

// WithWriter sets a custom writer for the spinner.
// Must be called before Start().
func (s *Spinner) WithWriter(w io.Writer) *Spinner {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.writer = w
	s.isTTY = isWriterTTY(w)

	return s
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	s.mu.Lock()

	if s.running {
		s.mu.Unlock()
		return
	}

	s.running = true
	s.mu.Unlock()

	if !s.isTTY {
		// For non-TTY, just print the initial message
		_, _ = fmt.Fprintf(s.writer, "%s\n", s.message) //nolint:forbidigo // spinner output
		return
	}

	go s.animate()
}

// SetMessage updates the spinner message.
func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.message = msg
}

// Stop stops the spinner and clears the line.
func (s *Spinner) Stop() {
	s.mu.Lock()

	if !s.running || s.finished {
		s.mu.Unlock()
		return
	}

	s.finished = true
	s.running = false
	close(s.done) // close while holding lock to prevent double-close
	isTTY := s.isTTY
	writer := s.writer
	s.mu.Unlock()

	if isTTY {
		// Clear the spinner line
		_, _ = fmt.Fprint(writer, "\r\033[K") //nolint:forbidigo // spinner output
	}
}

// StopWithMessage stops the spinner and displays a final message.
func (s *Spinner) StopWithMessage(icon, message string) {
	s.mu.Lock()

	if !s.running || s.finished {
		s.mu.Unlock()
		return
	}

	s.finished = true
	s.running = false
	close(s.done) // close while holding lock to prevent double-close
	isTTY := s.isTTY
	writer := s.writer
	s.mu.Unlock()

	if isTTY {
		_, _ = fmt.Fprintf(writer, "\r\033[K%s %s\n", icon, message) //nolint:forbidigo // spinner output
	} else {
		_, _ = fmt.Fprintf(writer, "%s %s\n", icon, message) //nolint:forbidigo // spinner output
	}
}

// StopWithSuccess stops the spinner with a success message.
func (s *Spinner) StopWithSuccess(message string) {
	icon := Colorize(SuccessStyle, s.icons.Success())
	s.StopWithMessage(icon, message)
}

// StopWithError stops the spinner with an error message.
func (s *Spinner) StopWithError(message string) {
	icon := Colorize(ErrorStyle, s.icons.Error())
	s.StopWithMessage(icon, message)
}

func (s *Spinner) animate() {
	frames := s.icons.SpinnerFrames()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			frame := frames[s.frame%len(frames)]
			msg := s.message
			s.frame++
			s.mu.Unlock()

			spinnerIcon := Colorize(InfoStyle, frame)
			_, _ = fmt.Fprintf(s.writer, "\r\033[K%s %s", spinnerIcon, msg) //nolint:forbidigo // spinner output
		}
	}
}

// MultiProgress tracks progress for multiple concurrent operations.
type MultiProgress struct {
	mu      sync.Mutex
	results []ProgressResult
	writer  io.Writer
	total   int
	icons   Icons
	isTTY   bool
}

// ProgressResult represents the result of a single operation.
type ProgressResult struct {
	Name    string
	Success bool
	Message string
}

// NewMultiProgress creates a new multi-operation progress tracker.
func NewMultiProgress(total int) *MultiProgress {
	writer := os.Stderr

	return &MultiProgress{
		results: make([]ProgressResult, 0, total),
		writer:  writer,
		total:   total,
		icons:   NewIcons(),
		isTTY:   isWriterTTY(writer),
	}
}

// WithWriter sets a custom writer.
// Must be called before any Complete() or Fail() calls.
func (m *MultiProgress) WithWriter(w io.Writer) *MultiProgress {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.writer = w
	m.isTTY = isWriterTTY(w)

	return m
}

// Complete marks an operation as complete with success.
func (m *MultiProgress) Complete(name, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := ProgressResult{
		Name:    name,
		Success: true,
		Message: message,
	}
	m.results = append(m.results, result)
	m.render(result)
}

// Fail marks an operation as failed.
func (m *MultiProgress) Fail(name, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := ProgressResult{
		Name:    name,
		Success: false,
		Message: message,
	}
	m.results = append(m.results, result)
	m.render(result)
}

func (m *MultiProgress) render(result ProgressResult) {
	icon := m.icons.Success()
	style := SuccessStyle

	if !result.Success {
		icon = m.icons.Error()
		style = ErrorStyle
	}

	coloredIcon := Colorize(style, icon)
	_, _ = fmt.Fprintf(m.writer, "%s %-20s %s\n", coloredIcon, result.Name, result.Message) //nolint:forbidigo // progress output
}

// Summary returns counts of successful and failed operations.
func (m *MultiProgress) Summary() (success, failed int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range m.results {
		if r.Success {
			success++
		} else {
			failed++
		}
	}

	return success, failed
}
