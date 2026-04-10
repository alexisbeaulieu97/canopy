package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// Load initializes and loads the configuration.
func Load(configPath string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, cerrors.NewIOFailed("get user home dir", err)
	}

	viper.SetConfigType("yaml")

	explicitConfigPath := false
	if configPath != "" {
		viper.SetConfigFile(expandPath(configPath, home))
		explicitConfigPath = true
	} else if envPath := os.Getenv("CANOPY_CONFIG"); envPath != "" {
		viper.SetConfigFile(expandPath(envPath, home))
		explicitConfigPath = true
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".canopy"))
		viper.AddConfigPath(filepath.Join(home, ".config", "canopy"))
	}

	viper.SetDefault("projects_root", filepath.Join(home, ".canopy", "projects"))
	viper.SetDefault("workspaces_root", filepath.Join(home, ".canopy", "workspaces"))
	viper.SetDefault("closed_root", filepath.Join(home, ".canopy", "closed"))
	viper.SetDefault("workspace_close_default", CloseDefaultDelete)
	viper.SetDefault("workspace_naming", "{{.ID}}")
	viper.SetDefault("stale_threshold_days", 14)
	viper.SetDefault("lock_timeout", DefaultLockTimeout.String())
	viper.SetDefault("lock_stale_threshold", DefaultLockStaleThreshold.String())
	viper.SetDefault("parallel_workers", DefaultParallelWorkers)
	viper.SetDefault("git.retry.max_attempts", 3)
	viper.SetDefault("git.retry.initial_delay", "1s")
	viper.SetDefault("git.retry.max_delay", "30s")
	viper.SetDefault("git.retry.multiplier", 2.0)
	viper.SetDefault("git.retry.jitter_factor", 0.25)

	viper.SetEnvPrefix("CANOPY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if explicitConfigPath {
				return nil, cerrors.NewIOFailed("read config file", fmt.Errorf("config file not found: %s", viper.ConfigFileUsed()))
			}
		} else {
			return nil, cerrors.NewIOFailed("read config file", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg, func(config *mapstructure.DecoderConfig) {
		config.ErrorUnused = true
	}); err != nil {
		return nil, handleUnmarshalError(err)
	}

	cfg.ProjectsRoot = expandPath(cfg.ProjectsRoot, home)
	cfg.WorkspacesRoot = expandPath(cfg.WorkspacesRoot, home)
	cfg.ClosedRoot = expandPath(cfg.ClosedRoot, home)
	cfg.CloseDefault = strings.ToLower(cfg.CloseDefault)

	registry, err := LoadRepoRegistry("")
	if err != nil {
		return nil, cerrors.NewRegistryError("load", "repository registry", err)
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
