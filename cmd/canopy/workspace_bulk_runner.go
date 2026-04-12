package main

import (
	"bufio"
	"context"
	"os"
	"strings"

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
	Parallelism   int
	ShowProgress  bool
	OnStart       func(id string, current, total int)
	OnSuccess     func(id string, current, total int, value T)
	OnFailure     func(id string, current, total int, err error)
	ProgressLabel func(id string, value T, err error) string
}

type bulkWorkspaceRunReport[T any] struct {
	Results    []bulkWorkspaceResult[T]
	SuccessIDs []string
	FailedIDs  []string
	FirstErr   error
	Cancelled  bool
}

type bulkWorkspaceRecorder[T any] struct {
	report   *bulkWorkspaceRunReport[T]
	opts     bulkWorkspaceRunOptions[T]
	progress *output.Progress
	total    int
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

	recorder := bulkWorkspaceRecorder[T]{
		report:   &report,
		opts:     opts,
		progress: newBulkWorkspaceProgress(len(ids), opts.ShowProgress),
		total:    len(ids),
	}

	runBulkWorkspaceExecution(ctx, ids, normalizeBulkParallelism(len(ids), opts.Parallelism), opts, execute, recorder)

	recorder.finish()

	return report
}

func normalizeBulkParallelism(total, requested int) int {
	if requested <= 0 || requested > total {
		return total
	}

	return requested
}

func newBulkWorkspaceProgress(total int, show bool) *output.Progress {
	if !show {
		return nil
	}

	return output.NewProgress(output.DefaultProgressOptions(total))
}

func runBulkWorkspaceExecution[T any](
	ctx context.Context,
	ids []string,
	parallelism int,
	opts bulkWorkspaceRunOptions[T],
	execute func(context.Context, string) (T, error),
	recorder bulkWorkspaceRecorder[T],
) {
	results, runErrCh := startBulkWorkspaceExecution(ctx, ids, parallelism, opts, execute)

	completed := 0
	for res := range results {
		completed++

		if ctx.Err() != nil {
			recorder.cancel()
		}

		recorder.record(res.index, completed, res.id, res.value, res.err)
	}

	if err := <-runErrCh; err != nil {
		recorder.cancel()
	}
}

type bulkWorkspaceExecutionResult[T any] struct {
	index int
	id    string
	value T
	err   error
}

func startBulkWorkspaceExecution[T any](
	ctx context.Context,
	ids []string,
	parallelism int,
	opts bulkWorkspaceRunOptions[T],
	execute func(context.Context, string) (T, error),
) (<-chan bulkWorkspaceExecutionResult[T], <-chan error) {
	results := make(chan bulkWorkspaceExecutionResult[T], len(ids))
	runErrCh := make(chan error, 1)
	executor := workspaces.NewParallelExecutor(parallelism)

	go func() {
		runErrCh <- executor.Run(ctx, len(ids), func(runCtx context.Context, index int) error {
			runBulkWorkspaceTask(runCtx, ids, index, parallelism, opts, execute, results)

			return nil
		}, workspaces.ParallelOptions{
			Workers:         parallelism,
			ContinueOnError: true,
		})

		close(results)
	}()

	return results, runErrCh
}

func runBulkWorkspaceTask[T any](
	runCtx context.Context,
	ids []string,
	index int,
	parallelism int,
	opts bulkWorkspaceRunOptions[T],
	execute func(context.Context, string) (T, error),
	results chan<- bulkWorkspaceExecutionResult[T],
) {
	id := ids[index]
	if parallelism == 1 && opts.OnStart != nil && !opts.ShowProgress {
		opts.OnStart(id, index+1, len(ids))
	}

	if runCtx.Err() != nil {
		var zero T
		results <- bulkWorkspaceExecutionResult[T]{index: index, id: id, value: zero, err: runCtx.Err()}

		return
	}

	value, err := execute(runCtx, id)
	results <- bulkWorkspaceExecutionResult[T]{index: index, id: id, value: value, err: err}
}

func (r bulkWorkspaceRecorder[T]) record(index, current int, id string, value T, err error) {
	r.report.Results[index] = bulkWorkspaceResult[T]{ID: id, Value: value, Err: err}

	if err != nil {
		if r.report.FirstErr == nil {
			r.report.FirstErr = err
		}

		r.report.FailedIDs = append(r.report.FailedIDs, id)
		if r.opts.OnFailure != nil {
			r.opts.OnFailure(id, current, r.total, err)
		}
	} else {
		r.report.SuccessIDs = append(r.report.SuccessIDs, id)
		if r.opts.OnSuccess != nil && !r.opts.ShowProgress {
			r.opts.OnSuccess(id, current, r.total, value)
		}
	}

	if r.progress == nil || r.report.Cancelled {
		return
	}

	label := id
	if r.opts.ProgressLabel != nil {
		label = r.opts.ProgressLabel(id, value, err)
	} else if err != nil {
		label = id + " (failed)"
	}

	r.progress.Increment(label)
}

func (r bulkWorkspaceRecorder[T]) cancel() {
	if r.report.Cancelled {
		return
	}

	r.report.Cancelled = true
	if r.progress != nil {
		r.progress.Cancel()
	}
}

func (r bulkWorkspaceRecorder[T]) finish() {
	if r.progress != nil && !r.report.Cancelled {
		r.progress.Finish()
	}
}
