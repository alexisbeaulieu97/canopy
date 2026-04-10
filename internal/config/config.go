// Package config provides configuration loading and management for Canopy.
//
// # Configuration Loading Priority
//
// Configuration is loaded with the following priority (highest to lowest):
//  1. Explicit --config flag path
//  2. CANOPY_CONFIG environment variable
//  3. Default search paths (in order):
//     - ./config.yaml (current directory)
//     - ~/.canopy/config.yaml
//     - ~/.config/canopy/config.yaml
//
// When an explicit config path is provided via --config flag or CANOPY_CONFIG
// environment variable, the file must exist or loading will fail. Default search
// paths are optional - if no config file is found, defaults are used.
//
// Paths support tilde (~) expansion to the user's home directory.
//
// Environment variables with the CANOPY_ prefix can override configuration values.
//
// # Configuration Options
//
// Key configuration options include:
//   - projects_root: Directory for canonical bare repositories
//   - workspaces_root: Directory for active workspaces
//   - closed_root: Directory for archived workspace metadata
//   - workspace_close_default: Default close behavior ("delete" or "archive")
//   - stale_threshold_days: Days before a workspace is considered stale
//
// # Workspace Patterns
//
// Workspace patterns allow automatic repository assignment based on workspace ID:
//
//	defaults:
//	  workspace_patterns:
//	    - pattern: "^PROJ-"
//	      repos: ["backend", "frontend"]
//
// # Lifecycle Hooks
//
// Hooks execute commands at workspace lifecycle events:
//
//	hooks:
//	  post_create:
//	    - command: "npm install"
//	      repos: ["frontend"]
//	  pre_close:
//	    - command: "git stash"
//
// See the configuration documentation for complete reference.
package config

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// Hook defines a single lifecycle hook command.
type Hook struct {
	Command         string   `mapstructure:"command"`
	Description     string   `mapstructure:"description,omitempty"`       // human-readable description
	Repos           []string `mapstructure:"repos,omitempty"`             // filter to specific repos
	Shell           string   `mapstructure:"shell,omitempty"`             // default: sh -c
	Timeout         int      `mapstructure:"timeout,omitempty"`           // default: 30 seconds
	ContinueOnError bool     `mapstructure:"continue_on_error,omitempty"` // don't fail workspace operation
}

// Hooks holds lifecycle hook configurations.
type Hooks struct {
	PostCreate []Hook `mapstructure:"post_create"`
	PreClose   []Hook `mapstructure:"pre_close"`
}

// Keybindings holds TUI keybinding configurations.
type Keybindings struct {
	Quit        []string `mapstructure:"quit"`
	Search      []string `mapstructure:"search"`
	Sync        []string `mapstructure:"sync"`
	Push        []string `mapstructure:"push"`
	Close       []string `mapstructure:"close"`
	OpenEditor  []string `mapstructure:"open_editor"`
	ToggleStale []string `mapstructure:"toggle_stale"`
	Details     []string `mapstructure:"details"`
	Select      []string `mapstructure:"select"`
	SelectAll   []string `mapstructure:"select_all"`
	DeselectAll []string `mapstructure:"deselect_all"`
	Confirm     []string `mapstructure:"confirm"`
	Cancel      []string `mapstructure:"cancel"`
}

// TUIConfig holds TUI-specific configuration.
type TUIConfig struct {
	Keybindings Keybindings `mapstructure:"keybindings"`
	UseEmoji    *bool       `mapstructure:"use_emoji"` // nil means default (true)
}

// GetUseEmoji returns whether emoji should be used in the TUI.
// Defaults to true for backward compatibility.
func (t TUIConfig) GetUseEmoji() bool {
	if t.UseEmoji == nil {
		return true
	}

	return *t.UseEmoji
}

// GitRetrySettings holds YAML configuration for git network operation retry behavior.
// This is the config-file representation; use ParsedRetryConfig for runtime use.
type GitRetrySettings struct {
	MaxAttempts  int     `mapstructure:"max_attempts"`
	InitialDelay string  `mapstructure:"initial_delay"` // Duration string, e.g. "1s"
	MaxDelay     string  `mapstructure:"max_delay"`     // Duration string, e.g. "30s"
	Multiplier   float64 `mapstructure:"multiplier"`
	JitterFactor float64 `mapstructure:"jitter_factor"`
}

// ParsedRetryConfig holds the parsed retry configuration with proper Go types.
type ParsedRetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}

// Parse converts the string-based GitRetrySettings to ParsedRetryConfig with proper duration types.
func (r GitRetrySettings) Parse() (ParsedRetryConfig, error) {
	initialDelay, err := time.ParseDuration(r.InitialDelay)
	if err != nil {
		return ParsedRetryConfig{}, cerrors.NewConfigValidation("git.retry.initial_delay", fmt.Sprintf("invalid duration %q: %v", r.InitialDelay, err))
	}

	maxDelay, err := time.ParseDuration(r.MaxDelay)
	if err != nil {
		return ParsedRetryConfig{}, cerrors.NewConfigValidation("git.retry.max_delay", fmt.Sprintf("invalid duration %q: %v", r.MaxDelay, err))
	}

	return ParsedRetryConfig{
		MaxAttempts:  r.MaxAttempts,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   r.Multiplier,
		JitterFactor: r.JitterFactor,
	}, nil
}

// GitConfig holds git-related configuration.
type GitConfig struct {
	Retry GitRetrySettings `mapstructure:"retry"`
}

// Config holds the global configuration
type Config struct {
	ProjectsRoot       string              `mapstructure:"projects_root"`
	WorkspacesRoot     string              `mapstructure:"workspaces_root"`
	ClosedRoot         string              `mapstructure:"closed_root"`
	CloseDefault       string              `mapstructure:"workspace_close_default"`
	WorkspaceNaming    string              `mapstructure:"workspace_naming"`
	StaleThresholdDays int                 `mapstructure:"stale_threshold_days"`
	ParallelWorkers    int                 `mapstructure:"parallel_workers"`
	LockTimeout        string              `mapstructure:"lock_timeout"`
	LockStaleThreshold string              `mapstructure:"lock_stale_threshold"`
	Defaults           Defaults            `mapstructure:"defaults"`
	Templates          map[string]Template `mapstructure:"templates"`
	Hooks              Hooks               `mapstructure:"hooks"`
	TUI                TUIConfig           `mapstructure:"tui"`
	Git                GitConfig           `mapstructure:"git"`
	Registry           *RepoRegistry       `mapstructure:"-"`
}

// WorkspaceNamingTemplateData defines the data available to workspace naming templates.
type WorkspaceNamingTemplateData struct {
	ID string
}

// WorkspacePattern defines a regex pattern and default repos
type WorkspacePattern struct {
	Pattern string   `mapstructure:"pattern"`
	Repos   []string `mapstructure:"repos"`
}

// Defaults holds default configurations
type Defaults struct {
	WorkspacePatterns []WorkspacePattern `mapstructure:"workspace_patterns"`
}

// Template defines reusable workspace defaults.
type Template struct {
	Name          string   `mapstructure:"-"`
	Repos         []string `mapstructure:"repos"`
	DefaultBranch string   `mapstructure:"default_branch"`
	Description   string   `mapstructure:"description"`
	SetupCommands []string `mapstructure:"setup_commands"`
}

// ComputeWorkspaceDir computes the workspace directory name for a given ID.
func (c *Config) ComputeWorkspaceDir(id string) (string, error) {
	if err := validation.ValidateWorkspaceID(id); err != nil {
		return "", err
	}

	return c.computeWorkspaceDir(id)
}

func (c *Config) computeWorkspaceDir(id string) (string, error) {
	if strings.TrimSpace(c.WorkspaceNaming) == "" {
		c.WorkspaceNaming = "{{.ID}}"
	}

	tmpl, err := template.New("workspace_naming").Option("missingkey=error").Parse(c.WorkspaceNaming)
	if err != nil {
		return "", cerrors.NewConfigValidation("workspace_naming", fmt.Sprintf("invalid template: %v", err))
	}

	var rendered strings.Builder
	if err := tmpl.Execute(&rendered, WorkspaceNamingTemplateData{ID: id}); err != nil {
		return "", cerrors.NewConfigValidation("workspace_naming", fmt.Sprintf("template execution failed: %v", err))
	}

	rawDir := rendered.String()

	dirName, err := validation.NormalizeWorkspaceDirName(rawDir)
	if err != nil {
		return "", cerrors.NewConfigValidation("workspace_naming", fmt.Sprintf("template output %q is invalid: %v", rawDir, err))
	}

	return dirName, nil
}
