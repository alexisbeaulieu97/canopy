// Package logging provides simple structured logging helpers.
package logging

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

// Logger wraps the application logger
type Logger struct {
	*log.Logger
}

// New creates a new logger instance
func New(debug bool) *Logger {
	l := log.New(os.Stderr)
	l.SetReportTimestamp(true)
	l.SetTimeFormat(time.Kitchen)

	if debug {
		l.SetLevel(log.DebugLevel)
	} else {
		l.SetLevel(log.InfoLevel)
	}

	return &Logger{Logger: l}
}

// SetDebug enables debug logging
func (l *Logger) SetDebug(debug bool) {
	if debug {
		l.SetLevel(log.DebugLevel)
	} else {
		l.SetLevel(log.InfoLevel)
	}
}
