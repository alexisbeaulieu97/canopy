package main

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// workspace_new.go defines the "workspace new" subcommand.

var workspaceNewCmd = &cobra.Command{
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
		dryRunHooks, _ := cmd.Flags().GetBool("dry-run-hooks")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		templateName, _ := cmd.Flags().GetString("template")

		if hooksOnly && noHooks {
			return cerrors.NewInvalidArgument("flags", "cannot use --hooks-only with --no-hooks")
		}

		if dryRunHooks && noHooks {
			return cerrors.NewInvalidArgument("flags", "cannot use --dry-run-hooks with --no-hooks")
		}

		if dryRunHooks && hooksOnly {
			return cerrors.NewInvalidArgument("flags", "cannot use --dry-run-hooks with --hooks-only")
		}

		if jsonOutput && !dryRunHooks {
			return cerrors.NewInvalidArgument("flags", "--json is only supported with --dry-run-hooks")
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

		var templateRepos []string
		var templatePtr *config.Template

		if templateName != "" {
			template, err := cfg.ResolveTemplate(templateName)
			if err != nil {
				return err
			}
			templateRepos = template.Repos
			templatePtr = &template
			if branch == "" && template.DefaultBranch != "" {
				branch = template.DefaultBranch
			}
		}

		// Resolve repos.
		var resolvedRepos []domain.Repo
		mergedRepos := mergeTemplateRepos(templateRepos, repos)
		if len(mergedRepos) > 0 {
			resolvedRepos, err = service.ResolveRepos(id, mergedRepos)
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
			SkipHooks: noHooks || dryRunHooks,
			Template:  templatePtr,
		}

		dirName, err := service.CreateWorkspaceWithOptions(cmd.Context(), id, branch, resolvedRepos, opts)
		if err != nil {
			return err
		}

		workspacePath := filepath.Join(cfg.GetWorkspacesRoot(), dirName)

		if dryRunHooks {
			previews, err := service.PreviewHooks(id, workspaces.HookPhasePostCreate)
			if err != nil {
				return err
			}

			if jsonOutput {
				return output.PrintJSON(hookPreviewEnvelope{
					DryRunHooks:   true,
					Phase:         string(workspaces.HookPhasePostCreate),
					WorkspaceID:   id,
					WorkspacePath: workspacePath,
					Commands:      previews,
					Action:        "create",
				})
			}

			printHookPreview(string(workspaces.HookPhasePostCreate), previews)
		}

		if printPath {
			output.Printf("%s", workspacePath)
		} else {
			output.SuccessWithPath("Created workspace", id, workspacePath)
		}
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceNewCmd)

	workspaceNewCmd.Flags().StringSlice("repos", []string{}, "List of repositories to include")
	workspaceNewCmd.Flags().String("branch", "", "Custom branch name (optional)")
	workspaceNewCmd.Flags().Bool("print-path", false, "Print the created workspace path to stdout")
	workspaceNewCmd.Flags().Bool("no-hooks", false, "Skip post_create hooks")
	workspaceNewCmd.Flags().Bool("hooks-only", false, "Run post_create hooks without creating the workspace")
	workspaceNewCmd.Flags().Bool("dry-run-hooks", false, "Preview post_create hooks without executing them")
	workspaceNewCmd.Flags().Bool("json", false, "Output in JSON format (use with --dry-run-hooks)")
	workspaceNewCmd.Flags().String("template", "", "Workspace template to apply")
}

func mergeTemplateRepos(templateRepos, explicitRepos []string) []string {
	seen := make(map[string]bool)

	var merged []string

	for _, repo := range templateRepos {
		repo = strings.TrimSpace(repo)
		if repo == "" || seen[repo] {
			continue
		}

		seen[repo] = true
		merged = append(merged, repo)
	}

	for _, repo := range explicitRepos {
		repo = strings.TrimSpace(repo)
		if repo == "" || seen[repo] {
			continue
		}

		seen[repo] = true
		merged = append(merged, repo)
	}

	return merged
}
