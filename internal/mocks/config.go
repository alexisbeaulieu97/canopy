// Package mocks provides mock implementations for testing.
package mocks

import (
	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Compile-time check that MockConfigProvider implements ports.ConfigProvider.
var _ ports.ConfigProvider = (*MockConfigProvider)(nil)

// MockConfigProvider is a mock implementation of ports.ConfigProvider for testing.
type MockConfigProvider struct {
	GetReposForWorkspaceFunc  func(workspaceID string) []string
	ValidateFunc              func() error
	GetProjectsRootFunc       func() string
	GetWorkspacesRootFunc     func() string
	GetClosedRootFunc         func() string
	GetCloseDefaultFunc       func() string
	GetWorkspaceNamingFunc    func() string
	GetStaleThresholdDaysFunc func() int
	GetRegistryFunc           func() *config.RepoRegistry
	GetHooksFunc              func() config.Hooks

	// Default values for simple getters.
	ProjectsRoot       string
	WorkspacesRoot     string
	ClosedRoot         string
	CloseDefault       string
	WorkspaceNaming    string
	StaleThresholdDays int
	Registry           *config.RepoRegistry
	Hooks              config.Hooks
	RepoNames          []string
}

// NewMockConfigProvider creates a new MockConfigProvider with sensible defaults.
func NewMockConfigProvider() *MockConfigProvider {
	return &MockConfigProvider{
		ProjectsRoot:       "/projects",
		WorkspacesRoot:     "/workspaces",
		ClosedRoot:         "/closed",
		CloseDefault:       "archive",
		WorkspaceNaming:    "{{.ID}}",
		StaleThresholdDays: 14,
		Registry:           &config.RepoRegistry{},
		RepoNames:          []string{},
	}
}

// GetReposForWorkspace calls the mock function if set, otherwise returns RepoNames.
func (m *MockConfigProvider) GetReposForWorkspace(workspaceID string) []string {
	if m.GetReposForWorkspaceFunc != nil {
		return m.GetReposForWorkspaceFunc(workspaceID)
	}

	return m.RepoNames
}

// Validate calls the mock function if set, otherwise returns nil.
func (m *MockConfigProvider) Validate() error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc()
	}

	return nil
}

// GetProjectsRoot calls the mock function if set, otherwise returns ProjectsRoot.
func (m *MockConfigProvider) GetProjectsRoot() string {
	if m.GetProjectsRootFunc != nil {
		return m.GetProjectsRootFunc()
	}

	return m.ProjectsRoot
}

// GetWorkspacesRoot calls the mock function if set, otherwise returns WorkspacesRoot.
func (m *MockConfigProvider) GetWorkspacesRoot() string {
	if m.GetWorkspacesRootFunc != nil {
		return m.GetWorkspacesRootFunc()
	}

	return m.WorkspacesRoot
}

// GetClosedRoot calls the mock function if set, otherwise returns ClosedRoot.
func (m *MockConfigProvider) GetClosedRoot() string {
	if m.GetClosedRootFunc != nil {
		return m.GetClosedRootFunc()
	}

	return m.ClosedRoot
}

// GetCloseDefault calls the mock function if set, otherwise returns CloseDefault.
func (m *MockConfigProvider) GetCloseDefault() string {
	if m.GetCloseDefaultFunc != nil {
		return m.GetCloseDefaultFunc()
	}

	return m.CloseDefault
}

// GetWorkspaceNaming calls the mock function if set, otherwise returns WorkspaceNaming.
func (m *MockConfigProvider) GetWorkspaceNaming() string {
	if m.GetWorkspaceNamingFunc != nil {
		return m.GetWorkspaceNamingFunc()
	}

	return m.WorkspaceNaming
}

// GetStaleThresholdDays calls the mock function if set, otherwise returns StaleThresholdDays.
func (m *MockConfigProvider) GetStaleThresholdDays() int {
	if m.GetStaleThresholdDaysFunc != nil {
		return m.GetStaleThresholdDaysFunc()
	}

	return m.StaleThresholdDays
}

// GetRegistry calls the mock function if set, otherwise returns Registry.
func (m *MockConfigProvider) GetRegistry() *config.RepoRegistry {
	if m.GetRegistryFunc != nil {
		return m.GetRegistryFunc()
	}

	return m.Registry
}

// GetHooks calls the mock function if set, otherwise returns Hooks.
func (m *MockConfigProvider) GetHooks() config.Hooks {
	if m.GetHooksFunc != nil {
		return m.GetHooksFunc()
	}

	return m.Hooks
}

// GetKeybindings returns the TUI keybindings with defaults applied.
func (m *MockConfigProvider) GetKeybindings() config.Keybindings {
	return config.Keybindings{}.WithDefaults()
}
