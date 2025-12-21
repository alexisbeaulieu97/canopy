package workspaces

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// WorkspaceStatusResult captures status results for a single workspace.
type WorkspaceStatusResult struct {
	WorkspaceID string
	Status      *domain.WorkspaceStatus
	Err         error
}

// GetWorkspaceStatusBatch fetches workspace status in parallel with bounded concurrency.
func (s *Service) GetWorkspaceStatusBatch(ctx context.Context, workspaceIDs []string, timeout time.Duration) ([]WorkspaceStatusResult, error) {
	results := make([]WorkspaceStatusResult, len(workspaceIDs))
	if len(workspaceIDs) == 0 {
		return results, nil
	}

	workers := s.statusBatchWorkers()
	if len(workspaceIDs) == 1 || workers == 1 {
		return s.getWorkspaceStatusSequential(ctx, workspaceIDs, timeout)
	}

	return s.getWorkspaceStatusParallel(ctx, workspaceIDs, timeout, workers)
}

func (s *Service) getStatusWithTimeout(ctx context.Context, workspaceID string, timeout time.Duration) (*domain.WorkspaceStatus, error) {
	if timeout <= 0 {
		return s.GetStatus(ctx, workspaceID)
	}

	statusCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.GetStatus(statusCtx, workspaceID)
}

func (s *Service) statusBatchWorkers() int {
	workers := s.config.GetParallelWorkers()
	if workers <= 0 {
		return 1
	}

	return workers
}

func (s *Service) getWorkspaceStatusSequential(ctx context.Context, workspaceIDs []string, timeout time.Duration) ([]WorkspaceStatusResult, error) {
	results := make([]WorkspaceStatusResult, len(workspaceIDs))
	for i, workspaceID := range workspaceIDs {
		if ctx.Err() != nil {
			return results, ctx.Err()
		}

		status, err := s.getStatusWithTimeout(ctx, workspaceID, timeout)
		results[i] = WorkspaceStatusResult{WorkspaceID: workspaceID, Status: status, Err: err}
	}

	return results, nil
}

func (s *Service) getWorkspaceStatusParallel(ctx context.Context, workspaceIDs []string, timeout time.Duration, workers int) ([]WorkspaceStatusResult, error) {
	type statusResult struct {
		index  int
		result WorkspaceStatusResult
	}

	resultsCh := make(chan statusResult, len(workspaceIDs))
	g, groupCtx := errgroup.WithContext(ctx)
	g.SetLimit(workers)

	for i, workspaceID := range workspaceIDs {
		i := i
		workspaceID := workspaceID

		g.Go(func() error {
			if groupCtx.Err() != nil {
				return groupCtx.Err()
			}

			status, err := s.getStatusWithTimeout(groupCtx, workspaceID, timeout)
			res := WorkspaceStatusResult{WorkspaceID: workspaceID, Status: status, Err: err}

			select {
			case resultsCh <- statusResult{index: i, result: res}:
				return nil
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		})
	}

	err := g.Wait()

	close(resultsCh)

	results := make([]WorkspaceStatusResult, len(workspaceIDs))
	for result := range resultsCh {
		results[result.index] = result.result
	}

	return results, err
}
