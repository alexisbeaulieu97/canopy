package workspaces

import (
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/giturl"
)

// URLStrategy resolves direct Git URLs to repositories.
type URLStrategy struct {
	// registryLookup optionally checks the registry to find an alias for a URL.
	registryLookup URLRegistryLookup
}

// URLRegistryLookup is a function that looks up a URL in the registry.
// It returns the alias and URL if found, or empty values and false otherwise.
type URLRegistryLookup func(url string) (alias, resolvedURL string, found bool)

// NewURLStrategy creates a new URL resolution strategy.
// If registryLookup is nil, URLs will be resolved without registry lookup.
func NewURLStrategy(registryLookup URLRegistryLookup) *URLStrategy {
	return &URLStrategy{
		registryLookup: registryLookup,
	}
}

// Name returns the strategy name for debugging.
func (s *URLStrategy) Name() string {
	return "url"
}

// Resolve attempts to resolve a URL to a repository.
func (s *URLStrategy) Resolve(input string) (domain.Repo, bool) {
	if !giturl.IsURL(input) {
		return domain.Repo{}, false
	}

	// Try registry lookup first to get an alias if available
	if s.registryLookup != nil {
		if alias, url, ok := s.registryLookup(input); ok {
			return domain.Repo{Name: alias, URL: url}, true
		}
	}

	// Fall back to deriving name from URL
	return domain.Repo{Name: giturl.ExtractRepoName(input), URL: input}, true
}
