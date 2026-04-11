// Package ports defines interfaces for external dependencies (hexagonal architecture).
package ports

import "time"

// RegistryEntry represents a repository alias entry in the runtime layer.
type RegistryEntry struct {
	Alias         string
	URL           string
	DefaultBranch string
	Description   string
	Tags          []string
}

// RepoRegistry defines the runtime contract for repository registry operations.
type RepoRegistry interface {
	Save() error
	Resolve(alias string) (RegistryEntry, bool)
	ResolveByURL(url string) (RegistryEntry, bool)
	Register(alias string, entry RegistryEntry, force bool) error
	RegisterWithSuffix(alias string, entry RegistryEntry) (string, error)
	Unregister(alias string) error
	List(tags []string) []RegistryEntry
	Path() string
}

// HookSpec represents a single lifecycle hook command in the runtime layer.
type HookSpec struct {
	Command         string
	Description     string
	Repos           []string
	Shell           string
	Timeout         int
	ContinueOnError bool
}

// HooksConfig groups configured lifecycle hooks.
type HooksConfig struct {
	PostCreate []HookSpec
	PreClose   []HookSpec
}

// Keybindings holds resolved TUI keybinding configuration.
type Keybindings struct {
	Quit        []string
	Search      []string
	Sync        []string
	Push        []string
	Close       []string
	OpenEditor  []string
	ToggleStale []string
	Details     []string
	Select      []string
	SelectAll   []string
	DeselectAll []string
	Confirm     []string
	Cancel      []string
}

// WorkspaceTemplate represents resolved workspace template settings.
type WorkspaceTemplate struct {
	Name          string
	Repos         []string
	DefaultBranch string
	Description   string
	SetupCommands []string
}

// GitRetryConfig is the runtime git retry configuration.
type GitRetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}
