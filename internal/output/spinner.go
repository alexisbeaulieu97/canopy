// Package output provides helpers for CLI output formatting.
package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Spinner provides an animated loading indicator for long-running operations.
type Spinner struct {
	mu       sync.Mutex
	message  string
	writer   io.Writer
	icons    Icons
	frame    int
	running  bool
	done     chan struct{}
	stopped  chan struct{}
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
		stopped: make(chan struct{}),
		isTTY:   isWriterTTY(writer),
	}
}

// WithWriter sets a custom writer for the spinner. Must be called before Start().
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
	s.finished = false
	s.done = make(chan struct{})
	s.stopped = make(chan struct{})
	s.mu.Unlock()

	if !s.isTTY {
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
	close(s.done)
	stopped := s.stopped
	isTTY := s.isTTY
	writer := s.writer
	s.mu.Unlock()

	if isTTY {
		<-stopped

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
	close(s.done)
	stopped := s.stopped
	isTTY := s.isTTY
	writer := s.writer
	s.mu.Unlock()

	if !isTTY {
		_, _ = fmt.Fprintf(writer, "%s %s\n", icon, message) //nolint:forbidigo // spinner output
		return
	}

	<-stopped

	_, _ = fmt.Fprintf(writer, "\r\033[K%s %s\n", icon, message) //nolint:forbidigo // spinner output
}

// StopWithSuccess stops the spinner with a success message.
func (s *Spinner) StopWithSuccess(message string) {
	s.StopWithMessage(Colorize(SuccessStyle, s.icons.Success()), message)
}

// StopWithError stops the spinner with an error message.
func (s *Spinner) StopWithError(message string) {
	s.StopWithMessage(Colorize(ErrorStyle, s.icons.Error()), message)
}

func (s *Spinner) animate() {
	defer close(s.stopped)

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

			_, _ = fmt.Fprintf(s.writer, "\r\033[K%s %s", Colorize(InfoStyle, frame), msg) //nolint:forbidigo // spinner output
		}
	}
}
