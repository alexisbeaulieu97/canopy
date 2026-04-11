package main

import (
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

// workspace_sync.go defines the "workspace sync" subcommand.

var workspaceSyncCmd = &cobra.Command{
	Use:   "sync [ID]",
	Short: "Pull updates for all repositories in a workspace",
	Long: `Pull updates for all repositories in a workspace and display a summary.
Per-repository timeouts can be configured to prevent slow remotes from blocking the entire operation.
Bulk sync continues across workspaces and exits non-zero if any workspace fails.`,
	Args: func(cmd *cobra.Command, args []string) error {
		pattern, _ := cmd.Flags().GetString("pattern")

		all, _ := cmd.Flags().GetBool("all")
		if all && pattern != "" {
			return cerrors.NewInvalidArgument("pattern", "cannot use --pattern with --all")
		}

		if pattern != "" || all {
			if len(args) != 0 {
				return cerrors.NewInvalidArgument("id", "cannot provide workspace ID with --pattern or --all")
			}

			return nil
		}

		if len(args) != 1 {
			return cerrors.NewInvalidArgument("id", "workspace ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timeoutStr, _ := cmd.Flags().GetString("timeout")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		pattern, _ := cmd.Flags().GetString("pattern")
		all, _ := cmd.Flags().GetBool("all")
		noProgress, _ := cmd.Flags().GetBool("no-progress")

		var timeout time.Duration

		if timeoutStr != "" {
			var err error

			timeout, err = time.ParseDuration(timeoutStr)
			if err != nil {
				return cerrors.NewInvalidArgument("timeout", fmt.Sprintf("invalid duration: %v", err))
			}
		}

		app, err := getApp(cmd)
		if err != nil {
			return err
		}

		opts := workspaces.SyncOptions{
			Timeout: timeout,
		}

		if all {
			pattern = ".*"
		}

		if pattern != "" {
			ids, err := resolveBulkWorkspaceIDs(cmd.Context(), app.Service, pattern)
			if err != nil {
				return err
			}

			if len(ids) == 0 {
				if jsonOutput {
					return output.PrintJSON([]map[string]any{})
				}

				output.Info("No matching workspaces found.")

				return nil
			}

			if !jsonOutput {
				printMatchedWorkspaceIDs(ids)
			}

			numWorkers := app.Config.GetParallelWorkers()
			if numWorkers <= 0 {
				numWorkers = 1
			}

			report := runBulkWorkspaceOperations(cmd.Context(), ids, bulkWorkspaceRunOptions[*domain.SyncResult]{
				Parallelism:  numWorkers,
				ShowProgress: !noProgress && !jsonOutput,
				OnSuccess: func(id string, current, total int, _ *domain.SyncResult) {
					if !jsonOutput {
						output.Infof("Synced workspace %s (%d/%d)", id, current, total)
					}
				},
				OnFailure: func(id string, current, total int, err error) {
					if !jsonOutput {
						output.Warnf("Workspace %s sync failed (%d/%d): %v", id, current, total, err)
					}
				},
			}, func(ctx context.Context, id string) (*domain.SyncResult, error) {
				return app.Service.SyncWorkspace(ctx, id, opts)
			})

			if report.Cancelled {
				if !jsonOutput {
					output.Warnf("Bulk sync cancelled")
				}

				return cerrors.NewOperationCancelled("bulk sync")
			}

			failed, totalUpdated := summarizeBulkSyncReport(report.Results)
			if jsonOutput {
				payload := make([]map[string]any, 0, len(report.Results))
				for _, res := range report.Results {
					errText := ""
					if res.Err != nil {
						errText = res.Err.Error()
					}

					payload = append(payload, map[string]interface{}{
						"workspace_id": res.ID,
						"result":       res.Value,
						"error":        errText,
					})
				}

				if err := output.PrintJSON(payload); err != nil {
					return err
				}

				if failed > 0 {
					return cerrors.NewCommandFailed("sync", fmt.Errorf("%d workspaces failed", failed))
				}

				return nil
			}

			// Render bulk sync results
			output.Println("")

			icons := output.NewIcons()

			for _, res := range report.Results {
				var icon string

				style := output.SuccessStyle
				statusText := ""

				if res.Err != nil {
					icon = icons.Error()
					style = output.ErrorStyle
					errDetail := sanitizeErrorForDisplay(res.Err.Error())
					statusText = "error: " + errDetail
				} else if res.Value != nil {
					if res.Value.TotalErrors > 0 {
						icon = icons.Warning()
						style = output.WarningStyle
						statusText = fmt.Sprintf("partial: %d updated, %d errors", res.Value.TotalUpdated, res.Value.TotalErrors)
					} else if res.Value.TotalUpdated > 0 {
						icon = icons.Success()
						statusText = fmt.Sprintf("pulled %d commits", res.Value.TotalUpdated)
					} else {
						icon = icons.Success()
						statusText = "already up-to-date"
					}
				}

				coloredIcon := output.Colorize(style, icon)
				_, _ = fmt.Fprintf(os.Stdout, "%s %-20s %s\n", coloredIcon, res.ID, statusText) //nolint:forbidigo // user-facing CLI output
			}

			// Summary
			output.Println("")
			output.Println(output.HorizontalRule(48))

			summaryParts := []string{
				fmt.Sprintf("%d workspaces synced", len(ids)),
			}

			if totalUpdated > 0 {
				summaryParts = append(summaryParts, fmt.Sprintf("%d commits pulled", totalUpdated))
			}

			if failed > 0 {
				summaryParts = append(summaryParts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%d failed", failed)))
			}

			successIcon := output.Colorize(output.SuccessStyle, icons.Success())
			if failed > 0 {
				successIcon = output.Colorize(output.WarningStyle, icons.Warning())
			}

			_, _ = fmt.Fprintf(os.Stdout, "%s Bulk sync complete: %s\n", successIcon, strings.Join(summaryParts, ", ")) //nolint:forbidigo // user-facing CLI output

			if failed > 0 {
				return cerrors.NewCommandFailed("sync", fmt.Errorf("%d workspaces failed", failed))
			}

			return nil
		}

		id := args[0]

		result, err := app.Service.SyncWorkspace(cmd.Context(), id, opts)
		if err != nil {
			return err
		}

		if jsonOutput {
			return output.PrintJSON(result)
		}

		renderSyncResult(id, result)

		if result.TotalErrors > 0 {
			return cerrors.NewCommandFailed("sync", fmt.Errorf("%d repos failed", result.TotalErrors))
		}

		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceSyncCmd)

	workspaceSyncCmd.Flags().String("timeout", "60s", "Timeout for each repository sync (e.g. 30s, 2m)")
	workspaceSyncCmd.Flags().Bool("json", false, "Output in JSON format")
	workspaceSyncCmd.Flags().String("pattern", "", "Sync workspaces matching a regex pattern")
	workspaceSyncCmd.Flags().Bool("all", false, "Sync all workspaces (equivalent to --pattern \".*\")")
	workspaceSyncCmd.Flags().Bool("no-progress", false, "Disable progress indicators")
}

func summarizeBulkSyncReport(results []bulkWorkspaceResult[*domain.SyncResult]) (failed, totalUpdated int) {
	for _, res := range results {
		if res.Err != nil {
			failed++
			continue
		}

		if res.Value == nil {
			continue
		}

		totalUpdated += res.Value.TotalUpdated
		if res.Value.TotalErrors > 0 {
			failed++
		}
	}

	return failed, totalUpdated
}

func renderSyncResult(workspaceID string, result *domain.SyncResult) {
	icons := output.NewIcons()

	output.Infof("Workspace: %s", workspaceID)
	output.Println("")

	// Print per-repo results
	for _, r := range result.Repos {
		var (
			icon, statusText string
			style            = output.SuccessStyle
		)

		switch r.Status {
		case domain.SyncStatusUpdated:
			icon = icons.Success()
			statusText = fmt.Sprintf("pulled %d commits", r.Updated)
		case domain.SyncStatusUpToDate:
			icon = icons.Success()
			statusText = "already up-to-date"
		case domain.SyncStatusConflict:
			icon = icons.Error()
			statusText = "merge conflict"
			style = output.ErrorStyle
		case domain.SyncStatusTimeout:
			icon = icons.Error()
			statusText = "timeout"
			style = output.ErrorStyle
		case domain.SyncStatusError:
			icon = icons.Error()
			errDetail := sanitizeErrorForDisplay(r.Error)
			statusText = "error: " + errDetail
			style = output.ErrorStyle
		default:
			icon = icons.Info()
			statusText = string(r.Status)
		}

		coloredIcon := output.Colorize(style, icon)
		_, _ = fmt.Fprintf(os.Stdout, "%s %-20s %s\n", coloredIcon, r.Name, statusText) //nolint:forbidigo // user-facing CLI output
	}

	// Print summary
	output.Println("")
	output.Println(output.HorizontalRule(48))

	summaryParts := []string{
		fmt.Sprintf("%d repos synced", len(result.Repos)),
	}

	if result.TotalUpdated > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d commits pulled", result.TotalUpdated))
	}

	if result.TotalErrors > 0 {
		summaryParts = append(summaryParts, output.Colorize(output.ErrorStyle, fmt.Sprintf("%d failed", result.TotalErrors)))
	}

	successIcon := output.Colorize(output.SuccessStyle, icons.Success())
	if result.TotalErrors > 0 {
		successIcon = output.Colorize(output.WarningStyle, icons.Warning())
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s Sync complete: %s\n", successIcon, strings.Join(summaryParts, ", ")) //nolint:forbidigo // user-facing CLI output
}

func sanitizeErrorForDisplay(errText string) string {
	errText = strings.ReplaceAll(errText, "\n", " ")
	errText = strings.ReplaceAll(errText, "\r", " ")
	errText = strings.ReplaceAll(errText, "\t", " ")

	runes := []rune(errText)
	if len(runes) > 50 {
		errText = string(runes[:47]) + "..."
	}

	return errText
}
