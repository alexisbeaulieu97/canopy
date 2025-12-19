package workspaces

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// SyncOptions configures workspace sync behavior.
type SyncOptions struct {
	Timeout time.Duration
}

// SyncWorkspace pulls updates for all repositories in the workspace.
func (s *Service) SyncWorkspace(ctx context.Context, id string, opts SyncOptions) (*domain.SyncResult, error) {
	var result *domain.SyncResult

	if err := s.withWorkspaceLock(ctx, id, false, func() error {
		ws, _, err := s.findWorkspace(ctx, id)
		if err != nil {
			return err
		}

		if opts.Timeout == 0 {
			opts.Timeout = 60 * time.Second // Default timeout
		}

		results := make([]domain.RepoSyncStatus, len(ws.Repos))

		var (
			wg sync.WaitGroup
			mu sync.Mutex
		)

		numWorkers := s.config.GetParallelWorkers()
		if numWorkers <= 0 {
			numWorkers = 1
		}

		reposChan := make(chan struct {
			index int
			repo  domain.Repo
		}, len(ws.Repos))

		for i, repo := range ws.Repos {
			reposChan <- struct {
				index int
				repo  domain.Repo
			}{i, repo}
		}

		close(reposChan)

		if numWorkers > len(ws.Repos) {
			numWorkers = len(ws.Repos)
		}

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for r := range reposChan {
					repoResult := s.syncRepo(ctx, id, r.repo, opts.Timeout)

					mu.Lock()

					results[r.index] = repoResult

					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		result = s.aggregateSyncResults(id, results)

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) aggregateSyncResults(workspaceID string, results []domain.RepoSyncStatus) *domain.SyncResult {
	syncResult := &domain.SyncResult{
		WorkspaceID: workspaceID,
		Repos:       results,
	}

	for _, r := range results {
		if r.Status == domain.SyncStatusUpdated {
			syncResult.TotalUpdated += r.Updated
		}

		if r.Status == domain.SyncStatusError || r.Status == domain.SyncStatusTimeout || r.Status == domain.SyncStatusConflict {
			syncResult.TotalErrors++
		}
	}

	return syncResult
}

func (s *Service) syncRepo(ctx context.Context, wsID string, repo domain.Repo, timeout time.Duration) domain.RepoSyncStatus {
	result := domain.RepoSyncStatus{
		Name:   repo.Name,
		Status: domain.SyncStatusUpToDate,
	}

	repoCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. Fetch canonical
	if err := s.gitEngine.Fetch(repoCtx, repo.Name); err != nil {
		result.Updated = 0
		if isDeadlineExceeded(err) {
			result.Status = domain.SyncStatusTimeout
			result.Error = "timeout during fetch"

			return result
		}

		result.Status = domain.SyncStatusError
		result.Error = fmt.Sprintf("fetch failed: %v", err)

		return result
	}

	worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), wsID, repo.Name)

	// 2. Get status before pull to see behind count
	_, _, behind, _, err := s.gitEngine.Status(repoCtx, worktreePath)
	if err != nil {
		result.Status = domain.SyncStatusError
		result.Error = fmt.Sprintf("status failed: %v", err)
		result.Updated = 0

		return result
	}

	result.Updated = behind

	// 3. Pull worktree only if behind remote
	if result.Updated > 0 {
		err = s.gitEngine.Pull(repoCtx, worktreePath)
		if err != nil {
			result.Updated = 0
			if isDeadlineExceeded(err) {
				result.Status = domain.SyncStatusTimeout
				result.Error = "timeout during pull"

				return result
			}

			// Check for conflicts - Usually go-git returns error if pull cannot be done cleanly.
			// For now we treat it as error, but we could improve detection.
			result.Status = domain.SyncStatusError
			result.Error = fmt.Sprintf("pull failed: %v", err)

			return result
		}

		result.Status = domain.SyncStatusUpdated
	}

	return result
}
