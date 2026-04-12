package main

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/giturl"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

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

		if _, err := registerRegistryEntry(app.Config.GetRegistry(), alias, entry, force, app.Logger); err != nil {
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

		status, err := canonicalPathStatus(canonicalPath)
		if err != nil {
			return err
		}

		output.Infof("Canonical:    %s (%s)", canonicalPath, status)

		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoRegisterCmd)
	repoCmd.AddCommand(repoUnregisterCmd)
	repoCmd.AddCommand(repoListRegistryCmd)
	repoCmd.AddCommand(repoShowCmd)

	repoRegisterCmd.Flags().Bool("force", false, "Overwrite existing alias if present")
	repoRegisterCmd.Flags().String("branch", "", "Default branch for the repository")
	repoRegisterCmd.Flags().String("description", "", "Description for the repository")
	repoRegisterCmd.Flags().String("tags", "", "Comma-separated tags for filtering")
	repoListRegistryCmd.Flags().String("tags", "", "Filter registry entries by comma-separated tags")
}
