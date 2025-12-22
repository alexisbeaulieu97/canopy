// Package workspaces provides the core business logic for workspace management.
package workspaces

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// ParallelOptions configures parallel execution behavior.
type ParallelOptions struct {
	Workers         int
	ContinueOnError bool
	AggregateErrors bool
}

// ParallelResult captures the result of a parallel task.
type ParallelResult[T any] struct {
	Value T
	Err   error
}

// ParallelExecutor provides bounded parallel execution helpers.
type ParallelExecutor struct {
	workers int
}

// NewParallelExecutor creates a ParallelExecutor with a default worker limit.
func NewParallelExecutor(workers int) *ParallelExecutor {
	return &ParallelExecutor{workers: workers}
}

// Run executes a task for each index with bounded concurrency.
// When ContinueOnError is false, the first error cancels remaining tasks.
func (e *ParallelExecutor) Run(ctx context.Context, total int, fn func(ctx context.Context, index int) error, opts ParallelOptions) error {
	if total == 0 {
		return nil
	}

	workers := e.workerLimit(total, opts.Workers)
	if workers == 1 {
		return e.runSequential(ctx, total, fn, opts)
	}

	g, groupCtx := errgroup.WithContext(ctx)
	g.SetLimit(workers)

	for i := 0; i < total; i++ {
		i := i

		g.Go(func() error {
			err := fn(groupCtx, i)
			if err != nil && !opts.ContinueOnError {
				return err
			}

			return nil
		})
	}

	err := g.Wait()
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	return err
}

// ParallelMap executes a task for each index and captures its result.
// When ContinueOnError is false, the first error cancels remaining tasks.
func ParallelMap[T any](ctx context.Context, executor *ParallelExecutor, total int, fn func(ctx context.Context, index int) (T, error), opts ParallelOptions) ([]ParallelResult[T], error) {
	results := make([]ParallelResult[T], total)
	if total == 0 {
		return results, nil
	}

	workers := executor.workerLimit(total, opts.Workers)
	if workers == 1 {
		err := parallelMapSequential(ctx, total, fn, opts, results)
		return results, finalizeMapResults(ctx, results, err, opts)
	}

	err := parallelMapConcurrent(ctx, workers, total, fn, opts, results)

	return results, finalizeMapResults(ctx, results, err, opts)
}

// ExtractValues returns the ordered values from a result slice.
func ExtractValues[T any](results []ParallelResult[T]) []T {
	values := make([]T, len(results))
	for i, result := range results {
		values[i] = result.Value
	}

	return values
}

// FirstError returns the first error in the result slice, if any.
func FirstError[T any](results []ParallelResult[T]) error {
	for _, result := range results {
		if result.Err != nil {
			return result.Err
		}
	}

	return nil
}

// CountErrors returns the number of results that include errors.
func CountErrors[T any](results []ParallelResult[T]) int {
	count := 0

	for _, result := range results {
		if result.Err != nil {
			count++
		}
	}

	return count
}

// AggregateErrors joins all errors from the result slice into a single error.
func AggregateErrors[T any](results []ParallelResult[T]) error {
	errs := make([]error, 0, len(results))
	for _, result := range results {
		if result.Err != nil {
			errs = append(errs, result.Err)
		}
	}

	return joinErrors(errs...)
}

func (e *ParallelExecutor) workerLimit(total, override int) int {
	workers := override
	if workers <= 0 {
		workers = e.workers
	}

	if workers <= 0 {
		workers = 1
	}

	if workers > total {
		workers = total
	}

	return workers
}

func (e *ParallelExecutor) runSequential(ctx context.Context, total int, fn func(ctx context.Context, index int) error, opts ParallelOptions) error {
	for i := 0; i < total; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn(ctx, i)
		if err != nil && !opts.ContinueOnError {
			return err
		}
	}

	return ctx.Err()
}

func parallelMapSequential[T any](ctx context.Context, total int, fn func(ctx context.Context, index int) (T, error), opts ParallelOptions, results []ParallelResult[T]) error {
	for i := 0; i < total; i++ {
		if ctx.Err() != nil {
			results[i] = ParallelResult[T]{Err: ctx.Err()}
			if !opts.ContinueOnError {
				return ctx.Err()
			}

			continue
		}

		value, err := fn(ctx, i)

		results[i] = ParallelResult[T]{Value: value, Err: err}
		if err != nil && !opts.ContinueOnError {
			return err
		}
	}

	return ctx.Err()
}

func parallelMapConcurrent[T any](ctx context.Context, workers, total int, fn func(ctx context.Context, index int) (T, error), opts ParallelOptions, results []ParallelResult[T]) error {
	g, groupCtx := errgroup.WithContext(ctx)
	g.SetLimit(workers)

	for i := 0; i < total; i++ {
		i := i

		g.Go(func() error {
			value, err := fn(groupCtx, i)
			results[i] = ParallelResult[T]{Value: value, Err: err}

			if err != nil && !opts.ContinueOnError {
				return err
			}

			return nil
		})
	}

	return g.Wait()
}

func finalizeMapResults[T any](ctx context.Context, results []ParallelResult[T], err error, opts ParallelOptions) error {
	if err == nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	if err != nil {
		return err
	}

	if opts.AggregateErrors {
		return AggregateErrors(results)
	}

	return nil
}
