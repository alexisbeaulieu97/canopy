// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// usageEntry caches disk usage data for a directory.
type usageEntry struct {
	usage     int64
	lastMod   time.Time
	scannedAt time.Time
	err       error
}

// DiskUsageCalculator calculates and caches disk usage for directories.
type DiskUsageCalculator struct {
	cache map[string]usageEntry
	mu    sync.Mutex
	ttl   time.Duration
}

// NewDiskUsageCalculator creates a new DiskUsageCalculator with the specified cache TTL.
func NewDiskUsageCalculator(ttl time.Duration) *DiskUsageCalculator {
	return &DiskUsageCalculator{
		cache: make(map[string]usageEntry),
		ttl:   ttl,
	}
}

// DefaultDiskUsageCalculator creates a DiskUsageCalculator with a 1-minute TTL.
func DefaultDiskUsageCalculator() *DiskUsageCalculator {
	return NewDiskUsageCalculator(time.Minute)
}

// CachedUsage returns cached disk usage/mtime with a short TTL to avoid repeated scans.
func (d *DiskUsageCalculator) CachedUsage(root string) (int64, time.Time, error) {
	d.mu.Lock()

	entry, ok := d.cache[root]
	if ok && time.Since(entry.scannedAt) < d.ttl {
		d.mu.Unlock()
		return entry.usage, entry.lastMod, entry.err
	}

	d.mu.Unlock()

	usage, latest, err := d.Calculate(root)

	d.mu.Lock()
	d.cache[root] = usageEntry{
		usage:     usage,
		lastMod:   latest,
		scannedAt: time.Now(),
		err:       err,
	}
	d.mu.Unlock()

	return usage, latest, err
}

// Calculate sums file sizes under the provided root and returns latest mtime.
// Note: .git directories are skipped so LastModified reflects working tree activity.
func (d *DiskUsageCalculator) Calculate(root string) (int64, time.Time, error) {
	var (
		total   int64
		latest  time.Time
		skipDir = map[string]struct{}{".git": {}}
	)

	err := filepath.WalkDir(root, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if _, ok := skipDir[entry.Name()]; ok {
				return fs.SkipDir
			}

			return nil
		}

		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}

		total += info.Size()

		if mod := info.ModTime(); mod.After(latest) {
			latest = mod
		}

		return nil
	})
	if err != nil {
		return 0, time.Time{}, err
	}

	return total, latest, nil
}

// InvalidateCache clears the cache entry for a specific path.
func (d *DiskUsageCalculator) InvalidateCache(root string) {
	d.mu.Lock()
	delete(d.cache, root)
	d.mu.Unlock()
}

// ClearCache clears all cached entries.
func (d *DiskUsageCalculator) ClearCache() {
	d.mu.Lock()
	d.cache = make(map[string]usageEntry)
	d.mu.Unlock()
}
