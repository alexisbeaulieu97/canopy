// Package workspaces provides the core business logic for workspace management.
package workspaces

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// parallelCanonicalOptions configures the parallel execution behavior.
type parallelCanonicalOptions struct {
	workers int
}

// runParallelCanonical executes EnsureCanonical for multiple repos in parallel
// with bounded concurrency and fail-fast behavior.
//
// On success, returns nil.
// On failure, cancels remaining operations and returns the first error.
func (s *Service) runParallelCanonical(ctx context.Context, repos []domain.Repo, opts parallelCanonicalOptions) error {
	if len(repos) == 0 {
		return nil
	}

	// For single repo or single worker, run sequentially
	if len(repos) == 1 || opts.workers == 1 {
		return s.runSequentialCanonical(ctx, repos)
	}

	// Determine worker count (ensure at least 1)
	numWorkers := opts.workers
	if numWorkers <= 0 {
		numWorkers = 1
	}

	// Create errgroup with bounded concurrency and fail-fast cancellation
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(numWorkers)

	for _, repo := range repos {
		g.Go(func() error {
			_, err := s.gitEngine.EnsureCanonical(ctx, repo.URL, repo.Name)
			if err != nil {
				return cerrors.WrapGitError(err, "ensure canonical for "+repo.Name)
			}

			return nil
		})
	}

	return g.Wait()
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
