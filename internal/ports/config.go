// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import "github.com/alexisbeaulieu97/canopy/internal/config"

// ConfigProvider defines the interface for configuration access.
type ConfigProvider interface {
	// GetReposForWorkspace returns default repos for a given workspace ID based on patterns.
	GetReposForWorkspace(workspaceID string) []string

	// Validate checks if the configuration is valid.
	Validate() error

	// GetProjectsRoot returns the projects root directory.
	GetProjectsRoot() string

	// GetWorkspacesRoot returns the workspaces root directory.
	GetWorkspacesRoot() string

	// GetClosedRoot returns the closed workspaces root directory.
	GetClosedRoot() string

	// GetCloseDefault returns the default close behavior.
	GetCloseDefault() string

	// GetWorkspaceNaming returns the workspace naming pattern.
	GetWorkspaceNaming() string

	// GetStaleThresholdDays returns the stale threshold in days.
	GetStaleThresholdDays() int

	// GetRegistry returns the repository registry.
	GetRegistry() *config.RepoRegistry
}
