package main

import (
	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
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
		jsonOutput, _ := cmd.Flags().GetBool("json")

		validationErr := cfg.Validate()

		if jsonOutput {
			configInfo := map[string]interface{}{
				"projects_root":    cfg.GetProjectsRoot(),
				"workspaces_root":  cfg.GetWorkspacesRoot(),
				"workspace_naming": cfg.GetWorkspaceNaming(),
				"valid":            validationErr == nil,
			}

			if registry := cfg.GetRegistry(); registry != nil {
				configInfo["registry_path"] = registry.Path()
			}

			if validationErr != nil {
				configInfo["validation_error"] = validationErr.Error()
			}

			return output.PrintJSON(configInfo)
		}

		app.Logger.Info("Configuration loaded successfully.")
		app.Logger.Infof("Projects Root: %s", cfg.GetProjectsRoot())
		app.Logger.Infof("Workspaces Root: %s", cfg.GetWorkspacesRoot())
		app.Logger.Infof("Naming Pattern: %s", cfg.GetWorkspaceNaming())
		if registry := cfg.GetRegistry(); registry != nil {
			app.Logger.Infof("Registry File: %s", registry.Path())
		}

		if validationErr != nil {
			app.Logger.Errorf("Configuration is invalid: %v", validationErr)
			return cerrors.Wrap(cerrors.ErrConfigInvalid, "configuration is invalid", validationErr)
		}

		app.Logger.Info("Configuration is valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("json", false, "Output in JSON format")
}
