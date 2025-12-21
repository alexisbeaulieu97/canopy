package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// template.go defines commands for workspace templates.

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage workspace templates",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available workspace templates",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		templates := app.Config.GetTemplates()
		if len(templates) == 0 {
			output.Info("No templates defined.")
			return nil
		}

		names := make([]string, 0, len(templates))
		for name := range templates {
			names = append(names, name)
		}
		sort.Strings(names)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tREPOS\tDESCRIPTION")
		for _, name := range names {
			tmpl := templates[name]
			repos := strings.Join(tmpl.Repos, ", ")
			description := tmpl.Description
			if description == "" {
				description = "-"
			}
			_, _ = fmt.Fprintf(w, "\033[1;36m%s\033[0m\t%s\t%s\n", name, repos, description)
		}
		_ = w.Flush()

		return nil
	},
}

var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		tmpl, err := app.Config.ResolveTemplate(args[0])
		if err != nil {
			return err
		}

		canonicalRepos, err := app.Service.ListCanonicalRepos(cmd.Context())
		if err != nil {
			return err
		}

		canonicalSet := make(map[string]bool, len(canonicalRepos))
		for _, name := range canonicalRepos {
			canonicalSet[name] = true
		}

		output.Infof("Template: %s", tmpl.Name)
		if tmpl.Description != "" {
			output.Infof("Description: %s", tmpl.Description)
		} else {
			output.Infof("Description: -")
		}
		if tmpl.DefaultBranch != "" {
			output.Infof("Default branch: %s", tmpl.DefaultBranch)
		}

		if len(tmpl.Repos) == 0 {
			output.Info("Repos: -")
		} else {
			output.Info("Repos:")
			for _, repo := range tmpl.Repos {
				registryAvailable := false
				if registry := app.Config.GetRegistry(); registry != nil {
					_, registryAvailable = registry.Resolve(repo)
				}

				canonicalAvailable := canonicalSet[repo]
				output.Infof("  - %s (registry: %t, canonical: %t)", repo, registryAvailable, canonicalAvailable)
			}
		}

		if len(tmpl.SetupCommands) > 0 {
			output.Info("Setup commands:")
			for _, command := range tmpl.SetupCommands {
				output.Infof("  - %s", command)
			}
		}

		return nil
	},
}

var templateValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate template definitions",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		if err := app.Config.ValidateTemplates(); err != nil {
			return err
		}

		for name, tmpl := range app.Config.GetTemplates() {
			if _, err := app.Service.ResolveRepos(name, tmpl.Repos); err != nil {
				return err
			}
		}

		output.Info("Templates are valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateValidateCmd)
}
