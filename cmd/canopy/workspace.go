package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
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

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service
			cfg := app.Config

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

			dirName, err := service.CreateWorkspace(id, branch, resolvedRepos)
			if err != nil {
				return err
			}

			if printPath {
				fmt.Printf("%s/%s", cfg.GetWorkspacesRoot(), dirName) //nolint:forbidigo // user-facing CLI output
			} else {
				fmt.Printf("Created workspace %s in %s/%s\n", id, cfg.GetWorkspacesRoot(), dirName) //nolint:forbidigo // user-facing CLI output
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

			if err := app.Service.RestoreWorkspace(id, force); err != nil {
				return err
			}

			fmt.Printf("Restored workspace %s\n", id) //nolint:forbidigo // user-facing CLI output
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

					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					return encoder.Encode(payload)
				}

				for _, a := range archives {
					closedDate := "unknown"
					if a.Metadata.ClosedAt != nil {
						closedDate = a.Metadata.ClosedAt.Format(time.RFC3339)
					}

					fmt.Printf("%s (Closed: %s)\n", a.Metadata.ID, closedDate) //nolint:forbidigo // user-facing CLI output
					for _, r := range a.Metadata.Repos {
						fmt.Printf("  - %s (%s)\n", r.Name, r.URL) //nolint:forbidigo // user-facing CLI output
					}
				}

				return nil
			}

			list, err := service.ListWorkspaces()
			if err != nil {
				return err
			}

			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(list)
			}

			for _, w := range list {
				fmt.Printf("%s (Branch: %s)\n", w.ID, w.BranchName) //nolint:forbidigo // user-facing CLI output
				for _, r := range w.Repos {
					fmt.Printf("  - %s (%s)\n", r.Name, r.URL) //nolint:forbidigo // user-facing CLI output
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

			if keepFlag && deleteFlag {
				return cerrors.NewInvalidArgument("flags", "cannot use --keep and --delete together")
			}

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service
			configDefaultArchive := strings.EqualFold(app.Config.GetCloseDefault(), "archive")
			interactive := isInteractiveTerminal()

			if keepFlag {
				return keepAndPrint(service, id, force)
			}

			if deleteFlag {
				return closeAndPrint(service, id, force)
			}

			if !interactive {
				if configDefaultArchive {
					return keepAndPrint(service, id, force)
				}

				return closeAndPrint(service, id, force)
			}

			reader := bufio.NewReader(os.Stdin)
			promptSuffix := "[y/N]"
			if configDefaultArchive {
				promptSuffix = "[Y/n]"
			}

			fmt.Printf("Keep workspace record without files? %s: ", promptSuffix) //nolint:forbidigo // user prompt

			answer, err := reader.ReadString('\n')
			if err != nil {
				if configDefaultArchive {
					return keepAndPrint(service, id, force)
				}

				return closeAndPrint(service, id, force)
			}

			answer = strings.ToLower(strings.TrimSpace(answer))

			switch answer {
			case "y", "yes":
				return keepAndPrint(service, id, force)
			case "n", "no":
				return closeAndPrint(service, id, force)
			case "":
				if configDefaultArchive {
					return keepAndPrint(service, id, force)
				}

				return closeAndPrint(service, id, force)
			default:
				if configDefaultArchive {
					return keepAndPrint(service, id, force)
				}

				return closeAndPrint(service, id, force)
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

			if err := service.AddRepoToWorkspace(workspaceID, repoName); err != nil {
				return err
			}

			fmt.Printf("Added repository %s to workspace %s\n", repoName, workspaceID) //nolint:forbidigo // user-facing CLI output
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

			if err := service.RemoveRepoFromWorkspace(workspaceID, repoName); err != nil {
				return err
			}

			fmt.Printf("Removed repository %s from workspace %s\n", repoName, workspaceID) //nolint:forbidigo // user-facing CLI output
			return nil
		},
	}

	workspaceViewCmd = &cobra.Command{
		Use:   "view <ID>",
		Short: "View details of a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service

			status, err := service.GetStatus(id)
			if err != nil {
				return err
			}

			fmt.Printf("Workspace: %s\n", status.ID)      //nolint:forbidigo // user-facing CLI output
			fmt.Printf("Branch: %s\n", status.BranchName) //nolint:forbidigo // user-facing CLI output

			fmt.Println("Repositories:") //nolint:forbidigo // user-facing CLI output
			for _, r := range status.Repos {
				statusStr := "Clean"
				if r.IsDirty {
					statusStr = "Dirty"
				}
				fmt.Printf("  - %s: %s (Branch: %s, Unpushed: %d)\n", r.Name, statusStr, r.Branch, r.UnpushedCommits) //nolint:forbidigo // user-facing CLI output
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

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			path, err := app.Service.WorkspacePath(id)
			if err != nil {
				return err
			}

			fmt.Println(path) //nolint:forbidigo // user-facing CLI output
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

			if err := service.SwitchBranch(id, branchName, create); err != nil {
				return err
			}

			fmt.Printf("Switched workspace %s to branch %s\n", id, branchName) //nolint:forbidigo // user-facing CLI output
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

			results, err := app.Service.RunGitInWorkspace(workspaceID, gitArgs, opts)
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
				fmt.Printf("\n%d/%d repos failed\n", failures, len(results)) //nolint:forbidigo // user-facing CLI output
				return cerrors.NewCommandFailed("git", fmt.Errorf("%d repos failed", failures))
			}

			fmt.Printf("\nAll %d repos completed successfully\n", len(results)) //nolint:forbidigo // user-facing CLI output

			return nil
		},
	}
)

func printGitResults(results []workspaces.RepoGitResult) {
	for i, r := range results {
		if i > 0 {
			fmt.Println() //nolint:forbidigo // user-facing CLI output
		}

		fmt.Printf("\033[1;36m=== %s ===\033[0m\n", r.RepoName) //nolint:forbidigo // user-facing CLI output

		if r.Error != nil {
			fmt.Printf("\033[1;31mError: %s\033[0m\n", r.Error) //nolint:forbidigo // user-facing CLI output
			continue
		}

		if r.Stdout != "" {
			fmt.Print(r.Stdout) //nolint:forbidigo // user-facing CLI output
		}

		if r.Stderr != "" {
			fmt.Print(r.Stderr) //nolint:forbidigo // user-facing CLI output
		}

		if r.ExitCode != 0 {
			fmt.Printf("\033[1;31mExit code: %d\033[0m\n", r.ExitCode) //nolint:forbidigo // user-facing CLI output
		}
	}
}

func keepAndPrint(service *workspaces.Service, id string, force bool) error {
	archived, err := service.CloseWorkspaceKeepMetadata(id, force)
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

func closeAndPrint(service *workspaces.Service, id string, force bool) error {
	if err := service.CloseWorkspace(id, force); err != nil {
		return err
	}

	fmt.Printf("Closed workspace %s\n", id) //nolint:forbidigo // user-facing CLI output

	return nil
}

func printClosed(id string, closedAt *time.Time) {
	if closedAt != nil {
		fmt.Printf("Closed workspace %s at %s\n", id, closedAt.Format(time.RFC3339)) //nolint:forbidigo // user-facing CLI output
		return
	}

	fmt.Printf("Closed workspace %s\n", id) //nolint:forbidigo // user-facing CLI output
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
	workspaceCmd.AddCommand(workspaceViewCmd)
	workspaceCmd.AddCommand(workspacePathCmd)
	workspaceCmd.AddCommand(workspaceBranchCmd)
	workspaceCmd.AddCommand(workspaceGitCmd)

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

	workspaceListCmd.Flags().Bool("json", false, "Output in JSON format")
	workspaceListCmd.Flags().Bool("closed", false, "List closed workspaces")

	workspaceCloseCmd.Flags().Bool("force", false, "Force close even if there are uncommitted changes")
	workspaceCloseCmd.Flags().Bool("keep", false, "Keep metadata (close without deleting)")
	workspaceCloseCmd.Flags().Bool("delete", false, "Delete without keeping metadata")
	workspaceReopenCmd.Flags().Bool("force", false, "Overwrite existing workspace if one already exists")

	workspaceBranchCmd.Flags().Bool("create", false, "Create branch if it doesn't exist")

	workspaceGitCmd.Flags().Bool("parallel", false, "Execute git command in repos concurrently")
	workspaceGitCmd.Flags().Bool("continue-on-error", false, "Continue execution even if a repo fails")
}
