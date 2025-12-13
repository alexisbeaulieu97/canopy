// Package mocks provides mock implementations for testing.
package mocks

import (
	"fmt"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockWorkspaceStorage implements ports.WorkspaceStorage.
var _ ports.WorkspaceStorage = (*MockWorkspaceStorage)(nil)

// MockWorkspaceStorage is a mock implementation of ports.WorkspaceStorage for testing.
type MockWorkspaceStorage struct {
	CreateFunc       func(dirName, id, branchName string, repos []domain.Repo) error
	SaveFunc         func(dirName string, ws domain.Workspace) error
	CloseFunc        func(dirName string, ws domain.Workspace, closedAt time.Time) (*domain.ClosedWorkspace, error)
	ListFunc         func() (map[string]domain.Workspace, error)
	ListClosedFunc   func() ([]domain.ClosedWorkspace, error)
	LoadFunc         func(dirName string) (*domain.Workspace, error)
	LoadByIDFunc     func(id string) (*domain.Workspace, string, error)
	DeleteFunc       func(workspaceID string) error
	LatestClosedFunc func(workspaceID string) (*domain.ClosedWorkspace, error)
	DeleteClosedFunc func(path string) error

	// Workspaces holds test workspace data.
	Workspaces map[string]domain.Workspace
}

// NewMockWorkspaceStorage creates a new MockWorkspaceStorage with initialized maps.
func NewMockWorkspaceStorage() *MockWorkspaceStorage {
	return &MockWorkspaceStorage{
		Workspaces: make(map[string]domain.Workspace),
	}
}

// Create calls the mock function if set, otherwise stores in Workspaces.
func (m *MockWorkspaceStorage) Create(dirName, id, branchName string, repos []domain.Repo) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(dirName, id, branchName, repos)
	}

	m.Workspaces[dirName] = domain.Workspace{
		ID:         id,
		BranchName: branchName,
		Repos:      repos,
	}

	return nil
}

// Save calls the mock function if set, otherwise updates Workspaces.
func (m *MockWorkspaceStorage) Save(dirName string, ws domain.Workspace) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(dirName, ws)
	}

	m.Workspaces[dirName] = ws

	return nil
}

// Close calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) Close(dirName string, ws domain.Workspace, closedAt time.Time) (*domain.ClosedWorkspace, error) {
	if m.CloseFunc != nil {
		return m.CloseFunc(dirName, ws, closedAt)
	}

	return &domain.ClosedWorkspace{
		DirName:  dirName,
		Path:     "/closed/" + dirName,
		Metadata: ws,
	}, nil
}

// List calls the mock function if set, otherwise returns Workspaces.
func (m *MockWorkspaceStorage) List() (map[string]domain.Workspace, error) {
	if m.ListFunc != nil {
		return m.ListFunc()
	}

	return m.Workspaces, nil
}

// ListClosed calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) ListClosed() ([]domain.ClosedWorkspace, error) {
	if m.ListClosedFunc != nil {
		return m.ListClosedFunc()
	}

	return nil, nil
}

// Load calls the mock function if set, otherwise returns from Workspaces.
func (m *MockWorkspaceStorage) Load(dirName string) (*domain.Workspace, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(dirName)
	}

	if ws, ok := m.Workspaces[dirName]; ok {
		return &ws, nil
	}

	return nil, fmt.Errorf("failed to open workspace metadata: workspace %s not found", dirName)
}

// LoadByID calls the mock function if set, otherwise searches Workspaces by ID.
func (m *MockWorkspaceStorage) LoadByID(id string) (*domain.Workspace, string, error) {
	if m.LoadByIDFunc != nil {
		return m.LoadByIDFunc(id)
	}

	// Default implementation: search workspaces by ID
	for dirName, ws := range m.Workspaces {
		if ws.ID == id {
			return &ws, dirName, nil
		}
	}

	return nil, "", fmt.Errorf("workspace %s not found", id)
}

// Delete calls the mock function if set, otherwise removes from Workspaces.
func (m *MockWorkspaceStorage) Delete(workspaceID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(workspaceID)
	}

	delete(m.Workspaces, workspaceID)

	return nil
}

// LatestClosed calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) LatestClosed(workspaceID string) (*domain.ClosedWorkspace, error) {
	if m.LatestClosedFunc != nil {
		return m.LatestClosedFunc(workspaceID)
	}

	return nil, nil
}

// DeleteClosed calls the mock function if set, otherwise returns nil.
func (m *MockWorkspaceStorage) DeleteClosed(path string) error {
	if m.DeleteClosedFunc != nil {
		return m.DeleteClosedFunc(path)
	}

	return nil
}
