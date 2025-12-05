// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// RepoResolver handles resolution of repository identifiers to domain.Repo objects.
// It supports resolution by alias, URL, or GitHub shorthand (org/repo).
type RepoResolver struct {
	registry *config.RepoRegistry
}

// NewRepoResolver creates a new RepoResolver with the given registry.
func NewRepoResolver(registry *config.RepoRegistry) *RepoResolver {
	return &RepoResolver{
		registry: registry,
	}
}

// Resolve attempts to resolve a raw repository identifier to a domain.Repo.
// The identifier can be:
// - A URL (https://, git://, git@, ssh://, file://)
// - A registry alias
// - A GitHub shorthand (org/repo)
//
// If userRequested is true, unresolved identifiers return an error.
// If userRequested is false, unresolved identifiers return (Repo{}, false, error).
func (r *RepoResolver) Resolve(raw string, userRequested bool) (domain.Repo, bool, error) {
	val := strings.TrimSpace(raw)
	if val == "" {
		return domain.Repo{}, false, nil
	}

	if isLikelyURL(val) {
		if r.registry != nil {
			if entry, ok := r.registry.ResolveByURL(val); ok {
				return domain.Repo{Name: entry.Alias, URL: entry.URL}, true, nil
			}
		}

		return domain.Repo{Name: repoNameFromURL(val), URL: val}, true, nil
	}

	if r.registry != nil {
		if entry, ok := r.registry.Resolve(val); ok {
			return domain.Repo{Name: entry.Alias, URL: entry.URL}, true, nil
		}
	}

	if strings.Count(val, "/") == 1 {
		parts := strings.Split(val, "/")
		url := "https://github.com/" + val

		return domain.Repo{Name: parts[1], URL: url}, true, nil
	}

	if userRequested {
		return domain.Repo{}, false, cerrors.NewUnknownRepository(val, true)
	}

	return domain.Repo{}, false, cerrors.NewUnknownRepository(val, false)
}

// isLikelyURL checks if the given string appears to be a URL.
func isLikelyURL(val string) bool {
	return strings.HasPrefix(val, "http://") ||
		strings.HasPrefix(val, "https://") ||
		strings.HasPrefix(val, "ssh://") ||
		strings.HasPrefix(val, "git://") ||
		strings.HasPrefix(val, "git@") ||
		strings.HasPrefix(val, "file://")
}

// repoNameFromURL extracts the repository name from a URL.
func repoNameFromURL(url string) string {
	// Strip scp-like prefix if present
	if strings.Contains(url, ":") && !strings.HasPrefix(url, "http") {
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
