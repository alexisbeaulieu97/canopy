package main

import (
	"bufio"
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

const allWorkspacesPattern = ".*"

type hookPreviewEnvelope struct {
	DryRunHooks   bool                        `json:"dry_run_hooks"`
	Phase         string                      `json:"phase"`
	WorkspaceID   string                      `json:"workspace_id"`
	WorkspacePath string                      `json:"workspace_path,omitempty"`
	Commands      []domain.HookCommandPreview `json:"commands"`
	Action        string                      `json:"action,omitempty"`
	ClosedAt      *time.Time                  `json:"closed_at,omitempty"`
}

func validateWorkspaceTargetArgs(
	cmd *cobra.Command,
	args []string,
	singleCount int,
	singleField string,
	singleMessage string,
	bulkCount int,
	bulkField string,
	bulkMessage string,
) error {
	pattern, err := resolveWorkspaceBulkPattern(cmd)
	if err != nil {
		return err
	}

	if pattern != "" {
		if len(args) != bulkCount {
			return cerrors.NewInvalidArgument(bulkField, bulkMessage)
		}

		return nil
	}

	if len(args) != singleCount {
		return cerrors.NewInvalidArgument(singleField, singleMessage)
	}

	return nil
}

func resolveWorkspaceBulkPattern(cmd *cobra.Command) (string, error) {
	pattern, _ := cmd.Flags().GetString("pattern")
	all, _ := cmd.Flags().GetBool("all")

	if all && pattern != "" {
		return "", cerrors.NewInvalidArgument("pattern", "cannot use --pattern with --all")
	}

	if all {
		return allWorkspacesPattern, nil
	}

	return pattern, nil
}

func validateHookExecutionFlags(noHooks, hooksOnly, dryRunHooks bool) error {
	if hooksOnly && noHooks {
		return cerrors.NewInvalidArgument("flags", "cannot use --hooks-only with --no-hooks")
	}

	if dryRunHooks && noHooks {
		return cerrors.NewInvalidArgument("flags", "cannot use --dry-run-hooks with --no-hooks")
	}

	if dryRunHooks && hooksOnly {
		return cerrors.NewInvalidArgument("flags", "cannot use --dry-run-hooks with --hooks-only")
	}

	return nil
}

func newHookPreviewEnvelope(
	phase workspaces.HookPhase,
	workspaceID, workspacePath, action string,
	previews []domain.HookCommandPreview,
	closedAt *time.Time,
) hookPreviewEnvelope {
	return hookPreviewEnvelope{
		DryRunHooks:   true,
		Phase:         string(phase),
		WorkspaceID:   workspaceID,
		WorkspacePath: workspacePath,
		Commands:      previews,
		Action:        action,
		ClosedAt:      closedAt,
	}
}

func printHookPreviewJSON(
	phase workspaces.HookPhase,
	workspaceID, workspacePath, action string,
	previews []domain.HookCommandPreview,
	closedAt *time.Time,
) error {
	return output.PrintJSON(newHookPreviewEnvelope(phase, workspaceID, workspacePath, action, previews, closedAt))
}

func closeWithHookDryRunJSON(
	ctx context.Context,
	service *workspaces.Service,
	id string,
	force bool,
	keepMetadata bool,
	opts workspaces.CloseOptions,
	previews []domain.HookCommandPreview,
) error {
	if keepMetadata {
		archived, err := service.CloseWorkspaceKeepMetadataWithOptions(ctx, id, force, opts)
		if err != nil {
			return err
		}

		var closedAt *time.Time
		if archived != nil {
			closedAt = archived.Metadata.ClosedAt
		}

		return printHookPreviewJSON(workspaces.HookPhasePreClose, id, "", "close_keep", previews, closedAt)
	}

	if err := service.CloseWorkspaceWithOptions(ctx, id, force, opts); err != nil {
		return err
	}

	return printHookPreviewJSON(workspaces.HookPhasePreClose, id, "", "close_delete", previews, nil)
}

func resolveWorkspaceCloseKeepMetadata(defaultKeep, keepFlag, deleteFlag, interactive bool, reader *bufio.Reader) bool {
	switch {
	case keepFlag:
		return true
	case deleteFlag:
		return false
	case !interactive || reader == nil:
		return defaultKeep
	}

	promptSuffix := "[y/N]"
	if defaultKeep {
		promptSuffix = "[Y/n]"
	}

	output.Printf("Keep workspace record without files? %s: ", promptSuffix)

	answer, err := reader.ReadString('\n')
	if err != nil {
		return defaultKeep
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return defaultKeep
	}
}
