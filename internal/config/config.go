package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the global configuration
type Config struct {
	ProjectsRoot       string        `mapstructure:"projects_root"`
	WorkspacesRoot     string        `mapstructure:"workspaces_root"`
	ClosedRoot         string        `mapstructure:"closed_root"`
	CloseDefault       string        `mapstructure:"workspace_close_default"`
	WorkspaceNaming    string        `mapstructure:"workspace_naming"`
	StaleThresholdDays int           `mapstructure:"stale_threshold_days"`
	Defaults           Defaults      `mapstructure:"defaults"`
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
	viper.SetDefault("closed_root", filepath.Join(home, ".canopy", "archives"))
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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if err := validateRoot("projects_root", c.ProjectsRoot); err != nil {
		return err
	}

	if err := validateRoot("workspaces_root", c.WorkspacesRoot); err != nil {
		return err
	}

	if err := validateRoot("closed_root", c.ClosedRoot); err != nil {
		return err
	}

	if c.CloseDefault == "" {
		c.CloseDefault = "delete"
	}

	if c.CloseDefault != "delete" && c.CloseDefault != "archive" {
		return fmt.Errorf("workspace_close_default must be either 'delete' or 'archive', got %q", c.CloseDefault)
	}

	// Check Patterns
	for _, p := range c.Defaults.WorkspacePatterns {
		if _, err := regexp.Compile(p.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", p.Pattern, err)
		}
	}

	if c.StaleThresholdDays < 0 {
		return fmt.Errorf("stale_threshold_days must be zero or positive, got %d", c.StaleThresholdDays)
	}

	return nil
}

func validateRoot(label, path string) error {
	if path == "" {
		return fmt.Errorf("%s is required", label)
	}

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
