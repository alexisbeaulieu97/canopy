package workspaces

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// BulkWorkspaceResult captures the outcome of a workspace-level bulk operation.
type BulkWorkspaceResult struct {
	WorkspaceID string
	Err         error
}

// BulkCloseResult contains results for bulk close operations.
type BulkCloseResult struct {
	Results []BulkWorkspaceResult
}

// WorkspaceSyncResult captures the outcome of syncing a workspace.
type WorkspaceSyncResult struct {
	WorkspaceID string
	Result      *domain.SyncResult
	Err         error
}

// BulkSyncResult contains results for bulk sync operations.
type BulkSyncResult struct {
	Results []WorkspaceSyncResult
}

// ListWorkspacesMatching returns workspaces with IDs that match the regex pattern.
func (s *Service) ListWorkspacesMatching(ctx context.Context, pattern string) ([]domain.Workspace, error) {
	re, err := compileWorkspacePattern(pattern)
	if err != nil {
		return nil, err
	}

	workspaces, err := s.wsEngine.List(ctx)
	if err != nil {
		return nil, err
	}

	var matched []domain.Workspace

	for _, ws := range workspaces {
		if re.MatchString(ws.ID) {
			matched = append(matched, ws)
		}
	}

	return matched, nil
}

// CloseWorkspacesMatching closes workspaces that match the regex pattern sequentially.
func (s *Service) CloseWorkspacesMatching(ctx context.Context, pattern string, force, keepMetadata bool, opts CloseOptions) (*BulkCloseResult, error) {
	workspaces, err := s.ListWorkspacesMatching(ctx, pattern)
	if err != nil {
		return nil, err
	}

	results := make([]BulkWorkspaceResult, len(workspaces))

	for i, ws := range workspaces {
		result := BulkWorkspaceResult{WorkspaceID: ws.ID}
		if keepMetadata {
			_, result.Err = s.CloseWorkspaceKeepMetadataWithOptions(ctx, ws.ID, force, opts)
		} else {
			result.Err = s.CloseWorkspaceWithOptions(ctx, ws.ID, force, opts)
		}

		results[i] = result
	}

	return &BulkCloseResult{Results: results}, nil
}

// SyncWorkspacesMatching syncs workspaces that match the regex pattern in parallel.
func (s *Service) SyncWorkspacesMatching(ctx context.Context, pattern string, opts SyncOptions) (*BulkSyncResult, error) {
	workspaces, err := s.ListWorkspacesMatching(ctx, pattern)
	if err != nil {
		return nil, err
	}

	if len(workspaces) == 0 {
		return &BulkSyncResult{Results: []WorkspaceSyncResult{}}, nil
	}

	results := make([]WorkspaceSyncResult, len(workspaces))

	type job struct {
		index int
		id    string
	}

	jobs := make(chan job, len(workspaces))
	for i, ws := range workspaces {
		jobs <- job{index: i, id: ws.ID}
	}

	close(jobs)

	numWorkers := s.config.GetParallelWorkers()
	if numWorkers <= 0 {
		numWorkers = 1
	}

	if numWorkers > len(workspaces) {
		numWorkers = len(workspaces)
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := range jobs {
				syncResult, syncErr := s.SyncWorkspace(ctx, j.id, opts)

				mu.Lock()

				results[j.index] = WorkspaceSyncResult{
					WorkspaceID: j.id,
					Result:      syncResult,
					Err:         syncErr,
				}

				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return &BulkSyncResult{Results: results}, nil
}

func compileWorkspacePattern(pattern string) (*regexp.Regexp, error) {
	if strings.TrimSpace(pattern) == "" {
		return nil, cerrors.NewInvalidArgument("pattern", "pattern is required")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, cerrors.NewInvalidArgument("pattern", err.Error())
	}

	return re, nil
}
