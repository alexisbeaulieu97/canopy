package main

import (
	"os"

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
		app.Logger.Infof("Projects Root: %s", cfg.ProjectsRoot)
		app.Logger.Infof("Workspaces Root: %s", cfg.WorkspacesRoot)
		app.Logger.Infof("Naming Pattern: %s", cfg.WorkspaceNaming)
		if cfg.Registry != nil {
			app.Logger.Infof("Registry File: %s", cfg.Registry.Path())
		}

		if err := cfg.Validate(); err != nil {
			app.Logger.Errorf("Configuration is invalid: %v", err)
			os.Exit(1)
		}

		app.Logger.Info("Configuration is valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
