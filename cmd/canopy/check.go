package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/app"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate the current configuration",
	RunE: func(cmd *cobra.Command, _ []string) error {
		appInstance, err := getApp(cmd)
		if err != nil {
			return err
		}

		cfg := appInstance.Config
		jsonOutput, _ := cmd.Flags().GetBool("json")
		checkOrphans, _ := cmd.Flags().GetBool("orphans")

		// If checking orphans, run orphan detection
		if checkOrphans {
			return runOrphanCheck(cmd, appInstance, jsonOutput)
		}

		validationErr := cfg.Validate()

		if jsonOutput {
			if validationErr != nil {
				// Print error JSON but still return the error for non-zero exit code
				_ = output.PrintErrorJSON(validationErr)
				return cerrors.Wrap(cerrors.ErrConfigInvalid, "configuration is invalid", validationErr)
			}

			configInfo := map[string]interface{}{
				"projects_root":    cfg.GetProjectsRoot(),
				"workspaces_root":  cfg.GetWorkspacesRoot(),
				"workspace_naming": cfg.GetWorkspaceNaming(),
				"valid":            true,
			}

			if registry := cfg.GetRegistry(); registry != nil {
				configInfo["registry_path"] = registry.Path()
			}

			return output.PrintJSON(configInfo)
		}

		appInstance.Logger.Info("Configuration loaded successfully.")
		appInstance.Logger.Infof("Projects Root: %s", cfg.GetProjectsRoot())
		appInstance.Logger.Infof("Workspaces Root: %s", cfg.GetWorkspacesRoot())
		appInstance.Logger.Infof("Naming Pattern: %s", cfg.GetWorkspaceNaming())
		if registry := cfg.GetRegistry(); registry != nil {
			appInstance.Logger.Infof("Registry File: %s", registry.Path())
		}

		if validationErr != nil {
			appInstance.Logger.Errorf("Configuration is invalid: %v", validationErr)
			return cerrors.Wrap(cerrors.ErrConfigInvalid, "configuration is invalid", validationErr)
		}

		appInstance.Logger.Info("Configuration is valid.")
		return nil
	},
}

func runOrphanCheck(_ *cobra.Command, appInstance *app.App, jsonOutput bool) error {
	orphans, err := appInstance.Service.DetectOrphans()
	if err != nil {
		if jsonOutput {
			_ = output.PrintErrorJSON(err)
		}

		return err
	}

	if jsonOutput {
		result := map[string]interface{}{
			"orphans": orphans,
			"count":   len(orphans),
		}

		return output.PrintJSON(result)
	}

	if len(orphans) == 0 {
		appInstance.Logger.Info("No orphaned worktrees found.")
		return nil
	}

	fmt.Printf("Found %d orphaned worktree(s):\n", len(orphans)) //nolint:forbidigo // user-facing CLI output

	// Group orphans by workspace for cleaner output
	byWorkspace := make(map[string][]domain.OrphanedWorktree)
	for _, orphan := range orphans {
		byWorkspace[orphan.WorkspaceID] = append(byWorkspace[orphan.WorkspaceID], orphan)
	}

	for wsID, wsOrphans := range byWorkspace {
		fmt.Printf("\n  Workspace: %s\n", wsID) //nolint:forbidigo // user-facing CLI output

		for _, orphan := range wsOrphans {
			fmt.Printf("    - %s: %s\n", orphan.RepoName, orphan.ReasonDescription()) //nolint:forbidigo // user-facing CLI output
		}
	}

	// Print remediation suggestions
	fmt.Println("\nRemediation:") //nolint:forbidigo // user-facing CLI output
	printRemediationSuggestions(orphans)

	return nil
}

func printRemediationSuggestions(orphans []domain.OrphanedWorktree) {
	hasMissingCanonical := false
	hasMissingDir := false
	hasInvalidGit := false

	for _, orphan := range orphans {
		switch orphan.Reason {
		case domain.OrphanReasonCanonicalMissing:
			hasMissingCanonical = true
		case domain.OrphanReasonDirectoryMissing:
			hasMissingDir = true
		case domain.OrphanReasonInvalidGitDir:
			hasInvalidGit = true
		}
	}

	if hasMissingCanonical {
		fmt.Println("  • For missing canonical repos: Run 'canopy repo add <url>' to restore the repo")   //nolint:forbidigo // user-facing CLI output
		fmt.Println("    or remove the reference with 'canopy workspace remove-repo <workspace> <repo>'") //nolint:forbidigo // user-facing CLI output
	}

	if hasMissingDir || hasInvalidGit {
		fmt.Println("  • For missing/invalid worktrees: Remove the workspace and recreate it")                 //nolint:forbidigo // user-facing CLI output
		fmt.Println("    or remove the repo reference with 'canopy workspace remove-repo <workspace> <repo>'") //nolint:forbidigo // user-facing CLI output
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("json", false, "Output in JSON format")
	checkCmd.Flags().Bool("orphans", false, "Check for orphaned worktrees")
}
