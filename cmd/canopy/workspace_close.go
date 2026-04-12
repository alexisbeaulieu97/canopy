package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

// workspace_close.go defines the "workspace close" subcommand.

var workspaceCloseCmd = &cobra.Command{
	Use:   "close [ID]",
	Short: "Close a workspace (keep metadata or delete)",
	Args: func(cmd *cobra.Command, args []string) error {
		return validateWorkspaceTargetArgs(
			cmd,
			args,
			1,
			"id",
			"workspace ID is required",
			0,
			"id",
			"cannot provide workspace ID with --pattern or --all",
		)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		keepFlag, _ := cmd.Flags().GetBool("keep")
		deleteFlag, _ := cmd.Flags().GetBool("delete")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		noHooks, _ := cmd.Flags().GetBool("no-hooks")
		hooksOnly, _ := cmd.Flags().GetBool("hooks-only")
		dryRunHooks, _ := cmd.Flags().GetBool("dry-run-hooks")

		pattern, err := resolveWorkspaceBulkPattern(cmd)
		if err != nil {
			return err
		}

		noProgress, _ := cmd.Flags().GetBool("no-progress")

		if keepFlag && deleteFlag {
			return cerrors.NewInvalidArgument("flags", "cannot use --keep and --delete together")
		}

		if err := validateHookExecutionFlags(noHooks, hooksOnly, dryRunHooks); err != nil {
			return err
		}

		if dryRunHooks && dryRun {
			return cerrors.NewInvalidArgument("flags", "cannot use --dry-run-hooks with --dry-run")
		}

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		service := app.Service
		configDefaultArchive := strings.EqualFold(app.Config.GetCloseDefault(), "archive")
		interactive := isInteractiveTerminal()

		closeOpts := workspaces.CloseOptions{
			SkipHooks: noHooks || dryRunHooks,
		}

		if pattern != "" {
			if hooksOnly || dryRunHooks {
				return cerrors.NewInvalidArgument("flags", "--hooks-only and --dry-run-hooks require a single workspace ID")
			}

			keepMetadata := resolveWorkspaceCloseKeepMetadata(configDefaultArchive, keepFlag, deleteFlag, false, false, nil)

			ids, err := resolveBulkWorkspaceIDs(cmd.Context(), service, pattern)
			if err != nil {
				return err
			}

			if len(ids) == 0 {
				output.Info("No matching workspaces found.")
				return nil
			}

			printMatchedWorkspaceIDs(ids)

			if dryRun {
				previews := make([]*domain.WorkspaceClosePreview, 0, len(ids))
				for _, id := range ids {
					preview, previewErr := service.PreviewCloseWorkspace(cmd.Context(), id, keepMetadata)
					if previewErr != nil {
						return previewErr
					}

					previews = append(previews, preview)
				}

				if jsonOutput {
					return output.PrintJSON(map[string]interface{}{
						"dry_run": true,
						"preview": previews,
					})
				}

				for _, preview := range previews {
					printWorkspaceClosePreview(preview)
					output.Println("")
				}

				return nil
			}

			if err := confirmBulkWorkspaceAction(force, interactive, len(ids), "Bulk close"); err != nil {
				return err
			}

			report := runBulkWorkspaceOperations(cmd.Context(), ids, bulkWorkspaceRunOptions[struct{}]{
				Parallelism:  1,
				ShowProgress: !noProgress && !jsonOutput,
				OnStart: func(id string, current, total int) {
					output.Infof("Closing workspace %s (%d/%d)", id, current, total)
				},
				OnFailure: func(id string, _, _ int, err error) {
					output.Warnf("Failed to close workspace %s: %v", id, err)
				},
			}, func(ctx context.Context, id string) (struct{}, error) {
				if keepMetadata {
					_, err := service.CloseWorkspaceKeepMetadataWithOptions(ctx, id, force, closeOpts)
					return struct{}{}, err
				}

				return struct{}{}, service.CloseWorkspaceWithOptions(ctx, id, force, closeOpts)
			})

			if report.Cancelled {
				skipped := len(ids) - len(report.SuccessIDs) - len(report.FailedIDs)
				output.Warnf("Bulk close cancelled: %d succeeded, %d failed, %d skipped", len(report.SuccessIDs), len(report.FailedIDs), skipped)

				return cerrors.NewOperationCancelled("bulk close")
			}

			output.Success("Bulk close completed", fmt.Sprintf("%d succeeded, %d failed", len(report.SuccessIDs), len(report.FailedIDs)))

			if len(report.FailedIDs) > 0 {
				output.Warnf("Failed workspaces: %s", strings.Join(report.FailedIDs, ", "))
				return cerrors.NewCommandFailed("bulk close", report.FirstErr)
			}

			return nil
		}

		id := args[0]

		if hooksOnly {
			if keepFlag || deleteFlag {
				return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --keep or --delete")
			}

			if dryRun {
				return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --dry-run")
			}

			if jsonOutput {
				return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --json")
			}

			if err := service.RunHooks(cmd.Context(), id, workspaces.HookPhasePreClose, false); err != nil {
				return err
			}

			output.Success("Ran pre_close hooks for workspace", id)

			return nil
		}

		var hookPreviews []domain.HookCommandPreview
		if dryRunHooks {
			hookPreviews, err = service.PreviewHooks(cmd.Context(), id, workspaces.HookPhasePreClose)
			if err != nil {
				return err
			}
		}

		// Handle dry-run mode.
		if dryRun {
			keepMetadata := resolveWorkspaceCloseKeepMetadata(
				configDefaultArchive,
				keepFlag,
				deleteFlag,
				true,
				false,
				nil,
			)

			preview, err := service.PreviewCloseWorkspace(cmd.Context(), id, keepMetadata)
			if err != nil {
				return err
			}

			if jsonOutput {
				return output.PrintJSON(map[string]interface{}{
					"dry_run": true,
					"preview": preview,
				})
			}

			printWorkspaceClosePreview(preview)

			return nil
		}

		var promptReader *bufio.Reader
		if interactive && !keepFlag && !deleteFlag {
			promptReader = bufio.NewReader(os.Stdin)
		}

		keepMetadata := resolveWorkspaceCloseKeepMetadata(
			configDefaultArchive,
			keepFlag,
			deleteFlag,
			false,
			interactive,
			promptReader,
		)

		if dryRunHooks && !jsonOutput {
			printHookPreview(string(workspaces.HookPhasePreClose), hookPreviews)
		}

		if dryRunHooks && jsonOutput {
			return closeWithHookDryRunJSON(cmd.Context(), service, id, force, keepMetadata, closeOpts, hookPreviews)
		}

		if keepMetadata {
			return keepAndPrint(cmd.Context(), service, id, force, closeOpts)
		}

		return closeAndPrint(cmd.Context(), service, id, force, closeOpts)
	},
}

func keepAndPrint(ctx context.Context, service *workspaces.Service, id string, force bool, opts workspaces.CloseOptions) error {
	archived, err := service.CloseWorkspaceKeepMetadataWithOptions(ctx, id, force, opts)
	if err != nil {
		return err
	}

	var archivedAt *time.Time
	if archived != nil {
		archivedAt = archived.Metadata.ClosedAt
	}

	printClosed(id, archivedAt)

	return nil
}

func closeAndPrint(ctx context.Context, service *workspaces.Service, id string, force bool, opts workspaces.CloseOptions) error {
	if err := service.CloseWorkspaceWithOptions(ctx, id, force, opts); err != nil {
		return err
	}

	output.Success("Closed workspace", id)

	return nil
}

func isInteractiveTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func init() {
	workspaceCmd.AddCommand(workspaceCloseCmd)

	workspaceCloseCmd.Flags().Bool("force", false, "Force close even if there are uncommitted changes")
	workspaceCloseCmd.Flags().Bool("keep", false, "Keep metadata (close without deleting)")
	workspaceCloseCmd.Flags().Bool("delete", false, "Delete without keeping metadata")
	workspaceCloseCmd.Flags().Bool("dry-run", false, "Preview what would be deleted without actually deleting")
	workspaceCloseCmd.Flags().Bool("json", false, "Output in JSON format (use with --dry-run)")
	workspaceCloseCmd.Flags().Bool("no-hooks", false, "Skip pre_close hooks")
	workspaceCloseCmd.Flags().Bool("hooks-only", false, "Run pre_close hooks without closing the workspace")
	workspaceCloseCmd.Flags().Bool("dry-run-hooks", false, "Preview pre_close hooks without executing them")
	workspaceCloseCmd.Flags().String("pattern", "", "Close workspaces matching a regex pattern")
	workspaceCloseCmd.Flags().Bool("all", false, "Close all workspaces (equivalent to --pattern \".*\")")
	workspaceCloseCmd.Flags().Bool("no-progress", false, "Disable progress indicators")
}
