// Package output provides helpers for CLI output formatting.
package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BoxChars defines the characters used for box drawing.
type BoxChars struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	TeeLeft     string
	TeeRight    string
}

// UnicodeBox provides Unicode box-drawing characters.
var UnicodeBox = BoxChars{
	TopLeft:     "┌",
	TopRight:    "┐",
	BottomLeft:  "└",
	BottomRight: "┘",
	Horizontal:  "─",
	Vertical:    "│",
	TeeLeft:     "├",
	TeeRight:    "┤",
}

// ASCIIBox provides ASCII fallback characters.
var ASCIIBox = BoxChars{
	TopLeft:     "+",
	TopRight:    "+",
	BottomLeft:  "+",
	BottomRight: "+",
	Horizontal:  "-",
	Vertical:    "|",
	TeeLeft:     "+",
	TeeRight:    "+",
}

// BoxStyle defines visual styling for a box.
type BoxStyle struct {
	Title      lipgloss.Style
	Content    lipgloss.Style
	Border     lipgloss.Style
	HeaderLine lipgloss.Style
}

// DefaultBoxStyle returns the default styling for boxes.
func DefaultBoxStyle() BoxStyle {
	return BoxStyle{
		Title:      AccentStyle,
		Content:    lipgloss.NewStyle(),
		Border:     MutedStyle,
		HeaderLine: MutedStyle,
	}
}

// Box renders content within a bordered box.
type Box struct {
	title  string
	width  int
	chars  BoxChars
	style  BoxStyle
	writer io.Writer
}

// NewBox creates a new Box with the given title.
func NewBox(title string) *Box {
	chars := UnicodeBox
	if !UnicodeEnabled() {
		chars = ASCIIBox
	}

	return &Box{
		title:  title,
		width:  70,
		chars:  chars,
		style:  DefaultBoxStyle(),
		writer: os.Stdout,
	}
}

// WithWidth sets the box width.
func (b *Box) WithWidth(width int) *Box {
	b.width = width
	return b
}

// WithStyle sets the box style.
func (b *Box) WithStyle(style BoxStyle) *Box {
	b.style = style
	return b
}

// WithWriter sets the output writer.
func (b *Box) WithWriter(w io.Writer) *Box {
	b.writer = w
	return b
}

// renderLine outputs a line with proper box borders.
func (b *Box) renderLine(content string) {
	border := Colorize(b.style.Border, b.chars.Vertical)

	padding := b.width - 4 - runeWidth(content)
	if padding < 0 {
		padding = 0
	}

	_, _ = fmt.Fprintf(b.writer, "%s  %s%s  %s\n",
		border, content, strings.Repeat(" ", padding), border)
}

// renderEmptyLine outputs an empty line within the box.
func (b *Box) renderEmptyLine() {
	border := Colorize(b.style.Border, b.chars.Vertical)
	_, _ = fmt.Fprintf(b.writer, "%s%s%s\n",
		border, strings.Repeat(" ", b.width-2), border)
}

// renderTopBorder outputs the top border with title.
func (b *Box) renderTopBorder() {
	titlePart := b.chars.Horizontal + " " + Colorize(b.style.Title, b.title) + " "
	titleLen := 3 + len(b.title) // "─ " + title + " "

	remaining := b.width - 2 - titleLen
	if remaining < 0 {
		remaining = 0
	}

	border := Colorize(b.style.Border,
		b.chars.TopLeft+titlePart+strings.Repeat(b.chars.Horizontal, remaining)+b.chars.TopRight)
	_, _ = fmt.Fprintln(b.writer, border)
}

// renderBottomBorder outputs the bottom border.
func (b *Box) renderBottomBorder() {
	border := Colorize(b.style.Border,
		b.chars.BottomLeft+strings.Repeat(b.chars.Horizontal, b.width-2)+b.chars.BottomRight)
	_, _ = fmt.Fprintln(b.writer, border)
}

// renderSectionDivider outputs a section divider with optional title.
func (b *Box) renderSectionDivider(title string) {
	if title == "" {
		line := Colorize(b.style.Border,
			b.chars.TeeLeft+strings.Repeat(b.chars.Horizontal, b.width-2)+b.chars.TeeRight)
		_, _ = fmt.Fprintln(b.writer, line)

		return
	}

	titlePart := b.chars.Horizontal + " " + Colorize(b.style.Title, title) + " "
	titleLen := 3 + len(title)

	remaining := b.width - 2 - titleLen
	if remaining < 0 {
		remaining = 0
	}

	line := Colorize(b.style.Border,
		b.chars.TeeLeft+titlePart+strings.Repeat(b.chars.Horizontal, remaining)+b.chars.TeeRight)
	_, _ = fmt.Fprintln(b.writer, line)
}

// Render renders complete box content.
func (b *Box) Render(lines []string) {
	b.renderTopBorder()
	b.renderEmptyLine()

	for _, line := range lines {
		b.renderLine(line)
	}

	b.renderEmptyLine()
	b.renderBottomBorder()
}

// RenderWithSections renders box content with multiple sections.
func (b *Box) RenderWithSections(sections []BoxSection) {
	if len(sections) == 0 {
		return
	}

	b.renderTopBorder()
	b.renderEmptyLine()

	for i, section := range sections {
		if i > 0 {
			b.renderEmptyLine()
			b.renderSectionDivider(section.Title)
			b.renderEmptyLine()
		}

		for _, line := range section.Lines {
			b.renderLine(line)
		}
	}

	b.renderEmptyLine()
	b.renderBottomBorder()
}

// BoxSection represents a titled section within a box.
type BoxSection struct {
	Title string
	Lines []string
}

// Table represents a formatted table with columns.
type Table struct {
	columns []TableColumn
	rows    [][]string
	icons   Icons
}

// TableColumn defines a column in a table.
type TableColumn struct {
	Header string
	Width  int
	Style  lipgloss.Style
	Align  Alignment
}

// Alignment specifies text alignment within a column.
type Alignment int

const (
	// AlignLeft aligns text to the left.
	AlignLeft Alignment = iota
	// AlignRight aligns text to the right.
	AlignRight
	// AlignCenter centers the text.
	AlignCenter
)

// NewTable creates a new table with the given columns.
func NewTable(columns ...TableColumn) *Table {
	return &Table{
		columns: columns,
		rows:    make([][]string, 0),
		icons:   NewIcons(),
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}

// Render outputs the table.
func (t *Table) Render(w io.Writer) {
	if len(t.columns) == 0 {
		return
	}

	// Render header
	header := t.renderRow(t.columnHeaders())
	_, _ = fmt.Fprintln(w, header)

	// Render separator
	sep := t.renderSeparator()
	_, _ = fmt.Fprintln(w, sep)

	// Render rows
	for _, row := range t.rows {
		line := t.renderRow(row)
		_, _ = fmt.Fprintln(w, line)
	}
}

func (t *Table) columnHeaders() []string {
	headers := make([]string, len(t.columns))
	for i, col := range t.columns {
		headers[i] = col.Header
	}

	return headers
}

func (t *Table) renderRow(values []string) string {
	var parts []string

	for i, col := range t.columns {
		val := ""
		if i < len(values) {
			val = values[i]
		}

		formatted := t.formatCell(val, col)
		parts = append(parts, formatted)
	}

	return strings.Join(parts, "  ")
}

func (t *Table) formatCell(value string, col TableColumn) string {
	// Truncate if necessary
	runes := []rune(value)
	if len(runes) > col.Width && col.Width > 0 {
		value = string(runes[:col.Width-1]) + "…"
	}

	if !ColorEnabled() {
		return padString(value, col.Width, col.Align)
	}

	return col.Style.Inline(true).Width(col.Width).Render(value)
}

func (t *Table) renderSeparator() string {
	chars := UnicodeBox
	if !t.icons.UseUnicode() {
		chars = ASCIIBox
	}

	var parts []string
	for _, col := range t.columns {
		parts = append(parts, strings.Repeat(chars.Horizontal, col.Width))
	}

	return Colorize(MutedStyle, strings.Join(parts, "  "))
}

// padString pads a string to the given width with specified alignment.
func padString(s string, width int, align Alignment) string {
	if width <= 0 {
		return s
	}

	runes := []rune(s)
	if len(runes) >= width {
		return s
	}

	padding := width - len(runes)

	switch align {
	case AlignRight:
		return strings.Repeat(" ", padding) + s
	case AlignCenter:
		left := padding / 2
		right := padding - left

		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	default: // AlignLeft
		return s + strings.Repeat(" ", padding)
	}
}

// runeWidth returns the display width of a string (treating each rune as width 1).
func runeWidth(s string) int {
	return len([]rune(s))
}

// KeyValue formats a key-value pair for display.
func KeyValue(key, value string, keyWidth int) string {
	paddedKey := padString(key, keyWidth, AlignLeft)
	return Colorize(MutedStyle, paddedKey) + "  " + value
}

// Summary outputs a summary line with bullet separators.
func Summary(parts ...string) string {
	icons := NewIcons()
	sep := " " + icons.Bullet() + " "

	return Colorize(MutedStyle, strings.Join(parts, sep))
}

// HorizontalRule outputs a horizontal divider line.
func HorizontalRule(width int) string {
	chars := UnicodeBox
	if !UnicodeEnabled() {
		chars = ASCIIBox
	}

	return Colorize(MutedStyle, strings.Repeat(chars.Horizontal, width))
}
