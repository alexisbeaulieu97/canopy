// Package mocks provides mock implementations for testing.
package mocks

import (
	"fmt"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
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
	ComputeWorkspaceDirFunc   func(id string) (string, error)
	GetStaleThresholdDaysFunc func() int
	GetParallelWorkersFunc    func() int
	GetLockTimeoutFunc        func() time.Duration
	GetLockStaleThresholdFunc func() time.Duration
	GetRegistryFunc           func() *config.RepoRegistry
	GetHooksFunc              func() config.Hooks
	GetTemplatesFunc          func() map[string]config.Template
	ResolveTemplateFunc       func(name string) (config.Template, error)
	ValidateTemplatesFunc     func() error

	// Default values for simple getters.
	ProjectsRoot       string
	WorkspacesRoot     string
	ClosedRoot         string
	CloseDefault       string
	WorkspaceNaming    string
	StaleThresholdDays int
	ParallelWorkers    int
	LockTimeout        time.Duration
	LockStaleThreshold time.Duration
	Registry           *config.RepoRegistry
	Hooks              config.Hooks
	Templates          map[string]config.Template
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
		ParallelWorkers:    config.DefaultParallelWorkers,
		LockTimeout:        0,
		LockStaleThreshold: 0,
		Registry:           &config.RepoRegistry{},
		Templates:          map[string]config.Template{},
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

// ComputeWorkspaceDir calls the mock function if set, otherwise returns the normalized ID.
func (m *MockConfigProvider) ComputeWorkspaceDir(id string) (string, error) {
	if m.ComputeWorkspaceDirFunc != nil {
		return m.ComputeWorkspaceDirFunc(id)
	}

	return validation.NormalizeWorkspaceDirName(id)
}

// GetStaleThresholdDays calls the mock function if set, otherwise returns StaleThresholdDays.
func (m *MockConfigProvider) GetStaleThresholdDays() int {
	if m.GetStaleThresholdDaysFunc != nil {
		return m.GetStaleThresholdDaysFunc()
	}

	return m.StaleThresholdDays
}

// GetParallelWorkers calls the mock function if set, otherwise returns ParallelWorkers.
func (m *MockConfigProvider) GetParallelWorkers() int {
	if m.GetParallelWorkersFunc != nil {
		return m.GetParallelWorkersFunc()
	}

	return m.ParallelWorkers
}

// GetLockTimeout calls the mock function if set, otherwise returns LockTimeout.
func (m *MockConfigProvider) GetLockTimeout() time.Duration {
	if m.GetLockTimeoutFunc != nil {
		return m.GetLockTimeoutFunc()
	}

	return m.LockTimeout
}

// GetLockStaleThreshold calls the mock function if set, otherwise returns LockStaleThreshold.
func (m *MockConfigProvider) GetLockStaleThreshold() time.Duration {
	if m.GetLockStaleThresholdFunc != nil {
		return m.GetLockStaleThresholdFunc()
	}

	return m.LockStaleThreshold
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

// GetTemplates calls the mock function if set, otherwise returns Templates.
func (m *MockConfigProvider) GetTemplates() map[string]config.Template {
	if m.GetTemplatesFunc != nil {
		return m.GetTemplatesFunc()
	}

	return m.Templates
}

// ResolveTemplate calls the mock function if set, otherwise resolves from Templates.
func (m *MockConfigProvider) ResolveTemplate(name string) (config.Template, error) {
	if m.ResolveTemplateFunc != nil {
		return m.ResolveTemplateFunc(name)
	}

	if tmpl, ok := m.Templates[name]; ok {
		tmpl.Name = name
		return tmpl, nil
	}

	return config.Template{}, fmt.Errorf("template %q not found", name)
}

// ValidateTemplates calls the mock function if set, otherwise returns nil.
func (m *MockConfigProvider) ValidateTemplates() error {
	if m.ValidateTemplatesFunc != nil {
		return m.ValidateTemplatesFunc()
	}

	return nil
}

// GetKeybindings returns the TUI keybindings with defaults applied.
func (m *MockConfigProvider) GetKeybindings() config.Keybindings {
	return config.Keybindings{}.WithDefaults()
}

// GetUseEmoji returns whether emoji should be used in the TUI (defaults to true).
func (m *MockConfigProvider) GetUseEmoji() bool {
	return true
}

// GetGitRetryConfig returns default git retry configuration.
func (m *MockConfigProvider) GetGitRetryConfig() config.ParsedRetryConfig {
	return config.ParsedRetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.25,
	}
}
