package workspaces

import (
	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// ResolutionStrategy defines the interface for resolving repository identifiers.
// Strategies are tried in order until one succeeds.
type ResolutionStrategy interface {
	// Name returns a human-readable name for debugging and logging.
	Name() string

	// Resolve attempts to resolve the input string to a domain.Repo.
	// Returns the resolved repo and true if successful, or an empty repo and false if
	// this strategy cannot handle the input.
	Resolve(input string) (domain.Repo, bool)
}
