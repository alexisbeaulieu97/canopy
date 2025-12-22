package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
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

			orderedResults := make([]syncResult, len(ids))
			done := 0
			for res := range results {
				done++
				orderedResults[res.index] = res
				if res.err != nil {
					output.Warnf("Workspace %s sync failed (%d/%d): %v", res.id, done, len(ids), res.err)
					continue
				}
				output.Infof("Synced workspace %s (%d/%d)", res.id, done, len(ids))
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

			output.Println("")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
			_, _ = fmt.Fprintln(w, "WORKSPACE\tSTATUS\tUPDATED\tERRORS\tDETAILS")

			var failed int
			var totalUpdated int
			for _, res := range orderedResults {
				status := "OK"
				updated := 0
				errorsCount := 0
				details := ""

				if res.err != nil {
					status = "ERROR"
					details = res.err.Error()
					failed++
				} else if res.result != nil {
					updated = res.result.TotalUpdated
					errorsCount = res.result.TotalErrors
					totalUpdated += updated
					if errorsCount > 0 {
						status = "PARTIAL"
						details = fmt.Sprintf("%d repo errors", errorsCount)
						failed++
					}
				}

				details = strings.ReplaceAll(details, "\n", " ")
				details = strings.ReplaceAll(details, "\r", " ")
				details = strings.ReplaceAll(details, "\t", " ")
				runes := []rune(details)
				if len(runes) > 100 {
					details = string(runes[:97]) + "..."
				}

				_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n", res.id, status, updated, errorsCount, details)
			}
			_ = w.Flush()

			output.Success("\nBulk sync completed", fmt.Sprintf("%d workspaces, %d total commits updated", len(ids), totalUpdated))
			if failed > 0 {
				output.Warnf("Bulk sync finished with %d failed workspaces", failed)
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

		output.Infof("Syncing workspace: %s", id)
		output.Println("")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		_, _ = fmt.Fprintln(w, "REPOSITORY\tSTATUS\tUPDATED\tDETAILS")

		for _, r := range result.Repos {
			status := strings.ToUpper(string(r.Status))
			updatedStr := fmt.Sprintf("%d commits", r.Updated)
			if r.Updated == 0 {
				updatedStr = "-"
			}

			// Sanitize error message to prevent breaking tabwriter layout.
			errDetail := r.Error
			errDetail = strings.ReplaceAll(errDetail, "\n", " ")
			errDetail = strings.ReplaceAll(errDetail, "\r", " ")
			errDetail = strings.ReplaceAll(errDetail, "\t", " ")
			runes := []rune(errDetail)
			if len(runes) > 100 {
				errDetail = string(runes[:97]) + "..."
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, status, updatedStr, errDetail)
		}
		_ = w.Flush()

		if result.TotalErrors > 0 {
			output.Warnf("\nPartial failure: %d repositories failed to sync", result.TotalErrors)
			return cerrors.NewCommandFailed("sync", fmt.Errorf("%d repos failed", result.TotalErrors))
		}

		output.Success("\nWorkspace sync completed", fmt.Sprintf("%d total commits updated", result.TotalUpdated))
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceSyncCmd)

	workspaceSyncCmd.Flags().String("timeout", "60s", "Timeout for each repository sync (e.g. 30s, 2m)")
	workspaceSyncCmd.Flags().Bool("json", false, "Output in JSON format")
	workspaceSyncCmd.Flags().String("pattern", "", "Sync workspaces matching a regex pattern")
	workspaceSyncCmd.Flags().Bool("all", false, "Sync all workspaces (equivalent to --pattern \".*\")")
}
