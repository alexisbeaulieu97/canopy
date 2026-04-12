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

type workspaceCloseOptions struct {
	force                bool
	keepFlag             bool
	deleteFlag           bool
	dryRun               bool
	jsonOutput           bool
	noHooks              bool
	hooksOnly            bool
	dryRunHooks          bool
	noProgress           bool
	pattern              string
	configDefaultArchive bool
	interactive          bool
	closeOpts            workspaces.CloseOptions
}

func runWorkspaceClose(cmd *cobra.Command, args []string) error {
	opts, err := resolveWorkspaceCloseOptions(cmd)
	if err != nil {
		return err
	}

	app, err := getApp(cmd)
	if err != nil {
		return err
	}

	opts.configDefaultArchive = strings.EqualFold(app.Config.GetCloseDefault(), "archive")
	opts.interactive = isInteractiveTerminal()

	if opts.pattern != "" {
		return runBulkWorkspaceClose(cmd, app.Service, opts)
	}

	return runSingleWorkspaceClose(cmd, args[0], app.Service, opts)
}

func resolveWorkspaceCloseOptions(cmd *cobra.Command) (workspaceCloseOptions, error) {
	force, _ := cmd.Flags().GetBool("force")
	keepFlag, _ := cmd.Flags().GetBool("keep")
	deleteFlag, _ := cmd.Flags().GetBool("delete")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	noHooks, _ := cmd.Flags().GetBool("no-hooks")
	hooksOnly, _ := cmd.Flags().GetBool("hooks-only")
	dryRunHooks, _ := cmd.Flags().GetBool("dry-run-hooks")
	noProgress, _ := cmd.Flags().GetBool("no-progress")

	pattern, err := resolveWorkspaceBulkPattern(cmd)
	if err != nil {
		return workspaceCloseOptions{}, err
	}

	if keepFlag && deleteFlag {
		return workspaceCloseOptions{}, cerrors.NewInvalidArgument("flags", "cannot use --keep and --delete together")
	}

	if err := validateHookExecutionFlags(noHooks, hooksOnly, dryRunHooks); err != nil {
		return workspaceCloseOptions{}, err
	}

	if dryRunHooks && dryRun {
		return workspaceCloseOptions{}, cerrors.NewInvalidArgument("flags", "cannot use --dry-run-hooks with --dry-run")
	}

	return workspaceCloseOptions{
		force:       force,
		keepFlag:    keepFlag,
		deleteFlag:  deleteFlag,
		dryRun:      dryRun,
		jsonOutput:  jsonOutput,
		noHooks:     noHooks,
		hooksOnly:   hooksOnly,
		dryRunHooks: dryRunHooks,
		noProgress:  noProgress,
		pattern:     pattern,
		closeOpts: workspaces.CloseOptions{
			SkipHooks: noHooks || dryRunHooks,
		},
	}, nil
}

func runBulkWorkspaceClose(cmd *cobra.Command, service *workspaces.Service, opts workspaceCloseOptions) error {
	if opts.hooksOnly || opts.dryRunHooks {
		return cerrors.NewInvalidArgument("flags", "--hooks-only and --dry-run-hooks require a single workspace ID")
	}

	keepMetadata := resolveWorkspaceCloseKeepMetadata(opts.configDefaultArchive, opts.keepFlag, opts.deleteFlag, false, false, nil)

	ids, err := resolveBulkWorkspaceIDs(cmd.Context(), service, opts.pattern)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		output.Info("No matching workspaces found.")
		return nil
	}

	printMatchedWorkspaceIDs(ids)

	if opts.dryRun {
		return previewBulkWorkspaceClose(cmd.Context(), service, ids, keepMetadata, opts.jsonOutput)
	}

	if err := confirmBulkWorkspaceAction(opts.force, opts.interactive, len(ids), "Bulk close"); err != nil {
		return err
	}

	report := runBulkWorkspaceOperations(cmd.Context(), ids, bulkWorkspaceRunOptions[struct{}]{
		Parallelism:  1,
		ShowProgress: !opts.noProgress && !opts.jsonOutput,
		OnStart: func(id string, current, total int) {
			output.Infof("Closing workspace %s (%d/%d)", id, current, total)
		},
		OnFailure: func(id string, _, _ int, err error) {
			output.Warnf("Failed to close workspace %s: %v", id, err)
		},
	}, func(ctx context.Context, id string) (struct{}, error) {
		if keepMetadata {
			_, err := service.CloseWorkspaceKeepMetadataWithOptions(ctx, id, opts.force, opts.closeOpts)
			return struct{}{}, err
		}

		return struct{}{}, service.CloseWorkspaceWithOptions(ctx, id, opts.force, opts.closeOpts)
	})

	return finishBulkWorkspaceClose(ids, report)
}

func previewBulkWorkspaceClose(ctx context.Context, service *workspaces.Service, ids []string, keepMetadata, jsonOutput bool) error {
	previews := make([]*domain.WorkspaceClosePreview, 0, len(ids))
	for _, id := range ids {
		preview, err := service.PreviewCloseWorkspace(ctx, id, keepMetadata)
		if err != nil {
			return err
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

func finishBulkWorkspaceClose(ids []string, report bulkWorkspaceRunReport[struct{}]) error {
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

func runSingleWorkspaceClose(cmd *cobra.Command, id string, service *workspaces.Service, opts workspaceCloseOptions) error {
	if opts.hooksOnly {
		return runWorkspaceCloseHooksOnly(cmd.Context(), service, id, opts)
	}

	var (
		hookPreviews []domain.HookCommandPreview
		err          error
	)
	if opts.dryRunHooks {
		hookPreviews, err = service.PreviewHooks(cmd.Context(), id, workspaces.HookPhasePreClose)
		if err != nil {
			return err
		}
	}

	if opts.dryRun {
		return previewSingleWorkspaceClose(cmd.Context(), service, id, opts)
	}

	keepMetadata := resolveWorkspaceCloseKeepMetadata(
		opts.configDefaultArchive,
		opts.keepFlag,
		opts.deleteFlag,
		false,
		opts.interactive,
		newWorkspaceClosePromptReader(opts),
	)

	if opts.dryRunHooks {
		if opts.jsonOutput {
			return closeWithHookDryRunJSON(cmd.Context(), service, id, opts.force, keepMetadata, opts.closeOpts, hookPreviews)
		}

		printHookPreview(string(workspaces.HookPhasePreClose), hookPreviews)
	}

	if keepMetadata {
		return keepAndPrint(cmd.Context(), service, id, opts.force, opts.closeOpts)
	}

	return closeAndPrint(cmd.Context(), service, id, opts.force, opts.closeOpts)
}

func runWorkspaceCloseHooksOnly(ctx context.Context, service *workspaces.Service, id string, opts workspaceCloseOptions) error {
	if opts.keepFlag || opts.deleteFlag {
		return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --keep or --delete")
	}

	if opts.dryRun {
		return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --dry-run")
	}

	if opts.jsonOutput {
		return cerrors.NewInvalidArgument("flags", "--hooks-only cannot be combined with --json")
	}

	if err := service.RunHooks(ctx, id, workspaces.HookPhasePreClose, false); err != nil {
		return err
	}

	output.Success("Ran pre_close hooks for workspace", id)

	return nil
}

func previewSingleWorkspaceClose(ctx context.Context, service *workspaces.Service, id string, opts workspaceCloseOptions) error {
	keepMetadata := resolveWorkspaceCloseKeepMetadata(
		opts.configDefaultArchive,
		opts.keepFlag,
		opts.deleteFlag,
		true,
		false,
		nil,
	)

	preview, err := service.PreviewCloseWorkspace(ctx, id, keepMetadata)
	if err != nil {
		return err
	}

	if opts.jsonOutput {
		return output.PrintJSON(map[string]interface{}{
			"dry_run": true,
			"preview": preview,
		})
	}

	printWorkspaceClosePreview(preview)

	return nil
}

func newWorkspaceClosePromptReader(opts workspaceCloseOptions) *bufio.Reader {
	if !opts.interactive || opts.keepFlag || opts.deleteFlag {
		return nil
	}

	return bufio.NewReader(os.Stdin)
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
