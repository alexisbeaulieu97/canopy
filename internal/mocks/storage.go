// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockWorkspaceStorage implements ports.WorkspaceStorage.
var _ ports.WorkspaceStorage = (*MockWorkspaceStorage)(nil)

// MockWorkspaceStorage is a mock implementation of ports.WorkspaceStorage for testing.
type MockWorkspaceStorage struct {
	CreateFunc       func(ctx context.Context, ws domain.Workspace) error
	SaveFunc         func(ctx context.Context, ws domain.Workspace) error
	CloseFunc        func(ctx context.Context, id string, closedAt time.Time) (*domain.ClosedWorkspace, error)
	ListFunc         func(ctx context.Context) ([]domain.Workspace, error)
	ListClosedFunc   func(ctx context.Context) ([]domain.ClosedWorkspace, error)
	LoadFunc         func(ctx context.Context, id string) (*domain.Workspace, error)
	DeleteFunc       func(ctx context.Context, id string) error
	RenameFunc       func(ctx context.Context, oldID, newID string) error
	LatestClosedFunc func(ctx context.Context, id string) (*domain.ClosedWorkspace, error)
	DeleteClosedFunc func(ctx context.Context, id string, closedAt time.Time) error

	// Workspaces holds test workspace data keyed by ID.
	Workspaces map[string]domain.Workspace
}

// NewMockWorkspaceStorage creates a new MockWorkspaceStorage with initialized maps.
func NewMockWorkspaceStorage() *MockWorkspaceStorage {
	return &MockWorkspaceStorage{
		Workspaces: make(map[string]domain.Workspace),
	}
}

// Create calls the mock function if set, otherwise stores in Workspaces.
func (m *MockWorkspaceStorage) Create(ctx context.Context, ws domain.Workspace) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, ws)
	}

	m.Workspaces[ws.ID] = ws

	return nil
}

// Save calls the mock function if set, otherwise updates Workspaces.
func (m *MockWorkspaceStorage) Save(ctx context.Context, ws domain.Workspace) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, ws)
	}

	m.Workspaces[ws.ID] = ws

	return nil
}

// Close calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) Close(ctx context.Context, id string, closedAt time.Time) (*domain.ClosedWorkspace, error) {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx, id, closedAt)
	}

	ws, ok := m.Workspaces[id]
	if !ok {
		return nil, fmt.Errorf("workspace %s not found", id)
	}

	ws.ClosedAt = &closedAt
	m.Workspaces[id] = ws // Persist the modification back to the map

	return &domain.ClosedWorkspace{
		DirName:  id,
		Path:     "/closed/" + id,
		Metadata: ws,
	}, nil
}

// List calls the mock function if set, otherwise returns Workspaces as slice.
func (m *MockWorkspaceStorage) List(ctx context.Context) ([]domain.Workspace, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}

	result := make([]domain.Workspace, 0, len(m.Workspaces))
	for _, ws := range m.Workspaces {
		result = append(result, ws)
	}

	// Sort by ID for deterministic test behavior
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

// ListClosed calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) ListClosed(ctx context.Context) ([]domain.ClosedWorkspace, error) {
	if m.ListClosedFunc != nil {
		return m.ListClosedFunc(ctx)
	}

	return nil, nil
}

// Load calls the mock function if set, otherwise returns from Workspaces.
func (m *MockWorkspaceStorage) Load(ctx context.Context, id string) (*domain.Workspace, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(ctx, id)
	}

	if ws, ok := m.Workspaces[id]; ok {
		return &ws, nil
	}

	return nil, fmt.Errorf("workspace %s not found", id)
}

// Delete calls the mock function if set, otherwise removes from Workspaces.
func (m *MockWorkspaceStorage) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}

	delete(m.Workspaces, id)

	return nil
}

// Rename calls the mock function if set, otherwise renames in Workspaces.
func (m *MockWorkspaceStorage) Rename(ctx context.Context, oldID, newID string) error {
	if m.RenameFunc != nil {
		return m.RenameFunc(ctx, oldID, newID)
	}

	ws, ok := m.Workspaces[oldID]
	if !ok {
		return fmt.Errorf("workspace %s not found", oldID)
	}

	delete(m.Workspaces, oldID)

	ws.ID = newID
	m.Workspaces[newID] = ws

	return nil
}

// LatestClosed calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) LatestClosed(ctx context.Context, id string) (*domain.ClosedWorkspace, error) {
	if m.LatestClosedFunc != nil {
		return m.LatestClosedFunc(ctx, id)
	}

	return nil, nil
}

// DeleteClosed calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) DeleteClosed(ctx context.Context, id string, closedAt time.Time) error {
	if m.DeleteClosedFunc != nil {
		return m.DeleteClosedFunc(ctx, id, closedAt)
	}

	return nil
}
