// Package mocks provides mock implementations for testing.
package mocks

import (
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockWorkspaceCache implements ports.WorkspaceCache.
var _ ports.WorkspaceCache = (*MockWorkspaceCache)(nil)

// MockWorkspaceCache is a mock implementation of ports.WorkspaceCache for testing.
type MockWorkspaceCache struct {
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
	m.GetCalls = append(m.GetCalls, id)

	if m.GetFunc != nil {
		return m.GetFunc(id)
	}

	entry, ok := m.entries[id]
	if !ok {
		return nil, "", false
	}

	return entry.workspace, entry.dirName, true
}

// Set calls the mock function if set, otherwise stores in internal storage.
func (m *MockWorkspaceCache) Set(id string, ws *domain.Workspace, dirName string) {
	m.SetCalls = append(m.SetCalls, SetCall{
		ID:        id,
		Workspace: ws,
		DirName:   dirName,
	})

	if m.SetFunc != nil {
		m.SetFunc(id, ws, dirName)
		return
	}

	m.entries[id] = cacheEntry{
		workspace: ws,
		dirName:   dirName,
	}
}

// Invalidate calls the mock function if set, otherwise removes from internal storage.
func (m *MockWorkspaceCache) Invalidate(id string) {
	m.InvalidateCalls = append(m.InvalidateCalls, id)

	if m.InvalidateFunc != nil {
		m.InvalidateFunc(id)
		return
	}

	delete(m.entries, id)
}

// InvalidateAll calls the mock function if set, otherwise clears internal storage.
func (m *MockWorkspaceCache) InvalidateAll() {
	m.InvalidateAllCalls++

	if m.InvalidateAllFunc != nil {
		m.InvalidateAllFunc()
		return
	}

	m.entries = make(map[string]cacheEntry)
}

// Size calls the mock function if set, otherwise returns internal storage size.
func (m *MockWorkspaceCache) Size() int {
	m.SizeCalls++

	if m.SizeFunc != nil {
		return m.SizeFunc()
	}

	return len(m.entries)
}

// ResetCalls clears all recorded calls.
func (m *MockWorkspaceCache) ResetCalls() {
	m.GetCalls = make([]string, 0)
	m.SetCalls = make([]SetCall, 0)
	m.InvalidateCalls = make([]string, 0)
	m.InvalidateAllCalls = 0
	m.SizeCalls = 0
}

// ResetStorage clears the internal storage.
func (m *MockWorkspaceCache) ResetStorage() {
	m.entries = make(map[string]cacheEntry)
}
