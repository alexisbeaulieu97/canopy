package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/yard/internal/domain"
	"github.com/alexisbeaulieu97/yard/internal/workspaces"
)

var (
	workspaceCmd = &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"w", "ticket", "t"}, // Keep aliases for backward compatibility
		Short:   "Manage workspaces",
	}

	workspaceNewCmd = &cobra.Command{
		Use:   "new <ID>",
		Short: "Create a new workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			repos, _ := cmd.Flags().GetStringSlice("repos")
			slug, _ := cmd.Flags().GetString("slug")
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
					if errors.Is(err, workspaces.ErrNoReposConfigured) {
						resolvedRepos = []domain.Repo{}
					} else {
						return err
					}
				}
			}

			dirName, err := service.CreateWorkspace(id, slug, branch, resolvedRepos)
			if err != nil {
				return err
			}

			if printPath {
				fmt.Printf("%s/%s", cfg.WorkspacesRoot, dirName) //nolint:forbidigo // user-facing CLI output
			} else {
				fmt.Printf("Created workspace %s in %s/%s\n", id, cfg.WorkspacesRoot, dirName) //nolint:forbidigo // user-facing CLI output
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

			list, err := service.ListWorkspaces()
			if err != nil {
				return err
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")
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
		Short: "Close (delete) a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			force, _ := cmd.Flags().GetBool("force")

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service

			if err := service.CloseWorkspace(id, force); err != nil {
				return err
			}

			fmt.Printf("Closed workspace %s\n", id) //nolint:forbidigo // user-facing CLI output
			return nil
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

	workspaceSyncCmd = &cobra.Command{
		Use:   "sync <ID>",
		Short: "Sync all repositories in a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			service := app.Service

			if err := service.SyncWorkspace(id); err != nil {
				return err
			}

			fmt.Printf("Synced workspace %s\n", id) //nolint:forbidigo // user-facing CLI output
			return nil
		},
	}

	workspaceSwitchCmd = &cobra.Command{
		Use:   "switch <ID>",
		Short: "Switch to a workspace (prints path for shell integration)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Reuse path command logic
			return workspacePathCmd.RunE(cmd, args)
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
)

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.AddCommand(workspaceNewCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceCloseCmd)
	workspaceCmd.AddCommand(workspaceViewCmd)
	workspaceCmd.AddCommand(workspacePathCmd)
	workspaceCmd.AddCommand(workspaceSyncCmd)
	workspaceCmd.AddCommand(workspaceSwitchCmd)
	workspaceCmd.AddCommand(workspaceBranchCmd)

	// Repo subcommands
	workspaceRepoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories in a workspace",
	}
	workspaceCmd.AddCommand(workspaceRepoCmd)
	workspaceRepoCmd.AddCommand(workspaceRepoAddCmd)
	workspaceRepoCmd.AddCommand(workspaceRepoRemoveCmd)

	workspaceNewCmd.Flags().StringSlice("repos", []string{}, "List of repositories to include")
	workspaceNewCmd.Flags().String("slug", "", "Slug for directory naming (optional)")
	workspaceNewCmd.Flags().String("branch", "", "Custom branch name (optional)")
	workspaceNewCmd.Flags().Bool("print-path", false, "Print the created workspace path to stdout")

	workspaceListCmd.Flags().Bool("json", false, "Output in JSON format")

	workspaceCloseCmd.Flags().Bool("force", false, "Force close even if there are uncommitted changes")

	workspaceBranchCmd.Flags().Bool("create", false, "Create branch if it doesn't exist")
}
