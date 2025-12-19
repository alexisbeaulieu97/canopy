package main

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

type hookPreviewEnvelope struct {
	DryRunHooks   bool                        `json:"dry_run_hooks"`
	Phase         string                      `json:"phase"`
	WorkspaceID   string                      `json:"workspace_id"`
	WorkspacePath string                      `json:"workspace_path,omitempty"`
	Commands      []domain.HookCommandPreview `json:"commands"`
	Action        string                      `json:"action,omitempty"`
	ClosedAt      *time.Time                  `json:"closed_at,omitempty"`
}

var (
	hooksCmd = &cobra.Command{
		Use:   "hooks",
		Short: "Inspect lifecycle hooks",
	}

	hooksListCmd = &cobra.Command{
		Use:   "list",
		Short: "List configured hooks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			hooksConfig := app.Config.GetHooks()
			if len(hooksConfig.PostCreate) == 0 && len(hooksConfig.PreClose) == 0 {
				output.Info("No hooks configured.")
				return nil
			}

			printHookList("post_create", hooksConfig.PostCreate)
			printHookList("pre_close", hooksConfig.PreClose)

			return nil
		},
	}

	hooksTestCmd = &cobra.Command{
		Use:   "test <event>",
		Short: "Dry-run a hook event for a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			event := args[0]
			workspaceID, _ := cmd.Flags().GetString("workspace")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if workspaceID == "" {
				return cerrors.NewInvalidArgument("workspace", "workspace ID is required")
			}

			phase, err := parseHookPhase(event)
			if err != nil {
				return err
			}

			app, err := getApp(cmd)
			if err != nil {
				return err
			}

			previews, err := app.Service.PreviewHooks(workspaceID, phase)
			if err != nil {
				return err
			}

			if jsonOutput {
				return output.PrintJSON(hookPreviewEnvelope{
					DryRunHooks: true,
					Phase:       string(phase),
					WorkspaceID: workspaceID,
					Commands:    previews,
					Action:      "test",
				})
			}

			printHookPreview(string(phase), previews)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(hooksCmd)
	hooksCmd.AddCommand(hooksListCmd)
	hooksCmd.AddCommand(hooksTestCmd)

	hooksTestCmd.Flags().String("workspace", "", "Workspace ID to use for hook context")
	hooksTestCmd.Flags().Bool("json", false, "Output in JSON format")
}

func parseHookPhase(event string) (workspaces.HookPhase, error) {
	normalized := strings.ToLower(strings.ReplaceAll(event, "-", "_"))

	switch normalized {
	case string(workspaces.HookPhasePostCreate):
		return workspaces.HookPhasePostCreate, nil
	case string(workspaces.HookPhasePreClose):
		return workspaces.HookPhasePreClose, nil
	default:
		return "", cerrors.NewInvalidArgument("hook_event", "must be post_create or pre_close")
	}
}

func printHookList(name string, hooks []config.Hook) {
	if len(hooks) == 0 {
		output.Infof("%s: (none)", name)
		return
	}

	output.Infof("%s:", name)

	for i, hook := range hooks {
		output.Infof("  [%d] %s", i, hook.Command)

		if len(hook.Repos) > 0 {
			output.Infof("      repos: %s", strings.Join(hook.Repos, ", "))
		}

		if hook.Shell != "" {
			output.Infof("      shell: %s", hook.Shell)
		}

		if hook.Timeout > 0 {
			output.Infof("      timeout: %ds", hook.Timeout)
		}

		if hook.ContinueOnError {
			output.Info("      continue_on_error: true")
		}
	}
}

func printHookPreview(phase string, previews []domain.HookCommandPreview) {
	if len(previews) == 0 {
		output.Infof("No %s hooks configured.", phase)
		return
	}

	output.Infof("Hook dry-run (%s):", phase)

	for _, preview := range previews {
		output.Infof("  [%d] %s", preview.Index, preview.Command)
		output.Infof("      workspace: %s (branch: %s)", preview.WorkspaceID, preview.BranchName)

		if preview.RepoName != "" {
			output.Infof("      repo: %s", preview.RepoName)
		}

		output.Infof("      working_dir: %s", preview.WorkingDir)
	}
}

func closeWithHookPreviewJSON(
	service *workspaces.Service,
	id string,
	force bool,
	keepMetadata bool,
	opts workspaces.CloseOptions,
	previews []domain.HookCommandPreview,
) error {
	if keepMetadata {
		archived, err := service.CloseWorkspaceKeepMetadataWithOptions(id, force, opts)
		if err != nil {
			return err
		}

		var closedAt *time.Time
		if archived != nil {
			closedAt = archived.Metadata.ClosedAt
		}

		return output.PrintJSON(hookPreviewEnvelope{
			DryRunHooks: true,
			Phase:       string(workspaces.HookPhasePreClose),
			WorkspaceID: id,
			Commands:    previews,
			Action:      "close_keep",
			ClosedAt:    closedAt,
		})
	}

	if err := service.CloseWorkspaceWithOptions(id, force, opts); err != nil {
		return err
	}

	return output.PrintJSON(hookPreviewEnvelope{
		DryRunHooks: true,
		Phase:       string(workspaces.HookPhasePreClose),
		WorkspaceID: id,
		Commands:    previews,
		Action:      "close_delete",
	})
}
