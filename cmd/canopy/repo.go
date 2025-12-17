package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/giturl"
	"github.com/alexisbeaulieu97/canopy/internal/gitx"
	"github.com/alexisbeaulieu97/canopy/internal/output"
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
			fmt.Printf("%s (%s)\n", repo, path) //nolint:forbidigo // user-facing CLI output
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

			entry := config.RegistryEntry{URL: url}
			realAlias, err := registerWithPrompt(cmd, app.Config.GetRegistry(), alias, entry)
			if err != nil {
				// Use a detached context for cleanup to ensure it runs even if cmd.Context() is cancelled
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), gitx.DefaultLocalTimeout)
				defer cleanupCancel()

				if rmErr := svc.RemoveCanonicalRepo(cleanupCtx, name, true); rmErr != nil {
					app.Logger.Errorf("Failed to rollback repo removal: %v", rmErr)
				}

				return cerrors.NewRegistryError("register", "registration failed", err)
			}
			fmt.Printf("Registered repository as '%s'\n", realAlias) //nolint:forbidigo // user-facing CLI output
		}

		fmt.Printf("Added repository %s\n", url) //nolint:forbidigo // user-facing CLI output
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

		fmt.Printf("Removed repository %s\n", name) //nolint:forbidigo // user-facing CLI output
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

		fmt.Printf("Synced repository %s\n", name) //nolint:forbidigo // user-facing CLI output
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

		entry := config.RegistryEntry{
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

		fmt.Printf("Registered '%s' -> %s\n", alias, url) //nolint:forbidigo // user-facing CLI output
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

		fmt.Printf("Unregistered '%s'\n", alias) //nolint:forbidigo // user-facing CLI output
		return nil
	},
}

const (
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

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

		fmt.Printf("%s%-16s%s %-45s %-20s\n", colorGreen, "ALIAS", colorReset, "URL", "TAGS") //nolint:forbidigo // user-facing CLI output
		for _, entry := range entries {
			fmt.Printf("%s%-16s%s %-45s %-20s\n", colorGreen, entry.Alias, colorReset, entry.URL, strings.Join(entry.Tags, ",")) //nolint:forbidigo // user-facing CLI output
		}
		fmt.Printf("\n%d entries\n", len(entries)) //nolint:forbidigo // user-facing CLI output
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

		fmt.Printf("Alias:        %s\n", alias)     //nolint:forbidigo // user-facing CLI output
		fmt.Printf("URL:          %s\n", entry.URL) //nolint:forbidigo // user-facing CLI output
		if entry.DefaultBranch != "" {
			fmt.Printf("Branch:       %s\n", entry.DefaultBranch) //nolint:forbidigo // user-facing CLI output
		}
		if entry.Description != "" {
			fmt.Printf("Description:  %s\n", entry.Description) //nolint:forbidigo // user-facing CLI output
		}
		if len(entry.Tags) > 0 {
			fmt.Printf("Tags:         %s\n", strings.Join(entry.Tags, ", ")) //nolint:forbidigo // user-facing CLI output
		}

		repoName := giturl.ExtractRepoName(entry.URL)
		canonicalPath := filepath.Join(app.Config.GetProjectsRoot(), repoName)
		if _, err := os.Stat(canonicalPath); err == nil {
			fmt.Printf("Canonical:    %s (present)\n", canonicalPath) //nolint:forbidigo // user-facing CLI output
		} else {
			fmt.Printf("Canonical:    %s (missing)\n", canonicalPath) //nolint:forbidigo // user-facing CLI output
		}

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

		fmt.Println(path) //nolint:forbidigo // user-facing CLI output
		return nil
	},
}

func parseTags(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")

	var tags []string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}

	return tags
}

func registerWithPrompt(cmd *cobra.Command, registry *config.RepoRegistry, alias string, entry config.RegistryEntry) (string, error) {
	if registry == nil {
		return alias, cerrors.NewConfigInvalid("registry not configured")
	}

	target := strings.TrimSpace(alias)
	if target == "" {
		return "", cerrors.NewInvalidArgument("alias", "alias is required")
	}

	for {
		if _, exists := registry.Resolve(target); !exists {
			return registerAlias(registry, target, entry)
		}

		suggested := nextAvailableAlias(registry, target)

		var err error

		target, err = promptAlias(cmd, target, suggested)
		if err != nil {
			return "", err
		}
	}
}

func nextAvailableAlias(registry *config.RepoRegistry, base string) string {
	target := base
	for idx := 2; ; idx++ {
		if _, exists := registry.Resolve(target); !exists {
			return target
		}

		target = fmt.Sprintf("%s-%d", base, idx)
	}
}

func registerAlias(registry *config.RepoRegistry, alias string, entry config.RegistryEntry) (string, error) {
	if err := registry.Register(alias, entry, false); err != nil {
		return "", err
	}

	rollbackFn := func() error {
		return registry.Unregister(alias)
	}
	if err := saveRegistryWithRollback(registry, rollbackFn, "registration", nil); err != nil {
		return "", err
	}

	return alias, nil
}

// rollbackLogger is an interface for logging rollback errors.
type rollbackLogger interface {
	Errorf(format string, args ...interface{})
}

// saveRegistryWithRollback saves the registry and performs a rollback on failure.
// It logs any errors that occur during rollback and returns the save error if present.
// If logger is nil, rollback errors are silently discarded.
func saveRegistryWithRollback(
	registry *config.RepoRegistry,
	rollbackFn func() error,
	rollbackDesc string,
	logger rollbackLogger,
) error {
	if err := registry.Save(); err != nil {
		if rollbackErr := rollbackFn(); rollbackErr != nil {
			if logger != nil {
				logger.Errorf("Failed to rollback %s: %v", rollbackDesc, rollbackErr)
			}
		} else if rollbackSaveErr := registry.Save(); rollbackSaveErr != nil {
			if logger != nil {
				logger.Errorf("Failed to save rollback: %v", rollbackSaveErr)
			}
		}

		return cerrors.NewRegistryError("save", "failed to save registry", err)
	}

	return nil
}

func promptAlias(cmd *cobra.Command, alias, suggested string) (string, error) {
	reader := bufio.NewReader(cmd.InOrStdin())
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Alias '%s' already exists. Enter a new alias or press Enter to use '%s': ", alias, suggested); err != nil {
		return "", err
	}

	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return suggested, nil
	}

	return input, nil
}

func printRepoRemovePreview(preview *domain.RepoRemovePreview) {
	if preview == nil {
		return
	}

	fmt.Printf("%s[DRY RUN]%s Would remove repository: %s\n", colorYellow, colorReset, preview.RepoName) //nolint:forbidigo // user-facing CLI output
	fmt.Printf("  Remove directory: %s\n", preview.RepoPath)                                             //nolint:forbidigo // user-facing CLI output

	if len(preview.WorkspacesAffected) > 0 {
		fmt.Printf("  Used by workspaces: %s (will become orphaned)\n", strings.Join(preview.WorkspacesAffected, ", ")) //nolint:forbidigo // user-facing CLI output
	}

	if preview.DiskUsageBytes > 0 {
		fmt.Printf("  Size: %s\n", output.FormatBytes(preview.DiskUsageBytes)) //nolint:forbidigo // user-facing CLI output
	}
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoSyncCmd)
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
}
