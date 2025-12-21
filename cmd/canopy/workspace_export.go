package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// Export format constants.
const (
	formatJSON = "json"
	formatYAML = "yaml"
)

// workspace_export.go defines workspace export/import subcommands.

var workspaceExportCmd = &cobra.Command{
	Use:   "export <ID>",
	Short: "Export a workspace definition to a portable file",
	Long: `Export a workspace definition to YAML or JSON format.

The exported file contains the workspace ID, branch, and repository URLs,
allowing the workspace to be recreated on another machine.

Note: Only workspace metadata is exported. Local changes, uncommitted work,
and worktree state are NOT included. If repository URLs contain credentials,
avoid committing export files to version control.

Examples:
  canopy workspace export my-workspace
  canopy workspace export my-workspace --output ws.yaml
  canopy workspace export my-workspace --format json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		outputFile, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		// --json flag is shorthand for --format json.
		if jsonOutput {
			format = formatJSON
		}

		// Validate format.
		if format != formatYAML && format != formatJSON {
			return cerrors.NewInvalidArgument("format", "must be 'yaml' or 'json'")
		}

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		export, err := app.Service.ExportWorkspace(cmd.Context(), id)
		if err != nil {
			return err
		}

		var data []byte
		switch format {
		case formatJSON:
			data, err = json.MarshalIndent(export, "", "  ")
		default:
			data, err = yaml.Marshal(export)
		}
		if err != nil {
			return cerrors.NewInternalError("marshal export", err)
		}

		// Write to file or stdout.
		if outputFile != "" {
			if err := os.WriteFile(outputFile, data, 0o644); err != nil { //nolint:gosec // user-specified output file
				return cerrors.NewIOFailed("write export file", err)
			}
			output.Infof("Exported workspace %s to %s", id, outputFile)
		} else {
			output.Print(string(data))
		}

		return nil
	},
}

var workspaceImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a workspace from an exported definition",
	Long: `Import a workspace from a YAML or JSON export file.

The import command recreates a workspace from a previously exported definition,
cloning any missing repositories and creating worktrees.

Warning: When using --force to overwrite an existing workspace, the old workspace
is deleted before the new one is created. If the import fails (e.g., network issues
cloning repos), the original workspace cannot be recovered.

Examples:
  canopy workspace import ws.yaml
  canopy workspace import ws.yaml --id NEW-WORKSPACE
  canopy workspace import ws.yaml --branch develop
  canopy workspace import - < ws.yaml  # read from stdin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		idOverride, _ := cmd.Flags().GetString("id")
		branchOverride, _ := cmd.Flags().GetString("branch")
		force, _ := cmd.Flags().GetBool("force")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		// Read input from file or stdin.
		var data []byte
		if inputFile == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(inputFile) //nolint:gosec // user-specified input file
		}
		if err != nil {
			return cerrors.NewIOFailed("read import file", err)
		}

		// Parse as YAML (which also handles JSON).
		var export domain.WorkspaceExport
		if err := yaml.Unmarshal(data, &export); err != nil {
			return cerrors.NewInvalidArgument("file", fmt.Sprintf("invalid export format: %v", err))
		}

		// Validate export.
		if export.ID == "" && idOverride == "" {
			return cerrors.NewInvalidArgument("id", "export has no workspace ID and --id was not provided")
		}

		dirName, err := app.Service.ImportWorkspace(cmd.Context(), &export, idOverride, branchOverride, force)
		if err != nil {
			return err
		}

		workspaceID := export.ID
		if idOverride != "" {
			workspaceID = idOverride
		}

		output.SuccessWithPath("Imported workspace", workspaceID, app.Config.GetWorkspacesRoot()+"/"+dirName)
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceExportCmd)
	workspaceCmd.AddCommand(workspaceImportCmd)

	workspaceExportCmd.Flags().StringP("output", "o", "", "Write export to file instead of stdout")
	workspaceExportCmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json")
	workspaceExportCmd.Flags().Bool("json", false, "Output in JSON format (shorthand for --format json)")

	workspaceImportCmd.Flags().String("id", "", "Override workspace ID from export file")
	workspaceImportCmd.Flags().String("branch", "", "Override branch name from export file")
	workspaceImportCmd.Flags().Bool("force", false, "Overwrite existing workspace if it exists")
}
