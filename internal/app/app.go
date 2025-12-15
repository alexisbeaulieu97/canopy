// Package app provides the shared application container for CLI commands.
package app

import (
	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/gitx"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
	"github.com/alexisbeaulieu97/canopy/internal/workspace"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// App holds shared services for CLI commands.
type App struct {
	Config  ports.ConfigProvider
	Logger  *logging.Logger
	Service *workspaces.Service
}

// Option is a functional option for configuring the App.
type Option func(*appOptions)

// appOptions holds optional dependencies that can be injected.
type appOptions struct {
	gitOps     ports.GitOperations
	wsStorage  ports.WorkspaceStorage
	configPrv  ports.ConfigProvider
	configPath string
	logger     *logging.Logger
}

// WithGitOperations sets a custom GitOperations implementation.
func WithGitOperations(g ports.GitOperations) Option {
	return func(o *appOptions) {
		o.gitOps = g
	}
}

// WithWorkspaceStorage sets a custom WorkspaceStorage implementation.
func WithWorkspaceStorage(s ports.WorkspaceStorage) Option {
	return func(o *appOptions) {
		o.wsStorage = s
	}
}

// WithConfigProvider sets a custom ConfigProvider implementation.
func WithConfigProvider(c ports.ConfigProvider) Option {
	return func(o *appOptions) {
		o.configPrv = c
	}
}

// WithConfigPath sets a custom path for loading configuration.
// This takes precedence over CANOPY_CONFIG environment variable and default locations.
func WithConfigPath(path string) Option {
	return func(o *appOptions) {
		o.configPath = path
	}
}

// WithLogger sets a custom Logger instance.
func WithLogger(l *logging.Logger) Option {
	return func(o *appOptions) {
		o.logger = l
	}
}

// New creates a new App instance with initialized dependencies.
// Options can be provided to override default implementations for testing.
func New(debug bool, opts ...Option) (*App, error) {
	// Apply all options
	options := &appOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Use provided config or load from disk
	var cfg ports.ConfigProvider
	if options.configPrv != nil {
		cfg = options.configPrv
	} else {
		loadedCfg, err := config.Load(options.configPath)
		if err != nil {
			return nil, err
		}

		if err := loadedCfg.Validate(); err != nil {
			return nil, err
		}

		cfg = loadedCfg
	}

	// Use provided logger or create new one
	logger := options.logger
	if logger == nil {
		logger = logging.New(debug)
	}

	// Use provided git operations or create default
	gitEngine := options.gitOps
	if gitEngine == nil {
		retryCfg := cfg.GetGitRetryConfig()
		gitEngine = gitx.NewWithRetry(cfg.GetProjectsRoot(), gitx.RetryConfig{
			MaxAttempts:  retryCfg.MaxAttempts,
			InitialDelay: retryCfg.InitialDelay,
			MaxDelay:     retryCfg.MaxDelay,
			Multiplier:   retryCfg.Multiplier,
			JitterFactor: retryCfg.JitterFactor,
		})
	}

	// Use provided workspace storage or create default
	wsEngine := options.wsStorage
	if wsEngine == nil {
		wsEngine = workspace.New(cfg.GetWorkspacesRoot(), cfg.GetClosedRoot())
	}

	return &App{
		Config:  cfg,
		Logger:  logger,
		Service: workspaces.NewService(cfg, gitEngine, wsEngine, logger),
	}, nil
}

// Shutdown is a placeholder for cleaning up resources when needed.
func (a *App) Shutdown() error {
	return nil
}
