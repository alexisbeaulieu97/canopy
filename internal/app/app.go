package app

import (
	"github.com/alexisbeaulieu97/yard/internal/config"
	"github.com/alexisbeaulieu97/yard/internal/gitx"
	"github.com/alexisbeaulieu97/yard/internal/logging"
	"github.com/alexisbeaulieu97/yard/internal/workspace"
	"github.com/alexisbeaulieu97/yard/internal/workspaces"
)

// App holds shared services for CLI commands.
type App struct {
	Config  *config.Config
	Logger  *logging.Logger
	Service *workspaces.Service
}

// New creates a new App instance with initialized dependencies.
func New(debug bool) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := logging.New(debug)
	gitEngine := gitx.New(cfg.ProjectsRoot)
	wsEngine := workspace.New(cfg.WorkspacesRoot)

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
