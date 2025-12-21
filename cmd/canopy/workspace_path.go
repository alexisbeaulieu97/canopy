package main

import (
	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_path.go defines the "workspace path" subcommand.

var workspacePathCmd = &cobra.Command{
	Use:   "path <ID>",
	Short: "Print the absolute path of a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		path, err := app.Service.WorkspacePath(cmd.Context(), id)
		if err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(map[string]string{
				"path": path,
			})
		}

		output.Println(path)
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspacePathCmd)

	workspacePathCmd.Flags().Bool("json", false, "Output in JSON format")
}
