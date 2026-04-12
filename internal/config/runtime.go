package config

import (
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

var defaultPortGitRetryConfig = ports.GitRetryConfig{
	MaxAttempts:  3,
	InitialDelay: time.Second,
	MaxDelay:     30 * time.Second,
	Multiplier:   2.0,
	JitterFactor: 0.25,
}

// GetProjectsRoot returns the projects root directory.
func (c *Config) GetProjectsRoot() string {
	return c.ProjectsRoot
}

// GetWorkspacesRoot returns the workspaces root directory.
func (c *Config) GetWorkspacesRoot() string {
	return c.WorkspacesRoot
}

// GetClosedRoot returns the closed workspaces root directory.
func (c *Config) GetClosedRoot() string {
	return c.ClosedRoot
}

// GetCloseDefault returns the default close behavior.
func (c *Config) GetCloseDefault() string {
	return c.CloseDefault
}

// GetWorkspaceNaming returns the workspace naming pattern.
func (c *Config) GetWorkspaceNaming() string {
	return c.WorkspaceNaming
}

// GetStaleThresholdDays returns the stale threshold in days.
func (c *Config) GetStaleThresholdDays() int {
	return c.StaleThresholdDays
}

// GetParallelWorkers returns the number of parallel workers for repository operations.
func (c *Config) GetParallelWorkers() int {
	return c.ParallelWorkers
}

// GetLockTimeout returns the workspace lock timeout.
func (c *Config) GetLockTimeout() time.Duration {
	if c.LockTimeout == "" {
		return DefaultLockTimeout
	}

	parsed, err := time.ParseDuration(c.LockTimeout)
	if err != nil {
		return DefaultLockTimeout
	}

	return parsed
}

// GetLockStaleThreshold returns the stale lock threshold.
func (c *Config) GetLockStaleThreshold() time.Duration {
	if c.LockStaleThreshold == "" {
		return DefaultLockStaleThreshold
	}

	parsed, err := time.ParseDuration(c.LockStaleThreshold)
	if err != nil {
		return DefaultLockStaleThreshold
	}

	return parsed
}

// GetRegistry returns the repository registry.
func (c *Config) GetRegistry() ports.RepoRegistry {
	if c.Registry == nil {
		return nil
	}

	return c.Registry
}

// GetHooks returns the lifecycle hooks configuration.
func (c *Config) GetHooks() ports.HooksConfig {
	return toPortHooksConfig(c.Hooks)
}

// GetKeybindings returns the TUI keybindings with defaults applied.
func (c *Config) GetKeybindings() ports.Keybindings {
	return toPortKeybindings(c.TUI.Keybindings.WithDefaults())
}

// GetUseEmoji returns whether emoji should be used in the TUI.
func (c *Config) GetUseEmoji() bool {
	return c.TUI.GetUseEmoji()
}

// GetGitRetryConfig returns the parsed git retry configuration.
func (c *Config) GetGitRetryConfig() ports.GitRetryConfig {
	parsed, err := c.Git.Retry.Parse()
	if err != nil {
		return defaultPortGitRetryConfig
	}

	return ports.GitRetryConfig{
		MaxAttempts:  parsed.MaxAttempts,
		InitialDelay: parsed.InitialDelay,
		MaxDelay:     parsed.MaxDelay,
		Multiplier:   parsed.Multiplier,
		JitterFactor: parsed.JitterFactor,
	}
}

func toPortHooksConfig(h Hooks) ports.HooksConfig {
	return ports.HooksConfig{
		PostCreate: toPortHooks(h.PostCreate),
		PreClose:   toPortHooks(h.PreClose),
	}
}

func toPortHooks(hooks []Hook) []ports.HookSpec {
	if hooks == nil {
		return nil
	}

	result := make([]ports.HookSpec, 0, len(hooks))
	for _, hook := range hooks {
		result = append(result, ports.HookSpec{
			Command:         hook.Command,
			Description:     hook.Description,
			Repos:           copyKeys(hook.Repos),
			Shell:           hook.Shell,
			Timeout:         hook.Timeout,
			ContinueOnError: hook.ContinueOnError,
		})
	}

	return result
}

func toPortKeybindings(k Keybindings) ports.Keybindings {
	return ports.Keybindings{
		Quit:        copyKeys(k.Quit),
		Search:      copyKeys(k.Search),
		Sync:        copyKeys(k.Sync),
		Push:        copyKeys(k.Push),
		Close:       copyKeys(k.Close),
		OpenEditor:  copyKeys(k.OpenEditor),
		ToggleStale: copyKeys(k.ToggleStale),
		Details:     copyKeys(k.Details),
		Select:      copyKeys(k.Select),
		SelectAll:   copyKeys(k.SelectAll),
		DeselectAll: copyKeys(k.DeselectAll),
		Confirm:     copyKeys(k.Confirm),
		Cancel:      copyKeys(k.Cancel),
	}
}

func toPortTemplate(name string, tmpl Template) ports.WorkspaceTemplate {
	return ports.WorkspaceTemplate{
		Name:          name,
		Repos:         copyKeys(tmpl.Repos),
		DefaultBranch: tmpl.DefaultBranch,
		Description:   tmpl.Description,
		SetupCommands: copyKeys(tmpl.SetupCommands),
	}
}
