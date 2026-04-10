// Package output provides helpers for CLI output formatting.
package output

import (
	"fmt"
	"io"
	"os"
	"sync"
)

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
func (m *MultiProgress) WithWriter(w io.Writer) *MultiProgress {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.writer = w
	m.isTTY = isWriterTTY(w)

	return m
}

// Complete marks an operation as complete with success.
func (m *MultiProgress) Complete(name, message string) {
	m.appendAndRender(ProgressResult{Name: name, Success: true, Message: message})
}

// Fail marks an operation as failed.
func (m *MultiProgress) Fail(name, message string) {
	m.appendAndRender(ProgressResult{Name: name, Success: false, Message: message})
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

func (m *MultiProgress) appendAndRender(result ProgressResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	_, _ = fmt.Fprintf(m.writer, "%s %-20s %s\n", Colorize(style, icon), result.Name, result.Message) //nolint:forbidigo // progress output
}
