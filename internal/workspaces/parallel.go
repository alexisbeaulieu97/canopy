// Package workspaces provides the core business logic for workspace management.
package workspaces

import (
	"context"
	"sync"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// canonicalResult holds the result of an EnsureCanonical operation.
type canonicalResult struct {
	repo domain.Repo
	err  error
}

// parallelCanonicalOptions configures the parallel execution behavior.
type parallelCanonicalOptions struct {
	workers int
}

// runParallelCanonical executes EnsureCanonical for multiple repos in parallel
// with bounded concurrency and fail-fast behavior.
//
// On success, returns nil.
// On failure, cancels remaining operations and returns the first error.
//
//nolint:gocyclo // Concurrent coordination inherently requires multiple control flow paths
func (s *Service) runParallelCanonical(ctx context.Context, repos []domain.Repo, opts parallelCanonicalOptions) error {
	if len(repos) == 0 {
		return nil
	}

	// For single repo or single worker, run sequentially
	if len(repos) == 1 || opts.workers == 1 {
		return s.runSequentialCanonical(ctx, repos)
	}

	// Create a cancellable context for fail-fast behavior
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Channels for work distribution and result collection
	jobs := make(chan int, len(repos))
	results := make(chan canonicalResult, len(repos))

	// Determine worker count
	numWorkers := opts.workers
	if numWorkers > len(repos) {
		numWorkers = len(repos)
	}

	// Start workers
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)

		go s.canonicalWorker(ctx, repos, jobs, results, &wg)
	}

	// Send all jobs
	for i := range repos {
		jobs <- i
	}

	close(jobs)

	// Wait for all workers to complete in background
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with fail-fast
	return s.collectCanonicalResults(results, cancel)
}

// canonicalWorker processes jobs from the jobs channel and sends results.
func (s *Service) canonicalWorker(ctx context.Context, repos []domain.Repo, jobs <-chan int, results chan<- canonicalResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case idx, ok := <-jobs:
			if !ok {
				return
			}

			repo := repos[idx]
			_, err := s.gitEngine.EnsureCanonical(ctx, repo.URL, repo.Name)

			select {
			case results <- canonicalResult{repo: repo, err: err}:
			case <-ctx.Done():
				return
			}
		}
	}
}

// collectCanonicalResults collects results and returns the first error encountered.
func (s *Service) collectCanonicalResults(results <-chan canonicalResult, cancel context.CancelFunc) error {
	var firstErr error

	for result := range results {
		if result.err != nil && firstErr == nil {
			firstErr = cerrors.WrapGitError(result.err, "ensure canonical for "+result.repo.Name)

			cancel() // Cancel remaining operations
		}
	}

	return firstErr
}

// runSequentialCanonical executes EnsureCanonical for repos sequentially.
func (s *Service) runSequentialCanonical(ctx context.Context, repos []domain.Repo) error {
	for _, repo := range repos {
		if ctx.Err() != nil {
			return cerrors.NewContextError(ctx, "ensure canonical", repo.Name)
		}

		_, err := s.gitEngine.EnsureCanonical(ctx, repo.URL, repo.Name)
		if err != nil {
			return cerrors.WrapGitError(err, "ensure canonical for "+repo.Name)
		}
	}

	return nil
}
