// Package mocks provides mock implementations for testing.
package mocks

import (
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockDiskUsage implements ports.DiskUsage.
var _ ports.DiskUsage = (*MockDiskUsage)(nil)

// MockDiskUsage is a mock implementation of ports.DiskUsage for testing.
type MockDiskUsage struct {
	// CachedUsageFunc is called when CachedUsage is invoked.
	CachedUsageFunc func(root string) (int64, time.Time, error)

	// CalculateFunc is called when Calculate is invoked.
	CalculateFunc func(root string) (int64, time.Time, error)

	// InvalidateCacheFunc is called when InvalidateCache is invoked.
	InvalidateCacheFunc func(root string)

	// ClearCacheFunc is called when ClearCache is invoked.
	ClearCacheFunc func()

	// CachedUsageCalls records all calls to CachedUsage for verification.
	CachedUsageCalls []string

	// CalculateCalls records all calls to Calculate for verification.
	CalculateCalls []string

	// InvalidateCacheCalls records all calls to InvalidateCache for verification.
	InvalidateCacheCalls []string

	// ClearCacheCalls records the number of times ClearCache was called.
	ClearCacheCalls int

	// Default return values
	DefaultUsage   int64
	DefaultModTime time.Time
	DefaultErr     error
}

// NewMockDiskUsage creates a new MockDiskUsage with sensible defaults.
func NewMockDiskUsage() *MockDiskUsage {
	return &MockDiskUsage{
		CachedUsageCalls:     make([]string, 0),
		CalculateCalls:       make([]string, 0),
		InvalidateCacheCalls: make([]string, 0),
		DefaultUsage:         0,
		DefaultModTime:       time.Now(),
	}
}

// CachedUsage calls the mock function if set, otherwise returns default values.
func (m *MockDiskUsage) CachedUsage(root string) (int64, time.Time, error) {
	m.CachedUsageCalls = append(m.CachedUsageCalls, root)

	if m.CachedUsageFunc != nil {
		return m.CachedUsageFunc(root)
	}

	return m.DefaultUsage, m.DefaultModTime, m.DefaultErr
}

// Calculate calls the mock function if set, otherwise returns default values.
func (m *MockDiskUsage) Calculate(root string) (int64, time.Time, error) {
	m.CalculateCalls = append(m.CalculateCalls, root)

	if m.CalculateFunc != nil {
		return m.CalculateFunc(root)
	}

	return m.DefaultUsage, m.DefaultModTime, m.DefaultErr
}

// InvalidateCache calls the mock function if set.
func (m *MockDiskUsage) InvalidateCache(root string) {
	m.InvalidateCacheCalls = append(m.InvalidateCacheCalls, root)

	if m.InvalidateCacheFunc != nil {
		m.InvalidateCacheFunc(root)
	}
}

// ClearCache calls the mock function if set.
func (m *MockDiskUsage) ClearCache() {
	m.ClearCacheCalls++

	if m.ClearCacheFunc != nil {
		m.ClearCacheFunc()
	}
}

// ResetCalls clears all recorded calls.
func (m *MockDiskUsage) ResetCalls() {
	m.CachedUsageCalls = make([]string, 0)
	m.CalculateCalls = make([]string, 0)
	m.InvalidateCacheCalls = make([]string, 0)
	m.ClearCacheCalls = 0
}
