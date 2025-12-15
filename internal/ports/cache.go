// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import "github.com/alexisbeaulieu97/canopy/internal/domain"

// WorkspaceCache defines the interface for in-memory caching of workspace metadata.
type WorkspaceCache interface {
	// Get retrieves a workspace from the cache by ID.
	// Returns a deep copy of the workspace, directory name, and a boolean indicating if the entry was found and valid.
	Get(id string) (*domain.Workspace, string, bool)

	// Set adds or updates a workspace in the cache.
	Set(id string, ws *domain.Workspace, dirName string)

	// Invalidate removes a specific workspace from the cache.
	Invalidate(id string)

	// InvalidateAll removes all entries from the cache.
	InvalidateAll()

	// Size returns the number of entries currently in the cache.
	Size() int
}
