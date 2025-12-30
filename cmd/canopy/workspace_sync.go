package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
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
			matched, err := app.Service.ListWorkspacesMatching(cmd.Context(), pattern)
			if err != nil {
				return err
			}

			if len(matched) == 0 {
				output.Info("No matching workspaces found.")
				return nil
			}

			ids := make([]string, len(matched))
			for i, ws := range matched {
				ids[i] = ws.ID
			}

			output.Infof("Matched %d workspaces:", len(ids))
			for _, id := range ids {
				output.Infof("  - %s", id)
			}

			type syncJob struct {
				index int
				id    string
			}
			type syncResult struct {
				index  int
				id     string
				result *domain.SyncResult
				err    error
			}

			jobs := make(chan syncJob, len(ids))
			results := make(chan syncResult, len(ids))

			for i, id := range ids {
				jobs <- syncJob{index: i, id: id}
			}
			close(jobs)

			numWorkers := app.Config.GetParallelWorkers()
			if numWorkers <= 0 {
				numWorkers = 1
			}
			if numWorkers > len(ids) {
				numWorkers = len(ids)
			}

			var wg sync.WaitGroup
			for w := 0; w < numWorkers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for job := range jobs {
						result, syncErr := app.Service.SyncWorkspace(cmd.Context(), job.id, opts)
						results <- syncResult{
							index:  job.index,
							id:     job.id,
							result: result,
							err:    syncErr,
						}
					}
				}()
			}

			go func() {
				wg.Wait()
				close(results)
			}()

			// Set up progress tracking
			showProgress := !noProgress && !jsonOutput
			var progress *output.Progress
			if showProgress {
				progressOpts := output.DefaultProgressOptions(len(ids))
				progress = output.NewProgress(progressOpts)
			}

			orderedResults := make([]syncResult, len(ids))
			cancelled := false
			done := 0
			for res := range results {
				// Check for context cancellation
				if cmd.Context().Err() != nil && !cancelled {
					cancelled = true
					if progress != nil {
						progress.Cancel()
					}
				}

				done++
				orderedResults[res.index] = res

				// Always log errors for visibility
				if res.err != nil {
					output.Warnf("Workspace %s sync failed (%d/%d): %v", res.id, done, len(ids), res.err)
				}

				if showProgress && !cancelled {
					if res.err != nil {
						progress.Increment(fmt.Sprintf("%s (failed)", res.id))
					} else {
						progress.Increment(res.id)
					}
				} else if res.err == nil {
					// Log success when progress is off or after cancellation
					output.Infof("Synced workspace %s (%d/%d)", res.id, done, len(ids))
				}
			}

			if progress != nil {
				progress.Finish()
			}

			// Handle cancellation
			if cancelled {
				output.Warnf("Bulk sync cancelled")
				return cerrors.NewOperationCancelled("bulk sync")
			}

			if jsonOutput {
				payload := make([]map[string]interface{}, 0, len(orderedResults))
				for _, res := range orderedResults {
					errText := ""
					if res.err != nil {
						errText = res.err.Error()
					}
					payload = append(payload, map[string]interface{}{
						"workspace_id": res.id,
						"result":       res.result,
						"error":        errText,
					})
				}
				return output.PrintJSON(payload)
			}

			// Render bulk sync results
			output.Println("")
			icons := output.NewIcons()

			var failed int
			var totalUpdated int
			for _, res := range orderedResults {
				var icon string
				style := output.SuccessStyle
				statusText := ""

				if res.err != nil {
					icon = icons.Error()
					style = output.ErrorStyle
					errDetail := sanitizeErrorForDisplay(res.err.Error())
					statusText = "error: " + errDetail
					failed++
				} else if res.result != nil {
					totalUpdated += res.result.TotalUpdated
					if res.result.TotalErrors > 0 {
						icon = icons.Warning()
						style = output.WarningStyle
						statusText = fmt.Sprintf("partial: %d updated, %d errors", res.result.TotalUpdated, res.result.TotalErrors)
						failed++
					} else if res.result.TotalUpdated > 0 {
						icon = icons.Success()
						statusText = fmt.Sprintf("pulled %d commits", res.result.TotalUpdated)
					} else {
						icon = icons.Success()
						statusText = "already up-to-date"
					}
				}

				coloredIcon := output.Colorize(style, icon)
				_, _ = fmt.Fprintf(os.Stdout, "%s %-20s %s\n", coloredIcon, res.id, statusText) //nolint:forbidigo // user-facing CLI output
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
