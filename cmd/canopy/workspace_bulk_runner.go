package main

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/output"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

type bulkWorkspaceResult[T any] struct {
	ID    string
	Value T
	Err   error
}

type bulkWorkspaceRunOptions[T any] struct {
	Parallelism int
	ShowProgress bool
	OnStart      func(id string, current, total int)
	OnSuccess    func(id string, current, total int, value T)
	OnFailure    func(id string, current, total int, err error)
	ProgressLabel func(id string, value T, err error) string
}

type bulkWorkspaceRunReport[T any] struct {
	Results    []bulkWorkspaceResult[T]
	SuccessIDs []string
	FailedIDs  []string
	FirstErr   error
	Cancelled  bool
}

func resolveBulkWorkspaceIDs(ctx context.Context, service *workspaces.Service, pattern string) ([]string, error) {
	matched, err := service.ListWorkspacesMatching(ctx, pattern)
	if err != nil {
		return nil, err
	}

	if len(matched) == 0 {
		return nil, nil
	}

	ids := make([]string, len(matched))
	for i, ws := range matched {
		ids[i] = ws.ID
	}

	return ids, nil
}

func printMatchedWorkspaceIDs(ids []string) {
	output.Infof("Matched %d workspaces:", len(ids))
	for _, id := range ids {
		output.Infof("  - %s", id)
	}
}

func confirmBulkWorkspaceAction(force, interactive bool, count int, operation string) error {
	if force {
		return nil
	}

	if !interactive {
		return cerrors.NewInvalidArgument("force", operation+" requires confirmation; rerun with --force")
	}

	reader := bufio.NewReader(os.Stdin)
	output.Printf("%s %d workspaces? [y/N]: ", operation, count)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return cerrors.NewOperationCancelled(strings.ToLower(operation))
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" && answer != "yes" {
		return cerrors.NewOperationCancelled(strings.ToLower(operation))
	}

	return nil
}

func runBulkWorkspaceOperations[T any](
	ctx context.Context,
	ids []string,
	opts bulkWorkspaceRunOptions[T],
	execute func(context.Context, string) (T, error),
) bulkWorkspaceRunReport[T] {
	report := bulkWorkspaceRunReport[T]{
		Results: make([]bulkWorkspaceResult[T], len(ids)),
	}
	if len(ids) == 0 {
		return report
	}

	parallelism := opts.Parallelism
	if parallelism <= 0 || parallelism > len(ids) {
		parallelism = len(ids)
	}

	var progress *output.Progress
	if opts.ShowProgress {
		progress = output.NewProgress(output.DefaultProgressOptions(len(ids)))
	}

	record := func(index, current int, id string, value T, err error) {
		report.Results[index] = bulkWorkspaceResult[T]{ID: id, Value: value, Err: err}

		if err != nil {
			if report.FirstErr == nil {
				report.FirstErr = err
			}
			report.FailedIDs = append(report.FailedIDs, id)
			if opts.OnFailure != nil {
				opts.OnFailure(id, current, len(ids), err)
			}
		} else {
			report.SuccessIDs = append(report.SuccessIDs, id)
			if opts.OnSuccess != nil && !opts.ShowProgress {
				opts.OnSuccess(id, current, len(ids), value)
			}
		}

		if progress != nil && !report.Cancelled {
			label := id
			if opts.ProgressLabel != nil {
				label = opts.ProgressLabel(id, value, err)
			} else if err != nil {
				label = id + " (failed)"
			}
			progress.Increment(label)
		}
	}

	if parallelism == 1 {
		for i, id := range ids {
			if ctx.Err() != nil {
				report.Cancelled = true
				if progress != nil {
					progress.Cancel()
				}
				break
			}

			if opts.OnStart != nil && !opts.ShowProgress {
				opts.OnStart(id, i+1, len(ids))
			}

			value, err := execute(ctx, id)
			record(i, i+1, id, value, err)
		}
	} else {
		type job struct {
			index int
			id    string
		}
		type result struct {
			index int
			id    string
			value T
			err   error
		}

		jobs := make(chan job, len(ids))
		results := make(chan result, len(ids))

		for i, id := range ids {
			jobs <- job{index: i, id: id}
		}
		close(jobs)

		var wg sync.WaitGroup
		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					value, err := execute(ctx, job.id)
					results <- result{index: job.index, id: job.id, value: value, err: err}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		completed := 0
		for res := range results {
			completed++
			if ctx.Err() != nil && !report.Cancelled {
				report.Cancelled = true
				if progress != nil {
					progress.Cancel()
				}
			}

			record(res.index, completed, res.id, res.value, res.err)
		}
	}

	if progress != nil && !report.Cancelled {
		progress.Finish()
	}

	return report
}
