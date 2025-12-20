package main

import (
	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// workspace_view.go defines the "workspace view" subcommand.

var workspaceViewCmd = &cobra.Command{
	Use:   "view <ID>",
	Short: "View details of a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service

		status, err := service.GetStatus(cmd.Context(), id)
		if err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(map[string]interface{}{
				"workspace": status.ID,
				"branch":    status.BranchName,
				"repos":     status.Repos,
			})
		}

		output.Infof("Workspace: %s", status.ID)
		output.Infof("Branch: %s", status.BranchName)

		output.Println("Repositories:")
		for _, r := range status.Repos {
			statusStr := "Clean"
			if r.IsDirty {
				statusStr = "Dirty"
			}
			output.Infof("  - %s: %s (Branch: %s, Unpushed: %d)", r.Name, statusStr, r.Branch, r.UnpushedCommits)
		}
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceViewCmd)

	workspaceViewCmd.Flags().Bool("json", false, "Output in JSON format")
}
