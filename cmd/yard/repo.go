package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/yard/internal/config"
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

		repos, err := svc.ListCanonicalRepos()
		if err != nil {
			return err
		}

		for _, repo := range repos {
			path := filepath.Join(cfg.ProjectsRoot, repo)
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

		name, err := svc.AddCanonicalRepo(url)
		if err != nil {
			return err
		}

		skipRegister, _ := cmd.Flags().GetBool("no-register")
		aliasOverride, _ := cmd.Flags().GetString("alias")

		if !skipRegister {
			alias := aliasOverride
			if alias == "" {
				alias = config.DeriveAliasFromURL(url)
			}
			if alias == "" {
				alias = name
			}

			entry := config.RegistryEntry{URL: url}
			realAlias, err := registerWithPrompt(cmd, app.Config.Registry, alias, entry)
			if err != nil {
				if rmErr := svc.RemoveCanonicalRepo(name, true); rmErr != nil {
					return fmt.Errorf("registration failed: %v (rollback failed: %v)", err, rmErr)
				}

				return fmt.Errorf("registration failed: %w", err)
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

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		svc := app.Service

		if err := svc.RemoveCanonicalRepo(name, force); err != nil {
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

		if err := svc.SyncCanonicalRepo(name); err != nil {
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

		if err := app.Config.Registry.Register(alias, entry, force); err != nil {
			return err
		}
		if err := app.Config.Registry.Save(); err != nil {
			if rollbackErr := app.Config.Registry.Unregister(alias); rollbackErr != nil {
				return fmt.Errorf("failed to save registry: %v (rollback failed: %v)", err, rollbackErr)
			}

			if rollbackSaveErr := app.Config.Registry.Save(); rollbackSaveErr != nil {
				return fmt.Errorf("failed to save registry: %v (rollback save failed: %v)", err, rollbackSaveErr)
			}

			return fmt.Errorf("failed to save registry: %w", err)
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

		entry, exists := app.Config.Registry.Resolve(alias)
		if !exists {
			return fmt.Errorf("alias '%s' not found", alias)
		}

		if err := app.Config.Registry.Unregister(alias); err != nil {
			return err
		}
		if err := app.Config.Registry.Save(); err != nil {
			if restoreErr := app.Config.Registry.Register(alias, entry, true); restoreErr != nil {
				return fmt.Errorf("failed to save registry: %v (restore failed: %v)", err, restoreErr)
			}

			if restoreSaveErr := app.Config.Registry.Save(); restoreSaveErr != nil {
				return fmt.Errorf("failed to save registry: %v (restore save failed: %v)", err, restoreSaveErr)
			}

			return fmt.Errorf("failed to save registry: %w", err)
		}

		fmt.Printf("Unregistered '%s'\n", alias) //nolint:forbidigo // user-facing CLI output
		return nil
	},
}

const (
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
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
		entries := app.Config.Registry.List(parseTags(tagsRaw))

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

		entry, ok := app.Config.Registry.Resolve(alias)
		if !ok {
			return fmt.Errorf("alias '%s' not found", alias)
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

		repoName := repoNameFromURL(entry.URL)
		canonicalPath := filepath.Join(app.Config.ProjectsRoot, repoName)
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

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		// Check if repo exists
		path := filepath.Join(app.Config.ProjectsRoot, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("repository %s not found", name)
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

func repoNameFromURL(url string) string {
	if strings.Contains(url, ":") && !strings.HasPrefix(url, "http") {
		parts := strings.Split(url, ":")
		url = parts[len(parts)-1]
	}

	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return url
	}

	name := parts[len(parts)-1]

	return strings.TrimSuffix(name, ".git")
}

func registerWithPrompt(cmd *cobra.Command, registry *config.RepoRegistry, alias string, entry config.RegistryEntry) (string, error) {
	if registry == nil {
		return alias, fmt.Errorf("registry not configured")
	}

	target := strings.TrimSpace(alias)
	if target == "" {
		return "", fmt.Errorf("alias is required")
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

	if err := registry.Save(); err != nil {
		if unregErr := registry.Unregister(alias); unregErr != nil {
			return "", fmt.Errorf("failed to persist registry: %v (rollback failed: %v)", err, unregErr)
		}

		return "", fmt.Errorf("failed to persist registry: %w", err)
	}

	return alias, nil
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

	repoRemoveCmd.Flags().BoolP("force", "f", false, "Force removal even if used by active tickets")
	repoAddCmd.Flags().String("alias", "", "Override derived alias when auto-registering")
	repoAddCmd.Flags().Bool("no-register", false, "Skip auto-registration in the registry")
	repoRegisterCmd.Flags().Bool("force", false, "Overwrite existing alias if present")
	repoRegisterCmd.Flags().String("branch", "", "Default branch for the repository")
	repoRegisterCmd.Flags().String("description", "", "Description for the repository")
	repoRegisterCmd.Flags().String("tags", "", "Comma-separated tags for filtering")
	repoListRegistryCmd.Flags().String("tags", "", "Filter registry entries by comma-separated tags")
}
