package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/giturl"
	"github.com/alexisbeaulieu97/canopy/internal/gitx"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage canonical repositories",
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List canonical repositories",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		cfg := app.Config
		svc := app.Service
		jsonOutput, _ := cmd.Flags().GetBool("json")

		repos, err := svc.ListCanonicalRepos(cmd.Context())
		if err != nil {
			return err
		}

		if jsonOutput {
			type repoInfo struct {
				Name string `json:"name"`
				Path string `json:"path"`
			}

			var repoList []repoInfo
			for _, repo := range repos {
				repoList = append(repoList, repoInfo{
					Name: repo,
					Path: filepath.Join(cfg.GetProjectsRoot(), repo),
				})
			}

			return output.PrintJSON(map[string]interface{}{
				"repos": repoList,
			})
		}

		for _, repo := range repos {
			path := filepath.Join(cfg.GetProjectsRoot(), repo)
			output.Infof("%s (%s)", repo, path)
		}

		return nil
	},
}

var repoAddCmd = &cobra.Command{
	Use:   "add <URL>",
	Short: "Add a canonical repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		svc := app.Service

		name, err := svc.AddCanonicalRepo(cmd.Context(), url)
		if err != nil {
			return err
		}

		skipRegister, _ := cmd.Flags().GetBool("no-register")
		aliasOverride, _ := cmd.Flags().GetString("alias")

		if !skipRegister {
			alias := aliasOverride
			if alias == "" {
				alias = giturl.DeriveAlias(url)
			}

			if alias == "" {
				alias = name
			}

			entry := ports.RegistryEntry{URL: url}

			realAlias, err := registerWithPrompt(cmd, app.Config.GetRegistry(), alias, entry, app.Logger)
			if err != nil {
				// Use a detached context for cleanup to ensure it runs even if cmd.Context() is cancelled
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), gitx.DefaultLocalTimeout)
				defer cleanupCancel()

				if rmErr := svc.RemoveCanonicalRepo(cleanupCtx, name, true); rmErr != nil {
					app.Logger.Errorf("Failed to rollback repo removal: %v", rmErr)
				}

				return cerrors.NewRegistryError("register", "registration failed", err)
			}

			output.Infof("Registered repository as '%s'", realAlias)
		}

		output.Success("Added repository", url)

		return nil
	},
}

var repoRemoveCmd = &cobra.Command{
	Use:   "remove <NAME>",
	Short: "Remove a canonical repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		svc := app.Service

		// Handle dry-run mode
		if dryRun {
			preview, err := svc.PreviewRemoveCanonicalRepo(cmd.Context(), name)
			if err != nil {
				return err
			}

			if jsonOutput {
				return output.PrintJSON(map[string]interface{}{
					"dry_run": true,
					"preview": preview,
				})
			}

			printRepoRemovePreview(preview)

			return nil
		}

		if err := svc.RemoveCanonicalRepo(cmd.Context(), name, force); err != nil {
			return err
		}

		output.Success("Removed repository", name)

		return nil
	},
}

var repoSyncCmd = &cobra.Command{
	Use:   "sync <NAME>",
	Short: "Sync a canonical repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		svc := app.Service

		if err := svc.SyncCanonicalRepo(cmd.Context(), name); err != nil {
			return err
		}

		output.Success("Synced repository", name)

		return nil
	},
}

var repoStatusCmd = &cobra.Command{
	Use:   "status [NAME]",
	Short: "Show status of canonical repositories",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		svc := app.Service
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if len(args) > 0 {
			name := args[0]

			status, err := svc.GetCanonicalRepoStatus(cmd.Context(), name)
			if err != nil {
				return err
			}

			if jsonOutput {
				return output.PrintJSON(status)
			}

			printSingleRepoStatus(status)

			return nil
		}

		statuses, err := svc.GetAllCanonicalRepoStatuses(cmd.Context())
		if err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(statuses)
		}

		printRepoStatusesTable(statuses)

		return nil
	},
}

var repoPathCmd = &cobra.Command{
	Use:   "path <NAME>",
	Short: "Print the absolute path of a canonical repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		// Check if repo exists
		path := filepath.Join(app.Config.GetProjectsRoot(), name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return cerrors.NewRepoNotFound(name)
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
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoSyncCmd)
	repoCmd.AddCommand(repoStatusCmd)
	repoCmd.AddCommand(repoPathCmd)

	repoListCmd.Flags().Bool("json", false, "Output in JSON format")
	repoPathCmd.Flags().Bool("json", false, "Output in JSON format")
	repoRemoveCmd.Flags().BoolP("force", "f", false, "Force removal even if used by active workspaces")
	repoRemoveCmd.Flags().Bool("dry-run", false, "Preview what would be removed without actually removing")
	repoRemoveCmd.Flags().Bool("json", false, "Output in JSON format (use with --dry-run)")
	repoAddCmd.Flags().String("alias", "", "Override derived alias when auto-registering")
	repoAddCmd.Flags().Bool("no-register", false, "Skip auto-registration in the registry")

	repoStatusCmd.Flags().Bool("json", false, "Output in JSON format")
}
