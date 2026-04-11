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

	v := viper.New()
	v.SetConfigType("yaml")

	explicitConfigPath := false
	if configPath != "" {
		v.SetConfigFile(expandPath(configPath, home))
		explicitConfigPath = true
	} else if envPath := os.Getenv("CANOPY_CONFIG"); envPath != "" {
		v.SetConfigFile(expandPath(envPath, home))
		explicitConfigPath = true
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Join(home, ".canopy"))
		v.AddConfigPath(filepath.Join(home, ".config", "canopy"))
	}

	v.SetDefault("projects_root", filepath.Join(home, ".canopy", "projects"))
	v.SetDefault("workspaces_root", filepath.Join(home, ".canopy", "workspaces"))
	v.SetDefault("closed_root", filepath.Join(home, ".canopy", "closed"))
	v.SetDefault("workspace_close_default", CloseDefaultDelete)
	v.SetDefault("workspace_naming", "{{.ID}}")
	v.SetDefault("stale_threshold_days", 14)
	v.SetDefault("lock_timeout", DefaultLockTimeout.String())
	v.SetDefault("lock_stale_threshold", DefaultLockStaleThreshold.String())
	v.SetDefault("parallel_workers", DefaultParallelWorkers)
	v.SetDefault("git.retry.max_attempts", 3)
	v.SetDefault("git.retry.initial_delay", "1s")
	v.SetDefault("git.retry.max_delay", "30s")
	v.SetDefault("git.retry.multiplier", 2.0)
	v.SetDefault("git.retry.jitter_factor", 0.25)

	v.SetEnvPrefix("CANOPY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if explicitConfigPath {
				return nil, cerrors.NewIOFailed("read config file", fmt.Errorf("config file not found: %s", v.ConfigFileUsed()))
			}
		} else {
			return nil, cerrors.NewIOFailed("read config file", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, func(config *mapstructure.DecoderConfig) {
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
