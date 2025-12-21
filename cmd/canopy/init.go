package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	RunE: func(_ *cobra.Command, _ []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		configDir := filepath.Join(home, ".canopy")
		if err := os.MkdirAll(configDir, 0o750); err != nil {
			return err
		}

		configFile := filepath.Join(configDir, "config.yaml")
		if _, err := os.Stat(configFile); err == nil {
			fmt.Println("Config file already exists:", configFile) //nolint:forbidigo // user-facing CLI output
			return nil
		}

		f, err := os.Create(configFile) //nolint:gosec // path is within user config directory
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		// Write defaults
		_, err = f.WriteString(`projects_root: ~/.canopy/projects
workspaces_root: ~/.canopy/workspaces
workspace_naming: "{{.ID}}"

templates:
  backend:
    description: "Backend workspace defaults"
    repos: ["backend", "common"]
`)
		if err != nil {
			return err
		}

		fmt.Println("Initialized config at:", configFile) //nolint:forbidigo // user-facing CLI output
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
