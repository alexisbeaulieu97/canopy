// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"sync"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// DefaultCacheTTL is the default time-to-live for cache entries.
const DefaultCacheTTL = 30 * time.Second

// cacheEntry holds a cached workspace along with its expiration time and directory name.
type cacheEntry struct {
	workspace *domain.Workspace
	dirName   string
	expiresAt time.Time
}

// WorkspaceCache provides in-memory caching of workspace metadata with TTL support.
type WorkspaceCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

// NewWorkspaceCache creates a new workspace cache with the specified TTL.
// If ttl is 0, DefaultCacheTTL is used.
func NewWorkspaceCache(ttl time.Duration) *WorkspaceCache {
	if ttl == 0 {
		ttl = DefaultCacheTTL
	}

	return &WorkspaceCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a workspace from the cache by ID.
// Returns the workspace, directory name, and a boolean indicating if the entry was found and valid.
func (c *WorkspaceCache) Get(id string) (*domain.Workspace, string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[id]
	if !ok {
		return nil, "", false
	}

	if time.Now().After(entry.expiresAt) {
		// Entry has expired; caller should reload from storage
		return nil, "", false
	}

	return entry.workspace, entry.dirName, true
}

// Set adds or updates a workspace in the cache.
func (c *WorkspaceCache) Set(id string, ws *domain.Workspace, dirName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[id] = cacheEntry{
		workspace: ws,
		dirName:   dirName,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a specific workspace from the cache.
func (c *WorkspaceCache) Invalidate(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, id)
}

// InvalidateAll removes all entries from the cache.
func (c *WorkspaceCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

// Size returns the number of entries currently in the cache (including expired ones).
func (c *WorkspaceCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}
