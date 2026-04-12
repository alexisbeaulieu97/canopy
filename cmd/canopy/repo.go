package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

var repoRegisterCmd = &cobra.Command{
	Use:   "register <alias> <url>",
	Short: "Register a repository alias",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]
		url := args[1]

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		branch, _ := cmd.Flags().GetString("branch")
		description, _ := cmd.Flags().GetString("description")
		tagsRaw, _ := cmd.Flags().GetString("tags")
		force, _ := cmd.Flags().GetBool("force")

		entry := ports.RegistryEntry{
			URL:           url,
			DefaultBranch: branch,
			Description:   description,
			Tags:          parseTags(tagsRaw),
		}

		if err := app.Config.GetRegistry().Register(alias, entry, force); err != nil {
			return err
		}

		rollbackFn := func() error {
			return app.Config.GetRegistry().Unregister(alias)
		}
		if err := saveRegistryWithRollback(app.Config.GetRegistry(), rollbackFn, "registration", app.Logger); err != nil {
			return err
		}

		output.Infof("Registered '%s' -> %s", alias, url)

		return nil
	},
}

var repoUnregisterCmd = &cobra.Command{
	Use:   "unregister <alias>",
	Short: "Remove a repository alias from the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		entry, exists := app.Config.GetRegistry().Resolve(alias)
		if !exists {
			return cerrors.NewRepoNotFound(alias)
		}

		if err := app.Config.GetRegistry().Unregister(alias); err != nil {
			return err
		}

		rollbackFn := func() error {
			return app.Config.GetRegistry().Register(alias, entry, true)
		}
		if err := saveRegistryWithRollback(app.Config.GetRegistry(), rollbackFn, "unregistration", app.Logger); err != nil {
			return err
		}

		output.Infof("Unregistered '%s'", alias)

		return nil
	},
}

var repoListRegistryCmd = &cobra.Command{
	Use:   "list-registry",
	Short: "List registered repository aliases",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		tagsRaw, _ := cmd.Flags().GetString("tags")
		entries := app.Config.GetRegistry().List(parseTags(tagsRaw))

		output.Printf("%s %s %s\n",
			output.Column("ALIAS", output.RepoAliasWidth, output.AccentStyle),
			output.Column("URL", output.RepoURLWidth, output.MutedStyle),
			output.Column("TAGS", output.RepoTagsWidth, output.MutedStyle),
		)

		for _, entry := range entries {
			output.Printf("%s %s %s\n",
				output.Column(entry.Alias, output.RepoAliasWidth, output.AccentStyle),
				output.Column(entry.URL, output.RepoURLWidth, lipgloss.NewStyle()),
				output.Column(strings.Join(entry.Tags, ", "), output.RepoTagsWidth, lipgloss.NewStyle()),
			)
		}

		output.Infof("\n%d entries", len(entries))

		return nil
	},
}

var repoShowCmd = &cobra.Command{
	Use:   "show <alias>",
	Short: "Show registry entry details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		entry, ok := app.Config.GetRegistry().Resolve(alias)
		if !ok {
			return cerrors.NewRepoNotFound(alias)
		}

		output.Infof("Alias:        %s", alias)
		output.Infof("URL:          %s", entry.URL)

		if entry.DefaultBranch != "" {
			output.Infof("Branch:       %s", entry.DefaultBranch)
		}

		if entry.Description != "" {
			output.Infof("Description:  %s", entry.Description)
		}

		if len(entry.Tags) > 0 {
			output.Infof("Tags:         %s", strings.Join(entry.Tags, ", "))
		}

		repoName := giturl.ExtractRepoName(entry.URL)

		canonicalPath := filepath.Join(app.Config.GetProjectsRoot(), repoName)
		if _, err := os.Stat(canonicalPath); err == nil {
			output.Infof("Canonical:    %s (present)", canonicalPath)
		} else {
			output.Infof("Canonical:    %s (missing)", canonicalPath)
		}

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
	repoCmd.AddCommand(repoRegisterCmd)
	repoCmd.AddCommand(repoUnregisterCmd)
	repoCmd.AddCommand(repoListRegistryCmd)
	repoCmd.AddCommand(repoShowCmd)
	repoCmd.AddCommand(repoPathCmd)

	repoListCmd.Flags().Bool("json", false, "Output in JSON format")
	repoPathCmd.Flags().Bool("json", false, "Output in JSON format")
	repoRemoveCmd.Flags().BoolP("force", "f", false, "Force removal even if used by active workspaces")
	repoRemoveCmd.Flags().Bool("dry-run", false, "Preview what would be removed without actually removing")
	repoRemoveCmd.Flags().Bool("json", false, "Output in JSON format (use with --dry-run)")
	repoAddCmd.Flags().String("alias", "", "Override derived alias when auto-registering")
	repoAddCmd.Flags().Bool("no-register", false, "Skip auto-registration in the registry")
	repoRegisterCmd.Flags().Bool("force", false, "Overwrite existing alias if present")
	repoRegisterCmd.Flags().String("branch", "", "Default branch for the repository")
	repoRegisterCmd.Flags().String("description", "", "Description for the repository")
	repoRegisterCmd.Flags().String("tags", "", "Comma-separated tags for filtering")
	repoListRegistryCmd.Flags().String("tags", "", "Filter registry entries by comma-separated tags")

	repoStatusCmd.Flags().Bool("json", false, "Output in JSON format")
}
