// Package app provides the shared application container for CLI commands.
package app

import (
	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/gitx"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/workspace"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
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
	wsEngine := workspace.New(cfg.WorkspacesRoot, cfg.ArchivesRoot)

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
