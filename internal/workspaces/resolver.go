// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/giturl"
)

// RepoResolver handles resolution of repository identifiers to domain.Repo objects.
// It iterates through a chain of resolution strategies until one succeeds.
type RepoResolver struct {
	strategies []ResolutionStrategy
}

// NewRepoResolver creates a new RepoResolver with the default strategy chain.
// The default order is: URL → Registry → GitHub Shorthand.
func NewRepoResolver(registry *config.RepoRegistry) *RepoResolver {
	return NewRepoResolverWithStrategies(DefaultStrategies(registry))
}

// NewRepoResolverWithStrategies creates a new RepoResolver with custom strategies.
// Strategies are tried in the order provided.
func NewRepoResolverWithStrategies(strategies []ResolutionStrategy) *RepoResolver {
	return &RepoResolver{
		strategies: strategies,
	}
}

// DefaultStrategies returns the default resolution strategy chain.
// Order: URL → Registry → GitHub Shorthand.
func DefaultStrategies(registry *config.RepoRegistry) []ResolutionStrategy {
	var urlLookup URLRegistryLookup
	var registryLookup RegistryLookup

	if registry != nil {
		urlLookup = func(url string) (string, string, bool) {
			if entry, ok := registry.ResolveByURL(url); ok {
				return entry.Alias, entry.URL, true
			}
			return "", "", false
		}

		registryLookup = func(alias string) (string, string, bool) {
			if entry, ok := registry.Resolve(alias); ok {
				return entry.Alias, entry.URL, true
			}
			return "", "", false
		}
	}

	return []ResolutionStrategy{
		NewURLStrategy(urlLookup),
		NewRegistryStrategy(registryLookup),
		NewGitHubShorthandStrategy(),
	}
}

// Resolve attempts to resolve a raw repository identifier to a domain.Repo.
// The identifier can be:
// - A URL (https://, git://, git@, ssh://, file://)
// - A registry alias
// - A GitHub shorthand (org/repo)
//
// Resolution strategies are tried in order until one succeeds.
// If no strategy can resolve the identifier, an error is returned.
func (r *RepoResolver) Resolve(raw string, userRequested bool) (domain.Repo, bool, error) {
	val := strings.TrimSpace(raw)
	if val == "" {
		return domain.Repo{}, false, nil
	}

	for _, strategy := range r.strategies {
		if repo, ok := strategy.Resolve(val); ok {
			return repo, true, nil
		}
	}

	return domain.Repo{}, false, cerrors.NewUnknownRepository(val, userRequested)
}

// isLikelyURL checks if the given string appears to be a URL.
// Deprecated: Use giturl.IsURL instead.
func isLikelyURL(val string) bool {
	return giturl.IsURL(val)
}

// repoNameFromURL extracts the repository name from a URL.
// Deprecated: Use giturl.ExtractRepoName instead.
func repoNameFromURL(url string) string {
	return giturl.ExtractRepoName(url)
}
