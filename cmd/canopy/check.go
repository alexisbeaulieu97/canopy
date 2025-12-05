package main

import (
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate the current configuration",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		cfg := app.Config

		app.Logger.Info("Configuration loaded successfully.")
		app.Logger.Infof("Projects Root: %s", cfg.GetProjectsRoot())
		app.Logger.Infof("Workspaces Root: %s", cfg.GetWorkspacesRoot())
		app.Logger.Infof("Naming Pattern: %s", cfg.GetWorkspaceNaming())
		if registry := cfg.GetRegistry(); registry != nil {
			app.Logger.Infof("Registry File: %s", registry.Path())
		}

		if err := cfg.Validate(); err != nil {
			app.Logger.Errorf("Configuration is invalid: %v", err)
			return cerrors.Wrap(cerrors.ErrConfigInvalid, "configuration is invalid", err)
		}

		app.Logger.Info("Configuration is valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
