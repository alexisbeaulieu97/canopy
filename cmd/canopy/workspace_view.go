package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
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

		renderWorkspaceView(status)

		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceViewCmd)

	workspaceViewCmd.Flags().Bool("json", false, "Output in JSON format")
}

func renderWorkspaceView(status *domain.WorkspaceStatus) {
	icons := output.NewIcons()

	// Build sections for the workspace view
	var sections []output.BoxSection

	// Metadata section (no title since it's the first section)
	var metadataLines []string

	metadataLines = append(metadataLines, output.FormatKeyValue("Branch", status.BranchName, 12))
	// TODO: Add path, disk size, modified, created when available in the status

	sections = append(sections, output.BoxSection{
		Title: "",
		Lines: metadataLines,
	})

	// Repositories section
	var repoLines []string

	// Header
	header := fmt.Sprintf("  %-14s %-20s %s", "NAME", "BRANCH", "STATUS")
	repoLines = append(repoLines, header)
	repoLines = append(repoLines, "  "+output.HorizontalRule(56))

	for _, r := range status.Repos {
		name := r.Name
		if len(name) > 12 {
			name = name[:11] + "…"
		}

		branch := r.Branch
		if len(branch) > 18 {
			branch = branch[:17] + "…"
		}

		statusStr := formatRepoViewStatus(r, icons)

		line := fmt.Sprintf("  %-14s %-20s %s", name, branch, statusStr)
		repoLines = append(repoLines, line)
	}

	sections = append(sections, output.BoxSection{
		Title: fmt.Sprintf("Repositories (%d)", len(status.Repos)),
		Lines: repoLines,
	})

	// Render the box
	title := "Workspace: " + status.ID
	box := output.NewBox(title).WithWidth(70)
	box.RenderWithSections(sections)

	// Print any warnings below the box
	// (In the future, we could detect orphaned worktrees and show a warning box)
}

func formatRepoViewStatus(r domain.RepoStatus, icons output.Icons) string {
	if r.Error != "" {
		errText := strings.ReplaceAll(string(r.Error), "\n", " ")
		if len(errText) > 30 {
			errText = errText[:27] + "..."
		}

		return output.Colorize(output.ErrorStyle, icons.Error()+" error: "+errText)
	}

	var parts []string

	if r.IsDirty {
		parts = append(parts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%s modified", icons.Dirty())))
	}

	if r.UnpushedCommits > 0 {
		parts = append(parts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%s %d unpushed", icons.Unpushed(), r.UnpushedCommits)))
	}

	if r.BehindRemote > 0 {
		parts = append(parts, output.Colorize(output.WarningStyle, fmt.Sprintf("%s %d behind", icons.Behind(), r.BehindRemote)))
	}

	if len(parts) == 0 {
		return output.Colorize(output.SuccessStyle, icons.Success()+" clean")
	}

	return strings.Join(parts, "  ")
}
