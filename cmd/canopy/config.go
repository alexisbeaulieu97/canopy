package main

import (
	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validate the configuration file without running other commands.

This command loads and validates the config file, checking for:
- Unknown or misspelled config fields
- Invalid values (e.g., negative timeouts, invalid regex patterns)
- Missing required fields

Exit codes:
  0 - Configuration is valid
  1 - Configuration has errors`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		validateConfigPath, _ := cmd.Flags().GetString("config")

		// Use the --config flag from this command if specified,
		// otherwise fall back to the root command's --config flag
		if validateConfigPath == "" {
			validateConfigPath = configPath
		}

		// Load config (this performs strict validation during unmarshal)
		cfg, err := config.Load(validateConfigPath)
		if err != nil {
			if jsonOutput {
				_ = output.PrintErrorJSON(err)
			}
			return err
		}

		// Run value validation (checks required fields, patterns, etc.)
		if err := cfg.ValidateValues(); err != nil {
			if jsonOutput {
				_ = output.PrintErrorJSON(err)
			}
			return cerrors.Wrap(cerrors.ErrConfigValidation, "configuration validation failed", err)
		}

		// Run environment validation (checks filesystem paths)
		if err := cfg.ValidateEnvironment(); err != nil {
			if jsonOutput {
				_ = output.PrintErrorJSON(err)
			}
			return cerrors.Wrap(cerrors.ErrConfigValidation, "configuration validation failed", err)
		}

		if jsonOutput {
			configInfo := map[string]interface{}{
				"valid":            true,
				"projects_root":    cfg.GetProjectsRoot(),
				"workspaces_root":  cfg.GetWorkspacesRoot(),
				"closed_root":      cfg.GetClosedRoot(),
				"workspace_naming": cfg.GetWorkspaceNaming(),
			}
			if registry := cfg.GetRegistry(); registry != nil {
				configInfo["registry_path"] = registry.Path()
			}
			if cfg.HasWarnings() {
				configInfo["warnings"] = cfg.GetWarnings()
			}
			return output.PrintJSON(configInfo)
		}

		// Display any warnings first
		if cfg.HasWarnings() {
			output.Warn("Configuration warnings:")
			for _, warning := range cfg.GetWarnings() {
				output.Warnf("  • %s", warning)
			}
			output.Info("") // Empty line for spacing
		}

		output.Info("Configuration is valid.")
		output.Infof("  Projects root:   %s", cfg.GetProjectsRoot())
		output.Infof("  Workspaces root: %s", cfg.GetWorkspacesRoot())
		output.Infof("  Closed root:     %s", cfg.GetClosedRoot())
		output.Infof("  Workspace naming: %s", cfg.GetWorkspaceNaming())
		if registry := cfg.GetRegistry(); registry != nil {
			output.Infof("  Registry file:   %s", registry.Path())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configValidateCmd)

	configValidateCmd.Flags().Bool("json", false, "Output in JSON format")
	configValidateCmd.Flags().String("config", "", "Path to config file to validate (overrides --config from root command)")
}
