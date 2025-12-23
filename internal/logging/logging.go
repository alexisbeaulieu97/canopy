// Package logging provides simple structured logging helpers.
package logging

import (
	"os"
	"regexp"
	"time"

	"github.com/charmbracelet/log"
)

// sensitivePatterns matches common sensitive data patterns for redaction.
var sensitivePatterns = []*regexp.Regexp{
	// API keys, tokens, secrets (key=value or key:value) - captures up to next whitespace or quote
	regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?secret|auth[_-]?token|access[_-]?token|secret[_-]?key|password|passwd|pwd)\s*[=:]\s*[^\s]+`),
	// Bearer tokens (Authorization: Bearer xxx)
	regexp.MustCompile(`(?i)bearer\s+[^\s]+`),
	// SSH URLs with embedded credentials
	regexp.MustCompile(`ssh://[^@\s]+@`),
	// HTTPS URLs with embedded credentials
	regexp.MustCompile(`https?://[^:@\s]+:[^@\s]+@`),
	// AWS-style keys (AKIA...)
	regexp.MustCompile(`(?i)(AKIA|ASIA)[A-Z0-9]{16}`),
	// Generic hex/base64 tokens that look like secrets (32+ chars)
	regexp.MustCompile(`(?i)(token|key|secret|password)[=:]["']?[A-Za-z0-9+/]{32,}=*["']?`),
}

// RedactSensitive replaces potentially sensitive data in a string with [REDACTED].
// This is used to sanitize log output that might contain secrets.
func RedactSensitive(input string) string {
	result := input
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllString(result, "[REDACTED]")
	}

	return result
}

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
