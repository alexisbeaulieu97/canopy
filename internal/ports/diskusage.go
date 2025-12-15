// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import "time"

// DiskUsage defines the interface for calculating and caching disk usage.
type DiskUsage interface {
	// CachedUsage returns cached disk usage/mtime with a short TTL to avoid repeated scans.
	CachedUsage(root string) (int64, time.Time, error)

	// Calculate sums file sizes under the provided root and returns latest mtime.
	Calculate(root string) (int64, time.Time, error)

	// InvalidateCache clears the cache entry for a specific path.
	InvalidateCache(root string)

	// ClearCache clears all cached entries.
	ClearCache()
}
