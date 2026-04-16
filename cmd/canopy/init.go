package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const starterConfig = `projects_root: ~/.canopy/projects
workspaces_root: ~/.canopy/workspaces
closed_root: ~/.canopy/closed
workspace_naming: "{{.ID}}"
parallel_workers: 4

defaults:
  workspace_patterns: []

templates: {}

# Example:
# defaults:
#   workspace_patterns:
#     - pattern: "^PROJ-"
#       repos: ["backend", "frontend"]
#
# templates:
#   web:
#     description: "Example template"
#     repos: ["frontend", "ui-kit"]
`

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

		_, err = f.WriteString(starterConfig)
		if err != nil {
			return err
		}

		fmt.Println("Initialized config at:", configFile)                                         //nolint:forbidigo // user-facing CLI output
		fmt.Println("")                                                                           //nolint:forbidigo // user-facing CLI output
		fmt.Println("Next steps:")                                                                //nolint:forbidigo // user-facing CLI output
		fmt.Println("  1. Add a repository: canopy repo add <repository-url>")                    //nolint:forbidigo // user-facing CLI output
		fmt.Println("  2. Or register an alias: canopy repo register <alias> <repository-url>")   //nolint:forbidigo // user-facing CLI output
		fmt.Println("  3. Create a workspace: canopy workspace new <id> --repos <alias1,alias2>") //nolint:forbidigo // user-facing CLI output

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
