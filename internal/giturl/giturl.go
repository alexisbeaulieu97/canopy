// Package giturl provides utilities for parsing and manipulating Git URLs.
package giturl

import (
	"net/url"
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
// Returns an empty string if no repository name can be extracted.
func ExtractRepoName(rawURL string) string {
	// Handle scp-style URLs (user@host:path) - they don't have "://"
	if strings.Contains(rawURL, ":") && !strings.Contains(rawURL, "://") {
		parts := strings.Split(rawURL, ":")
		rawURL = parts[len(parts)-1]

		return extractNameFromPath(rawURL)
	}

	// For scheme-based URLs, use net/url to properly parse
	if strings.Contains(rawURL, "://") {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return extractNameFromPath(rawURL)
		}

		// If there's no meaningful path, return empty string
		// (e.g., "git://host:9418" has no repo)
		if parsed.Path == "" || parsed.Path == "/" {
			return ""
		}

		return extractNameFromPath(parsed.Path)
	}

	return extractNameFromPath(rawURL)
}

// extractNameFromPath extracts the repository name from a URL path.
func extractNameFromPath(path string) string {
	parts := strings.Split(path, "/")

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
