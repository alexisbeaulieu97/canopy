package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// Export format constants.
const (
	formatJSON = "json"
	formatYAML = "yaml"
)

var (
	workspaceCmd = &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"w"},
		Short:   "Manage workspaces",
	}

	workspaceNewCmd = &cobra.Command{
		Use:   "new <ID>",
		Short: "Create a new workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			repos, _ := cmd.Flags().GetStringSlice("repos")
			branch, _ := cmd.Flags().GetString("branch")
			printPath, _ := cmd.Flags().GetBool("print-path")
			noHooks, _ := cmd.Flags().GetBool("no-hooks")
			hooksOnly, _ := cmd.Flags().GetBool("hooks-only")

			if hooksOnly && noHooks {
				return cerrors.NewInvalidArgument("flags", "cannot use --hooks-only with --no-hooks")
			}

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service
			cfg := app.Config

			if hooksOnly {
				if len(repos) > 0 {
					return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --repos")
				}

				if branch != "" {
					return cerrors.NewInvalidArgument("flags", "--hooks-only cannot override branch")
				}

				if printPath {
					return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --print-path")
				}

				if err := service.RunHooks(id, workspaces.HookPhasePostCreate, false); err != nil {
					return err
				}

				output.Success("Ran post_create hooks for workspace", id)
				return nil
			}

			// Resolve repos
			var resolvedRepos []domain.Repo
			if len(repos) > 0 {
				resolvedRepos, err = service.ResolveRepos(id, repos)
				if err != nil {
					return err
				}
			} else {
				resolvedRepos, err = service.ResolveRepos(id, nil)
				if err != nil {
					if errors.Is(err, cerrors.NoReposConfigured) {
						resolvedRepos = []domain.Repo{}
					} else {
						return err
					}
				}
			}

			opts := workspaces.CreateOptions{
				SkipHooks: noHooks,
			}

			dirName, err := service.CreateWorkspaceWithOptions(cmd.Context(), id, branch, resolvedRepos, opts)
			if err != nil {
				return err
			}

			if printPath {
				output.Printf("%s/%s", cfg.GetWorkspacesRoot(), dirName)
			} else {
				output.SuccessWithPath("Created workspace", id, cfg.GetWorkspacesRoot()+"/"+dirName)
			}
			return nil
		},
	}

	workspaceReopenCmd = &cobra.Command{
		Use:     "reopen <ID>",
		Aliases: []string{"open"},
		Short:   "Reopen a closed workspace",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			force, _ := cmd.Flags().GetBool("force")

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			if err := app.Service.RestoreWorkspace(cmd.Context(), id, force); err != nil {
				return err
			}

			output.Success("Restored workspace", id)
			return nil
		},
	}

	workspaceRenameCmd = &cobra.Command{
		Use:   "rename <OLD-ID> <NEW-ID>",
		Short: "Rename a workspace",
		Long:  `Rename a workspace to a new ID. Optionally renames branches in all repos if they match the old ID.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldID := args[0]
			newID := args[1]

			// Fast-fail if IDs are the same
			if oldID == newID {
				return fmt.Errorf("old and new workspace IDs are the same")
			}

			renameBranch, _ := cmd.Flags().GetBool("rename-branch")
			force, _ := cmd.Flags().GetBool("force")

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			if err := app.Service.RenameWorkspace(cmd.Context(), oldID, newID, renameBranch, force); err != nil {
				return err
			}

			if renameBranch {
				output.Infof("Renamed workspace %s to %s (branches also renamed)", oldID, newID)
			} else {
				output.Infof("Renamed workspace %s to %s", oldID, newID)
			}
			return nil
		},
	}

	workspaceListCmd = &cobra.Command{
		Use:   "list",
		Short: "List active workspaces",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service

			jsonOutput, _ := cmd.Flags().GetBool("json")
			closedOnly, _ := cmd.Flags().GetBool("closed")

			if closedOnly {
				archives, err := service.ListClosedWorkspaces()
				if err != nil {
					return err
				}

				if jsonOutput {
					var payload []domain.Workspace

					for _, a := range archives {
						payload = append(payload, a.Metadata)
					}

					return output.PrintJSON(map[string]interface{}{
						"workspaces": payload,
					})
				}

				for _, a := range archives {
					closedDate := "unknown"
					if a.Metadata.ClosedAt != nil {
						closedDate = a.Metadata.ClosedAt.Format(time.RFC3339)
					}

					output.Infof("%s (Closed: %s)", a.Metadata.ID, closedDate)
					for _, r := range a.Metadata.Repos {
						output.Infof("  - %s (%s)", r.Name, r.URL)
					}
				}

				return nil
			}

			list, err := service.ListWorkspaces()
			if err != nil {
				return err
			}

			if jsonOutput {
				return output.PrintJSON(map[string]interface{}{
					"workspaces": list,
				})
			}

			for _, w := range list {
				output.Infof("%s (Branch: %s)", w.ID, w.BranchName)
				for _, r := range w.Repos {
					output.Infof("  - %s (%s)", r.Name, r.URL)
				}
			}
			return nil
		},
	}

	workspaceCloseCmd = &cobra.Command{
		Use:   "close <ID>",
		Short: "Close a workspace (keep metadata or delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			force, _ := cmd.Flags().GetBool("force")
			keepFlag, _ := cmd.Flags().GetBool("keep")
			deleteFlag, _ := cmd.Flags().GetBool("delete")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			noHooks, _ := cmd.Flags().GetBool("no-hooks")
			hooksOnly, _ := cmd.Flags().GetBool("hooks-only")

			if keepFlag && deleteFlag {
				return cerrors.NewInvalidArgument("flags", "cannot use --keep and --delete together")
			}

			if hooksOnly && noHooks {
				return cerrors.NewInvalidArgument("flags", "cannot use --hooks-only with --no-hooks")
			}

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service
			configDefaultArchive := strings.EqualFold(app.Config.GetCloseDefault(), "archive")
			interactive := isInteractiveTerminal()

			closeOpts := workspaces.CloseOptions{
				SkipHooks: noHooks,
			}

			if hooksOnly {
				if keepFlag || deleteFlag {
					return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --keep or --delete")
				}

				if dryRun {
					return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --dry-run")
				}

				if jsonOutput {
					return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --json")
				}

				if err := service.RunHooks(id, workspaces.HookPhasePreClose, false); err != nil {
					return err
				}

				output.Success("Ran pre_close hooks for workspace", id)
				return nil
			}

			// Determine keepMetadata based on flags and config
			keepMetadata := configDefaultArchive
			if keepFlag {
				keepMetadata = true
			} else if deleteFlag {
				keepMetadata = false
			}

			// Handle dry-run mode
			if dryRun {
				preview, err := service.PreviewCloseWorkspace(id, keepMetadata)
				if err != nil {
					return err
				}

				if jsonOutput {
					return output.PrintJSON(map[string]interface{}{
						"dry_run": true,
						"preview": preview,
					})
				}

				printWorkspaceClosePreview(preview)
				return nil
			}

			if keepFlag {
				return keepAndPrint(service, id, force, closeOpts)
			}

			if deleteFlag {
				return closeAndPrint(service, id, force, closeOpts)
			}

			if !interactive {
				if configDefaultArchive {
					return keepAndPrint(service, id, force, closeOpts)
				}

				return closeAndPrint(service, id, force, closeOpts)
			}

			reader := bufio.NewReader(os.Stdin)
			promptSuffix := "[y/N]"
			if configDefaultArchive {
				promptSuffix = "[Y/n]"
			}

			output.Printf("Keep workspace record without files? %s: ", promptSuffix)

			answer, err := reader.ReadString('\n')
			if err != nil {
				if configDefaultArchive {
					return keepAndPrint(service, id, force, closeOpts)
				}

				return closeAndPrint(service, id, force, closeOpts)
			}

			answer = strings.ToLower(strings.TrimSpace(answer))

			switch answer {
			case "y", "yes":
				return keepAndPrint(service, id, force, closeOpts)
			case "n", "no":
				return closeAndPrint(service, id, force, closeOpts)
			case "":
				if configDefaultArchive {
					return keepAndPrint(service, id, force, closeOpts)
				}

				return closeAndPrint(service, id, force, closeOpts)
			default:
				if configDefaultArchive {
					return keepAndPrint(service, id, force, closeOpts)
				}

				return closeAndPrint(service, id, force, closeOpts)
			}
		},
	}

	workspaceRepoAddCmd = &cobra.Command{
		Use:   "add <WORKSPACE-ID> <REPO-NAME>",
		Short: "Add a repository to an existing workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID := args[0]
			repoName := args[1]

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service

			if err := service.AddRepoToWorkspace(cmd.Context(), workspaceID, repoName); err != nil {
				return err
			}

			output.Infof("Added repository %s to workspace %s", repoName, workspaceID)
			return nil
		},
	}

	workspaceRepoRemoveCmd = &cobra.Command{
		Use:   "remove <WORKSPACE-ID> <REPO-NAME>",
		Short: "Remove a repository from an existing workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID := args[0]
			repoName := args[1]

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service

			if err := service.RemoveRepoFromWorkspace(cmd.Context(), workspaceID, repoName); err != nil {
				return err
			}

			output.Infof("Removed repository %s from workspace %s", repoName, workspaceID)
			return nil
		},
	}

	workspaceViewCmd = &cobra.Command{
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

			status, err := service.GetStatus(id)
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

	workspacePathCmd = &cobra.Command{
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

			path, err := app.Service.WorkspacePath(id)
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

	workspaceBranchCmd = &cobra.Command{
		Use:   "branch <ID> <BRANCH-NAME>",
		Short: "Switch branch for all repositories in a workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			branchName := args[1]
			create, _ := cmd.Flags().GetBool("create")

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service

			if err := service.SwitchBranch(cmd.Context(), id, branchName, create); err != nil {
				return err
			}

			output.Infof("Switched workspace %s to branch %s", id, branchName)
			return nil
		},
	}

	workspaceGitCmd = &cobra.Command{
		Use:   "git <WORKSPACE-ID> [--] <git-args...>",
		Short: "Run a git command across all repositories in a workspace",
		Long: `Execute any git command in all repositories within a workspace.

The command is run in each repository and results are displayed with clear separation.
Use -- to separate flags for the git command from canopy flags.

Examples:
  canopy workspace git my-workspace status
  canopy workspace git my-workspace -- fetch --all
  canopy workspace git my-workspace --parallel pull`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID := args[0]
			gitArgs := args[1:]

			parallel, _ := cmd.Flags().GetBool("parallel")
			continueOnError, _ := cmd.Flags().GetBool("continue-on-error")

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			opts := workspaces.GitRunOptions{
				Parallel:        parallel,
				ContinueOnError: continueOnError,
			}

			results, err := app.Service.RunGitInWorkspace(cmd.Context(), workspaceID, gitArgs, opts)
			if err != nil && !continueOnError {
				// Print any results we got before the error
				printGitResults(results)
				return err
			}

			printGitResults(results)

			// Count failures for exit code
			var failures int
			for _, r := range results {
				if r.Error != nil || r.ExitCode != 0 {
					failures++
				}
			}

			if failures > 0 {
				output.Infof("\n%d/%d repos failed", failures, len(results))
				return cerrors.NewCommandFailed("git", fmt.Errorf("%d repos failed", failures))
			}

			output.Infof("\nAll %d repos completed successfully", len(results))

			return nil
		},
	}

	workspaceExportCmd = &cobra.Command{
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

			// --json flag is shorthand for --format json
			if jsonOutput {
				format = formatJSON
			}

			// Validate format
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

			// Write to file or stdout
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

	workspaceImportCmd = &cobra.Command{
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

			// Read input from file or stdin
			var data []byte
			if inputFile == "-" {
				data, err = io.ReadAll(os.Stdin)
			} else {
				data, err = os.ReadFile(inputFile) //nolint:gosec // user-specified input file
			}
			if err != nil {
				return cerrors.NewIOFailed("read import file", err)
			}

			// Parse as YAML (which also handles JSON)
			var export domain.WorkspaceExport
			if err := yaml.Unmarshal(data, &export); err != nil {
				return cerrors.NewInvalidArgument("file", fmt.Sprintf("invalid export format: %v", err))
			}

			// Validate export
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
)

func printGitResults(results []workspaces.RepoGitResult) {
	for i, r := range results {
		if i > 0 {
			output.Info("")
		}

		output.Printf("\033[1;36m=== %s ===\033[0m\n", r.RepoName)

		if r.Error != nil {
			output.Printf("\033[1;31mError: %s\033[0m\n", r.Error)
			continue
		}

		if r.Stdout != "" {
			output.Print(r.Stdout)
		}

		if r.Stderr != "" {
			output.Print(r.Stderr)
		}

		if r.ExitCode != 0 {
			output.Printf("\033[1;31mExit code: %d\033[0m\n", r.ExitCode)
		}
	}
}

func printWorkspaceClosePreview(preview *domain.WorkspaceClosePreview) {
	if preview == nil {
		return
	}

	output.Printf("\033[33m[DRY RUN]\033[0m Would close workspace: %s\n", preview.WorkspaceID)

	action := "Delete"
	if preview.KeepMetadata {
		action = "Archive (keep metadata)"
	}

	output.Infof("  Action: %s", action)
	output.Infof("  Remove directory: %s", preview.WorkspacePath)

	if len(preview.ReposAffected) > 0 {
		output.Infof("  Repos affected: %s", strings.Join(preview.ReposAffected, ", "))
	}

	// Show warnings for repos with uncommitted changes or unpushed commits
	for _, status := range preview.RepoStatuses {
		if status.IsDirty {
			output.Printf("  \033[33m⚠ %s has uncommitted changes\033[0m\n", status.Name)
		}

		if status.UnpushedCount > 0 {
			output.Printf("  \033[33m⚠ %s has %d unpushed commit(s)\033[0m\n", status.Name, status.UnpushedCount)
		}
	}

	if preview.DiskUsageBytes > 0 {
		output.Infof("  Total size: %s", output.FormatBytes(preview.DiskUsageBytes))
	}
}

func keepAndPrint(service *workspaces.Service, id string, force bool, opts workspaces.CloseOptions) error {
	archived, err := service.CloseWorkspaceKeepMetadataWithOptions(id, force, opts)
	if err != nil {
		return err
	}

	var archivedAt *time.Time
	if archived != nil {
		archivedAt = archived.Metadata.ClosedAt
	}

	printClosed(id, archivedAt)

	return nil
}

func closeAndPrint(service *workspaces.Service, id string, force bool, opts workspaces.CloseOptions) error {
	if err := service.CloseWorkspaceWithOptions(id, force, opts); err != nil {
		return err
	}

	output.Success("Closed workspace", id)

	return nil
}

func printClosed(id string, closedAt *time.Time) {
	if closedAt != nil {
		output.Infof("Closed workspace %s at %s", id, closedAt.Format(time.RFC3339))
		return
	}

	output.Success("Closed workspace", id)
}

func isInteractiveTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.AddCommand(workspaceNewCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceCloseCmd)
	workspaceCmd.AddCommand(workspaceReopenCmd)
	workspaceCmd.AddCommand(workspaceRenameCmd)
	workspaceCmd.AddCommand(workspaceViewCmd)
	workspaceCmd.AddCommand(workspacePathCmd)
	workspaceCmd.AddCommand(workspaceBranchCmd)
	workspaceCmd.AddCommand(workspaceGitCmd)
	workspaceCmd.AddCommand(workspaceExportCmd)
	workspaceCmd.AddCommand(workspaceImportCmd)

	// Repo subcommands
	workspaceRepoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories in a workspace",
	}
	workspaceCmd.AddCommand(workspaceRepoCmd)
	workspaceRepoCmd.AddCommand(workspaceRepoAddCmd)
	workspaceRepoCmd.AddCommand(workspaceRepoRemoveCmd)

	workspaceNewCmd.Flags().StringSlice("repos", []string{}, "List of repositories to include")
	workspaceNewCmd.Flags().String("branch", "", "Custom branch name (optional)")
	workspaceNewCmd.Flags().Bool("print-path", false, "Print the created workspace path to stdout")
	workspaceNewCmd.Flags().Bool("no-hooks", false, "Skip post_create hooks")
	workspaceNewCmd.Flags().Bool("hooks-only", false, "Run post_create hooks without creating the workspace")

	workspaceListCmd.Flags().Bool("json", false, "Output in JSON format")
	workspaceListCmd.Flags().Bool("closed", false, "List closed workspaces")

	workspaceViewCmd.Flags().Bool("json", false, "Output in JSON format")
	workspacePathCmd.Flags().Bool("json", false, "Output in JSON format")

	workspaceCloseCmd.Flags().Bool("force", false, "Force close even if there are uncommitted changes")
	workspaceCloseCmd.Flags().Bool("keep", false, "Keep metadata (close without deleting)")
	workspaceCloseCmd.Flags().Bool("delete", false, "Delete without keeping metadata")
	workspaceCloseCmd.Flags().Bool("dry-run", false, "Preview what would be deleted without actually deleting")
	workspaceCloseCmd.Flags().Bool("json", false, "Output in JSON format (use with --dry-run)")
	workspaceCloseCmd.Flags().Bool("no-hooks", false, "Skip pre_close hooks")
	workspaceCloseCmd.Flags().Bool("hooks-only", false, "Run pre_close hooks without closing the workspace")
	workspaceReopenCmd.Flags().Bool("force", false, "Overwrite existing workspace if one already exists")

	workspaceRenameCmd.Flags().Bool("rename-branch", true, "Rename branches in repos if they match the old workspace ID")
	workspaceRenameCmd.Flags().Bool("force", false, "Overwrite if target workspace already exists")

	workspaceBranchCmd.Flags().Bool("create", false, "Create branch if it doesn't exist")

	workspaceGitCmd.Flags().Bool("parallel", false, "Execute git command in repos concurrently")
	workspaceGitCmd.Flags().Bool("continue-on-error", false, "Continue execution even if a repo fails")

	// Export flags
	workspaceExportCmd.Flags().StringP("output", "o", "", "Write export to file instead of stdout")
	workspaceExportCmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json")
	workspaceExportCmd.Flags().Bool("json", false, "Output in JSON format (shorthand for --format json)")

	// Import flags
	workspaceImportCmd.Flags().String("id", "", "Override workspace ID from export file")
	workspaceImportCmd.Flags().String("branch", "", "Override branch name from export file")
	workspaceImportCmd.Flags().Bool("force", false, "Overwrite existing workspace if it exists")
}
