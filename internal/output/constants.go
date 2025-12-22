// Package output provides helpers for CLI output formatting.
package output

import "strings"

// Shared formatting constants for CLI output.
const (
	SeparatorWidth      = 50
	SeparatorChar       = "─"
	RepoNameWidth       = 20
	RepoSizeWidth       = 10
	RepoLastFetchWidth  = 20
	RepoWorkspacesWidth = 10
	RepoAliasWidth      = 16
	RepoURLWidth        = 45
	RepoTagsWidth       = 20
)

// SeparatorLine returns a horizontal separator line at the given width.
func SeparatorLine(width int) string {
	if width <= 0 {
		return ""
	}

	return strings.Repeat(SeparatorChar, width)
}
