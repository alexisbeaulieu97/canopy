// Package mocks provides mock implementations for testing.
package mocks

import (
	"sync"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockWorkspaceCache implements ports.WorkspaceCache.
var _ ports.WorkspaceCache = (*MockWorkspaceCache)(nil)

// MockWorkspaceCache is a mock implementation of ports.WorkspaceCache for testing.
// It is safe for concurrent use.
type MockWorkspaceCache struct {
	mu sync.RWMutex

	// GetFunc is called when Get is invoked.
	GetFunc func(id string) (*domain.Workspace, string, bool)

	// SetFunc is called when Set is invoked.
	SetFunc func(id string, ws *domain.Workspace, dirName string)

	// InvalidateFunc is called when Invalidate is invoked.
	InvalidateFunc func(id string)

	// InvalidateAllFunc is called when InvalidateAll is invoked.
	InvalidateAllFunc func()

	// SizeFunc is called when Size is invoked.
	SizeFunc func() int

	// GetCalls records all calls to Get for verification.
	GetCalls []string

	// SetCalls records all calls to Set for verification.
	SetCalls []SetCall

	// InvalidateCalls records all calls to Invalidate for verification.
	InvalidateCalls []string

	// InvalidateAllCalls records the number of times InvalidateAll was called.
	InvalidateAllCalls int

	// SizeCalls records the number of times Size was called.
	SizeCalls int

	// Internal storage for simple mock behavior
	entries map[string]cacheEntry
}

// cacheEntry holds a cached workspace for the mock.
type cacheEntry struct {
	workspace *domain.Workspace
	dirName   string
}

// SetCall records a single call to Set.
type SetCall struct {
	ID        string
	Workspace *domain.Workspace
	DirName   string
}

// NewMockWorkspaceCache creates a new MockWorkspaceCache with sensible defaults.
func NewMockWorkspaceCache() *MockWorkspaceCache {
	return &MockWorkspaceCache{
		GetCalls:        make([]string, 0),
		SetCalls:        make([]SetCall, 0),
		InvalidateCalls: make([]string, 0),
		entries:         make(map[string]cacheEntry),
	}
}

// Get calls the mock function if set, otherwise returns from internal storage.
func (m *MockWorkspaceCache) Get(id string) (*domain.Workspace, string, bool) {
	m.mu.Lock()
	m.GetCalls = append(m.GetCalls, id)
	m.mu.Unlock()

	if m.GetFunc != nil {
		return m.GetFunc(id)
	}

	m.mu.RLock()
	entry, ok := m.entries[id]
	m.mu.RUnlock()

	if !ok {
		return nil, "", false
	}

	return entry.workspace, entry.dirName, true
}

// Set calls the mock function if set, otherwise stores in internal storage.
func (m *MockWorkspaceCache) Set(id string, ws *domain.Workspace, dirName string) {
	m.mu.Lock()
	m.SetCalls = append(m.SetCalls, SetCall{
		ID:        id,
		Workspace: ws,
		DirName:   dirName,
	})
	m.mu.Unlock()

	if m.SetFunc != nil {
		m.SetFunc(id, ws, dirName)
		return
	}

	m.mu.Lock()
	m.entries[id] = cacheEntry{
		workspace: ws,
		dirName:   dirName,
	}
	m.mu.Unlock()
}

// Invalidate calls the mock function if set, otherwise removes from internal storage.
func (m *MockWorkspaceCache) Invalidate(id string) {
	m.mu.Lock()
	m.InvalidateCalls = append(m.InvalidateCalls, id)
	m.mu.Unlock()

	if m.InvalidateFunc != nil {
		m.InvalidateFunc(id)
		return
	}

	m.mu.Lock()
	delete(m.entries, id)
	m.mu.Unlock()
}

// InvalidateAll calls the mock function if set, otherwise clears internal storage.
func (m *MockWorkspaceCache) InvalidateAll() {
	m.mu.Lock()
	m.InvalidateAllCalls++
	m.mu.Unlock()

	if m.InvalidateAllFunc != nil {
		m.InvalidateAllFunc()
		return
	}

	m.mu.Lock()
	m.entries = make(map[string]cacheEntry)
	m.mu.Unlock()
}

// Size calls the mock function if set, otherwise returns internal storage size.
func (m *MockWorkspaceCache) Size() int {
	m.mu.Lock()
	m.SizeCalls++
	m.mu.Unlock()

	if m.SizeFunc != nil {
		return m.SizeFunc()
	}

	m.mu.RLock()
	size := len(m.entries)
	m.mu.RUnlock()

	return size
}

// ResetCalls clears all recorded calls.
func (m *MockWorkspaceCache) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetCalls = make([]string, 0)
	m.SetCalls = make([]SetCall, 0)
	m.InvalidateCalls = make([]string, 0)
	m.InvalidateAllCalls = 0
	m.SizeCalls = 0
}

// ResetStorage clears the internal storage.
func (m *MockWorkspaceCache) ResetStorage() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make(map[string]cacheEntry)
}
