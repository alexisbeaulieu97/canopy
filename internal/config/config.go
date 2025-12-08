package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// Hook defines a single lifecycle hook command.
type Hook struct {
	Command         string   `mapstructure:"command"`
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

// Config holds the global configuration
type Config struct {
	ProjectsRoot       string        `mapstructure:"projects_root"`
	WorkspacesRoot     string        `mapstructure:"workspaces_root"`
	ClosedRoot         string        `mapstructure:"closed_root"`
	CloseDefault       string        `mapstructure:"workspace_close_default"`
	WorkspaceNaming    string        `mapstructure:"workspace_naming"`
	StaleThresholdDays int           `mapstructure:"stale_threshold_days"`
	Defaults           Defaults      `mapstructure:"defaults"`
	Hooks              Hooks         `mapstructure:"hooks"`
	Registry           *RepoRegistry `mapstructure:"-"`
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

// Load initializes and loads the configuration
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(home, ".canopy"))
	viper.AddConfigPath(filepath.Join(home, ".config", "canopy"))

	viper.SetDefault("projects_root", filepath.Join(home, ".canopy", "projects"))
	viper.SetDefault("workspaces_root", filepath.Join(home, ".canopy", "workspaces"))
	viper.SetDefault("closed_root", filepath.Join(home, ".canopy", "closed"))
	viper.SetDefault("workspace_close_default", "delete")
	viper.SetDefault("workspace_naming", "{{.ID}}")
	viper.SetDefault("stale_threshold_days", 14)

	viper.SetEnvPrefix("CANOPY")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is okay, use defaults
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand tilde
	cfg.ProjectsRoot = expandPath(cfg.ProjectsRoot, home)
	cfg.WorkspacesRoot = expandPath(cfg.WorkspacesRoot, home)
	cfg.ClosedRoot = expandPath(cfg.ClosedRoot, home)
	cfg.CloseDefault = strings.ToLower(cfg.CloseDefault)

	registry, err := LoadRepoRegistry("")
	if err != nil {
		return nil, fmt.Errorf("failed to load repository registry: %w", err)
	}

	cfg.Registry = registry

	return &cfg, nil
}

func expandPath(path, home string) string {
	if path == "~" {
		return home
	}

	if len(path) > 1 && path[:2] == "~/" {
		return filepath.Join(home, path[2:])
	}

	return path
}

// GetReposForWorkspace returns default repos for a given workspace ID based on patterns
func (c *Config) GetReposForWorkspace(workspaceID string) []string {
	for _, p := range c.Defaults.WorkspacePatterns {
		matched, err := regexp.MatchString(p.Pattern, workspaceID)
		if err == nil && matched {
			return p.Repos
		}
	}

	return nil
}

// Validate performs complete configuration validation by first checking values
// (pure validation) and then verifying the environment (filesystem checks).
// This is the main validation entry point that maintains backward compatibility.
func (c *Config) Validate() error {
	if err := c.ValidateValues(); err != nil {
		return err
	}

	return c.ValidateEnvironment()
}

// ValidateValues performs pure validation of configuration values without any
// filesystem or I/O operations. This includes checking required fields, enum
// values, regex patterns, and numeric constraints. Use this method when you
// need fast validation that doesn't depend on the environment.
func (c *Config) ValidateValues() error {
	if err := c.validateRequiredFields(); err != nil {
		return err
	}

	if err := c.validateCloseDefault(); err != nil {
		return err
	}

	if err := c.validatePatterns(); err != nil {
		return err
	}

	if err := c.validateStaleThreshold(); err != nil {
		return err
	}

	return c.validateHooks()
}

// validateRequiredFields checks that all required configuration fields are set.
func (c *Config) validateRequiredFields() error {
	if err := validateRequiredField("projects_root", c.ProjectsRoot); err != nil {
		return err
	}

	if err := validateRequiredField("workspaces_root", c.WorkspacesRoot); err != nil {
		return err
	}

	return validateRequiredField("closed_root", c.ClosedRoot)
}

// validateCloseDefault validates and applies default for the close behavior.
func (c *Config) validateCloseDefault() error {
	if c.CloseDefault == "" {
		c.CloseDefault = "delete"
	}

	if c.CloseDefault != "delete" && c.CloseDefault != "archive" {
		return fmt.Errorf("workspace_close_default must be either 'delete' or 'archive', got %q", c.CloseDefault)
	}

	return nil
}

// validatePatterns checks that all workspace regex patterns are valid.
func (c *Config) validatePatterns() error {
	for _, p := range c.Defaults.WorkspacePatterns {
		if _, err := regexp.Compile(p.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", p.Pattern, err)
		}
	}

	return nil
}

// validateStaleThreshold checks that the stale threshold is non-negative.
func (c *Config) validateStaleThreshold() error {
	if c.StaleThresholdDays < 0 {
		return fmt.Errorf("stale_threshold_days must be zero or positive, got %d", c.StaleThresholdDays)
	}

	return nil
}

// validateHooks validates all hook configurations.
func (c *Config) validateHooks() error {
	for i, h := range c.Hooks.PostCreate {
		if err := validateHook(h, "post_create", i); err != nil {
			return err
		}
	}

	for i, h := range c.Hooks.PreClose {
		if err := validateHook(h, "pre_close", i); err != nil {
			return err
		}
	}

	return nil
}

// validateHook checks that a hook has valid configuration.
func validateHook(h Hook, hookType string, index int) error {
	if h.Command == "" {
		return fmt.Errorf("%s hook[%d] command cannot be empty", hookType, index)
	}

	if h.Timeout < 0 {
		return fmt.Errorf("%s hook[%d] timeout must be non-negative, got %d", hookType, index, h.Timeout)
	}

	return nil
}

// ValidateEnvironment verifies that the configuration's filesystem paths exist
// and are directories. This method performs I/O operations and should be called
// after ValidateValues() when you need to ensure the environment is ready.
func (c *Config) ValidateEnvironment() error {
	if err := validateRootPath("projects_root", c.ProjectsRoot); err != nil {
		return err
	}

	if err := validateRootPath("workspaces_root", c.WorkspacesRoot); err != nil {
		return err
	}

	if err := validateRootPath("closed_root", c.ClosedRoot); err != nil {
		return err
	}

	return nil
}

// validateRequiredField checks that a field value is non-empty.
func validateRequiredField(label, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", label)
	}

	return nil
}

// validateRootPath checks that a path exists and is a directory.
// If the path doesn't exist but is absolute, it's considered valid
// (it will be created later). Non-absolute paths that don't exist are invalid.
func validateRootPath(label, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if !filepath.IsAbs(path) {
				return fmt.Errorf("%s must be an absolute path: %s", label, path)
			}

			return nil
		}

		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory: %s", label, path)
	}

	return nil
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

// GetRegistry returns the repository registry.
func (c *Config) GetRegistry() *RepoRegistry {
	return c.Registry
}

// GetHooks returns the lifecycle hooks configuration.
func (c *Config) GetHooks() Hooks {
	return c.Hooks
}
