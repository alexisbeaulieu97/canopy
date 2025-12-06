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

	// Try URL resolution
	if repo, ok := r.resolveURL(val); ok {
		return repo, true, nil
	}

	// Try registry alias
	if repo, ok := r.resolveRegistry(val); ok {
		return repo, true, nil
	}

	// Try GitHub shorthand
	if repo, ok := r.resolveGitHubShorthand(val); ok {
		return repo, true, nil
	}

	return domain.Repo{}, false, cerrors.NewUnknownRepository(val, userRequested)
}

// resolveURL attempts to resolve a URL to a repo.
func (r *RepoResolver) resolveURL(val string) (domain.Repo, bool) {
	if !isLikelyURL(val) {
		return domain.Repo{}, false
	}

	if r.registry != nil {
		if entry, ok := r.registry.ResolveByURL(val); ok {
			return domain.Repo{Name: entry.Alias, URL: entry.URL}, true
		}
	}

	return domain.Repo{Name: repoNameFromURL(val), URL: val}, true
}

// resolveRegistry attempts to resolve via registry alias.
func (r *RepoResolver) resolveRegistry(val string) (domain.Repo, bool) {
	if r.registry == nil {
		return domain.Repo{}, false
	}

	if entry, ok := r.registry.Resolve(val); ok {
		return domain.Repo{Name: entry.Alias, URL: entry.URL}, true
	}

	return domain.Repo{}, false
}

// resolveGitHubShorthand attempts to resolve GitHub shorthand (owner/repo).
func (r *RepoResolver) resolveGitHubShorthand(val string) (domain.Repo, bool) {
	if strings.Count(val, "/") != 1 {
		return domain.Repo{}, false
	}

	parts := strings.Split(val, "/")
	owner := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])

	// Both owner and repo must be non-empty for valid GitHub shorthand
	if owner == "" || repo == "" {
		return domain.Repo{}, false
	}

	url := "https://github.com/" + owner + "/" + repo

	return domain.Repo{Name: repo, URL: url}, true
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
