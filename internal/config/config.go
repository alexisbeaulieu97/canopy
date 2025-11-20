package config

import (
	"fmt"
	"os"
	"path/filepath"

	"regexp"

	"github.com/spf13/viper"
)

// Config holds the global configuration
type Config struct {
	ProjectsRoot    string   `mapstructure:"projects_root"`
	WorkspacesRoot  string   `mapstructure:"workspaces_root"`
	WorkspaceNaming string   `mapstructure:"workspace_naming"`
	Defaults        Defaults `mapstructure:"defaults"`
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
	viper.AddConfigPath(filepath.Join(home, ".yard"))
	viper.AddConfigPath(filepath.Join(home, ".config", "yardmaster"))

	viper.SetDefault("projects_root", filepath.Join(home, ".yard", "projects"))
	viper.SetDefault("workspaces_root", filepath.Join(home, ".yard", "workspaces"))
	viper.SetDefault("workspace_naming", "{{.ID}}")

	viper.SetEnvPrefix("YARD")
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
	// Check ProjectsRoot
	if c.ProjectsRoot == "" {
		return fmt.Errorf("projects_root is required")
	}
	if info, err := os.Stat(c.ProjectsRoot); err != nil {
		if os.IsNotExist(err) {
			// It's okay if it doesn't exist, we might create it.
			// But for validation, maybe we warn?
			// Let's say it's valid but note it doesn't exist?
			// Actually, for 'check', we want to ensure it's usable.
			// But yard creates it on demand. So maybe just check if parent exists?
			// Let's just check if it's an absolute path for now.
			if !filepath.IsAbs(c.ProjectsRoot) {
				return fmt.Errorf("projects_root must be an absolute path: %s", c.ProjectsRoot)
			}
		} else {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("projects_root must be a directory: %s", c.ProjectsRoot)
	}

	// Check WorkspacesRoot
	if c.WorkspacesRoot == "" {
		return fmt.Errorf("workspaces_root is required")
	}
	if info, err := os.Stat(c.WorkspacesRoot); err != nil {
		if os.IsNotExist(err) {
			if !filepath.IsAbs(c.WorkspacesRoot) {
				return fmt.Errorf("workspaces_root must be an absolute path: %s", c.WorkspacesRoot)
			}
		} else {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("workspaces_root must be a directory: %s", c.WorkspacesRoot)
	}

	// Check Patterns
	for _, p := range c.Defaults.WorkspacePatterns {
		if _, err := regexp.Compile(p.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", p.Pattern, err)
		}
	}

	return nil
}
