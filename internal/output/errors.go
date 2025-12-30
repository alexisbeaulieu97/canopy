// Package output provides helpers for CLI output formatting.
package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// KeyValue represents an ordered key-value pair for deterministic output.
type KeyValue struct {
	Key   string
	Value string
}

// ErrorBox renders styled error messages within a box.
type ErrorBox struct {
	title       string
	message     string
	description string
	suggestions []string
	context     []KeyValue // ordered slice for deterministic output
	width       int
}

// NewErrorBox creates a new error box with the given title and message.
func NewErrorBox(title, message string) *ErrorBox {
	return &ErrorBox{
		title:   title,
		message: message,
		width:   70,
	}
}

// WithDescription adds a longer description to the error.
func (e *ErrorBox) WithDescription(desc string) *ErrorBox {
	e.description = desc
	return e
}

// WithSuggestions adds actionable suggestions to the error.
func (e *ErrorBox) WithSuggestions(suggestions ...string) *ErrorBox {
	e.suggestions = append(e.suggestions, suggestions...)
	return e
}

// WithContext adds key-value context to the error in insertion order.
func (e *ErrorBox) WithContext(key, value string) *ErrorBox {
	e.context = append(e.context, KeyValue{Key: key, Value: value})
	return e
}

// WithWidth sets the box width.
func (e *ErrorBox) WithWidth(width int) *ErrorBox {
	e.width = width
	return e
}

// Render outputs the error box to stderr.
func (e *ErrorBox) Render() {
	icons := NewIcons()
	box := NewBox(e.title).
		WithWriter(os.Stderr).
		WithWidth(e.width).
		WithStyle(BoxStyle{
			Title:      ErrorStyle,
			Border:     ErrorStyle,
			Content:    lipgloss.NewStyle(),
			HeaderLine: ErrorStyle,
		})

	var lines []string

	// Main error message with icon
	errorLine := Colorize(ErrorStyle, icons.Error()+" "+e.message)
	lines = append(lines, errorLine)

	// Description if present
	if e.description != "" {
		lines = append(lines, "")
		lines = append(lines, wrapText(e.description, e.width-6)...)
	}

	// Context key-value pairs (in insertion order)
	if len(e.context) > 0 {
		lines = append(lines, "")

		for _, kv := range e.context {
			contextLine := Colorize(MutedStyle, kv.Key+":") + " " + kv.Value
			lines = append(lines, contextLine)
		}
	}

	// Suggestions
	if len(e.suggestions) > 0 {
		lines = append(lines, "")

		for _, suggestion := range e.suggestions {
			bullet := icons.Bullet()
			suggestionLine := "  " + bullet + " " + suggestion
			lines = append(lines, suggestionLine)
		}
	}

	box.Render(lines)
}

// RenderString returns the error box as a string.
func (e *ErrorBox) RenderString() string {
	// For now, we render to stderr in Render().
	// This could be extended to return a string.
	return ""
}

// SuccessMessage renders a styled success message.
type SuccessMessage struct {
	title   string
	details []KeyValue // ordered slice for deterministic output
	hints   []string
}

// NewSuccessMessage creates a new success message.
func NewSuccessMessage(title string) *SuccessMessage {
	return &SuccessMessage{
		title: title,
	}
}

// WithDetail adds a detail key-value pair in insertion order.
func (s *SuccessMessage) WithDetail(key, value string) *SuccessMessage {
	s.details = append(s.details, KeyValue{Key: key, Value: value})
	return s
}

// WithHint adds a hint/suggestion.
func (s *SuccessMessage) WithHint(hint string) *SuccessMessage {
	s.hints = append(s.hints, hint)
	return s
}

// Render outputs the success message.
func (s *SuccessMessage) Render() {
	icons := NewIcons()

	// Main success line
	successLine := Colorize(SuccessStyle, icons.Success()) + " " + s.title
	fmt.Println(successLine) //nolint:forbidigo // user-facing CLI output

	// Details (in insertion order)
	if len(s.details) > 0 {
		fmt.Println() //nolint:forbidigo // user-facing CLI output

		for _, kv := range s.details {
			fmt.Printf("  %s  %s\n", Colorize(MutedStyle, kv.Key+":"), kv.Value) //nolint:forbidigo // user-facing CLI output
		}
	}

	// Hints
	if len(s.hints) > 0 {
		fmt.Println() //nolint:forbidigo // user-facing CLI output

		for _, hint := range s.hints {
			fmt.Println(Colorize(MutedStyle, hint)) //nolint:forbidigo // user-facing CLI output
		}
	}
}

// WarningBox renders a warning message within a styled box.
type WarningBox struct {
	title  string
	items  []string
	action string
	width  int
}

// NewWarningBox creates a new warning box.
func NewWarningBox(title string) *WarningBox {
	return &WarningBox{
		title: title,
		width: 70,
	}
}

// WithItems adds list items to the warning.
func (w *WarningBox) WithItems(items ...string) *WarningBox {
	w.items = append(w.items, items...)
	return w
}

// WithAction adds a suggested action.
func (w *WarningBox) WithAction(action string) *WarningBox {
	w.action = action
	return w
}

// WithWidth sets the box width.
func (w *WarningBox) WithWidth(width int) *WarningBox {
	w.width = width
	return w
}

// Render outputs the warning box.
func (w *WarningBox) Render() {
	icons := NewIcons()
	box := NewBox(icons.Warning() + " " + w.title).
		WithWidth(w.width).
		WithStyle(BoxStyle{
			Title:      WarningStyle,
			Border:     WarningStyle,
			Content:    lipgloss.NewStyle(),
			HeaderLine: WarningStyle,
		})

	var lines []string

	for _, item := range w.items {
		lines = append(lines, "  "+icons.Bullet()+" "+item)
	}

	if w.action != "" {
		lines = append(lines, "")
		lines = append(lines, Colorize(MutedStyle, w.action))
	}

	box.Render(lines)
}

// wrapText wraps text to fit within the given width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var (
		lines       []string
		currentLine strings.Builder
	)

	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
			continue
		}

		if currentLine.Len()+1+len(word) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
