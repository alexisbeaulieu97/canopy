// Package giturl provides utilities for parsing and manipulating Git URLs.
package giturl

import (
	"strings"
)

// IsURL checks if the given string appears to be a Git URL.
// It recognizes http, https, ssh, git, file, and scp-style (git@) URLs.
func IsURL(val string) bool {
	return strings.HasPrefix(val, "http://") ||
		strings.HasPrefix(val, "https://") ||
		strings.HasPrefix(val, "ssh://") ||
		strings.HasPrefix(val, "git://") ||
		strings.HasPrefix(val, "git@") ||
		strings.HasPrefix(val, "file://")
}

// ExtractRepoName extracts the repository name from a Git URL.
// It handles various URL formats including scp-style URLs and strips .git suffixes.
func ExtractRepoName(url string) string {
	// Strip scp-like prefix if present (e.g., git@github.com:org/repo)
	// Only match true scp-style URLs (user@host:path) by checking for ":" without "://"
	if strings.Contains(url, ":") && !strings.Contains(url, "://") {
		parts := strings.Split(url, ":")
		url = parts[len(parts)-1]
	}

	parts := strings.Split(url, "/")

	var name string

	for i := len(parts) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(parts[i]); trimmed != "" {
			name = trimmed
			break
		}
	}

	if name == "" {
		return ""
	}

	return strings.TrimSuffix(name, ".git")
}

// DeriveAlias returns a sensible alias from a Git URL.
// It extracts the repository name and converts it to lowercase.
func DeriveAlias(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}

	name := ExtractRepoName(url)
	if name == "" {
		return ""
	}

	return strings.ToLower(name)
}
