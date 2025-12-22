// Package output provides helpers for CLI output formatting.
package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// Formatter provides consistent success/error formatting.
type Formatter interface {
	PrintSuccess(data interface{}) error
	PrintError(err error) error
}

// Response is the standard JSON envelope for all CLI output.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents structured error information in JSON output.
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Context map[string]string `json:"context,omitempty"`
	Details string            `json:"details,omitempty"`
}

// JSONPrinter handles JSON output formatting.
type JSONPrinter struct {
	writer io.Writer
	indent string
}

// NewJSONPrinter creates a new JSONPrinter with default settings.
func NewJSONPrinter() *JSONPrinter {
	return &JSONPrinter{
		writer: os.Stdout,
		indent: "  ",
	}
}

// WithWriter sets a custom writer for the printer.
func (p *JSONPrinter) WithWriter(w io.Writer) *JSONPrinter {
	p.writer = w
	return p
}

// PrintSuccess prints a successful response with the given data.
func (p *JSONPrinter) PrintSuccess(data interface{}) error {
	response := Response{
		Success: true,
		Data:    data,
	}

	return p.encode(response)
}

// PrintError prints an error response.
func (p *JSONPrinter) PrintError(err error) error {
	errInfo := errorToInfo(err)
	response := Response{
		Success: false,
		Error:   errInfo,
	}

	return p.encode(response)
}

func (p *JSONPrinter) encode(v interface{}) error {
	encoder := json.NewEncoder(p.writer)
	encoder.SetIndent("", p.indent)

	return encoder.Encode(v)
}

// errorToInfo converts an error to ErrorInfo.
func errorToInfo(err error) *ErrorInfo {
	if err == nil {
		return nil
	}

	var canopyErr *cerrors.CanopyError
	if ok := errors.As(err, &canopyErr); ok {
		var details string
		if canopyErr.Cause != nil {
			details = canopyErr.Cause.Error()
		}

		return &ErrorInfo{
			Code:    string(canopyErr.Code),
			Message: canopyErr.Message,
			Context: canopyErr.Context,
			Details: details,
		}
	}

	// Fallback for non-CanopyError errors
	return &ErrorInfo{
		Code:    string(cerrors.ErrInternalError),
		Message: err.Error(),
	}
}

// PrintJSON writes the value as indented JSON to stdout.
func PrintJSON(v interface{}) error {
	return NewJSONPrinter().PrintSuccess(v)
}

// PrintErrorJSON writes an error as structured JSON to stdout.
func PrintErrorJSON(err error) error {
	return NewJSONPrinter().PrintError(err)
}

// FormatErrorJSON returns a standardized JSON error string.
func FormatErrorJSON(err error) string {
	var buf bytes.Buffer

	printer := NewJSONPrinter().WithWriter(&buf)
	if printErr := printer.PrintError(err); printErr != nil {
		return fmt.Sprintf(`{"success":false,"error":{"code":"%s","message":%q}}`, cerrors.ErrInternalError, err.Error())
	}

	return strings.TrimSpace(buf.String())
}

// FormatBytes formats a byte count as a human-readable string (B, KB, MB, GB).
func FormatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
