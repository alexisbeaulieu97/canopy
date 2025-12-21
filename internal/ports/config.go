// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import (
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
)

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

	// GetParallelWorkers returns the number of parallel workers for repository operations.
	GetParallelWorkers() int

	// GetLockTimeout returns the workspace lock timeout.
	GetLockTimeout() time.Duration

	// GetLockStaleThreshold returns the stale lock threshold.
	GetLockStaleThreshold() time.Duration

	// GetRegistry returns the repository registry.
	GetRegistry() *config.RepoRegistry

	// GetHooks returns the lifecycle hooks configuration.
	GetHooks() config.Hooks

	// GetKeybindings returns the TUI keybindings with defaults applied.
	GetKeybindings() config.Keybindings

	// GetUseEmoji returns whether emoji should be used in the TUI.
	GetUseEmoji() bool

	// GetGitRetryConfig returns the parsed git retry configuration.
	GetGitRetryConfig() config.ParsedRetryConfig

	// GetTemplates returns configured workspace templates.
	GetTemplates() map[string]config.Template

	// ResolveTemplate returns a template by name.
	ResolveTemplate(name string) (config.Template, error)

	// ValidateTemplates validates template definitions.
	ValidateTemplates() error
}
