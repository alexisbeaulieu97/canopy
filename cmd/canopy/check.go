package main

import (
	"fmt"

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
		if cfg.GetRegistry() != nil {
			app.Logger.Infof("Registry File: %s", cfg.GetRegistry().Path())
		}

		if err := cfg.Validate(); err != nil {
			app.Logger.Errorf("Configuration is invalid: %v", err)
			return fmt.Errorf("configuration is invalid: %w", err)
		}

		app.Logger.Info("Configuration is valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
