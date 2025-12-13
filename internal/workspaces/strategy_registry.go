package workspaces

import (
	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// RegistryStrategy resolves repository aliases from a registry.
type RegistryStrategy struct {
	lookup RegistryLookup
}

// RegistryLookup is a function that looks up an alias in the registry.
// It returns the alias and URL if found, or empty values and false otherwise.
type RegistryLookup func(alias string) (resolvedAlias, url string, found bool)

// NewRegistryStrategy creates a new registry resolution strategy.
func NewRegistryStrategy(lookup RegistryLookup) *RegistryStrategy {
	return &RegistryStrategy{
		lookup: lookup,
	}
}

// Name returns the strategy name for debugging.
func (s *RegistryStrategy) Name() string {
	return "registry"
}

// Resolve attempts to resolve an alias from the registry.
func (s *RegistryStrategy) Resolve(input string) (domain.Repo, bool) {
	if s.lookup == nil {
		return domain.Repo{}, false
	}

	if alias, url, ok := s.lookup(input); ok {
		return domain.Repo{Name: alias, URL: url}, true
	}

	return domain.Repo{}, false
}
