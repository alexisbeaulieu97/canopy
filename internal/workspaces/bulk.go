package workspaces

import (
	"context"
	"regexp"
	"strings"

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

	executor := NewParallelExecutor(s.config.GetParallelWorkers())

	results, err := ParallelMap(ctx, executor, len(workspaces), func(runCtx context.Context, index int) (WorkspaceSyncResult, error) {
		workspaceID := workspaces[index].ID
		syncResult, syncErr := s.SyncWorkspace(runCtx, workspaceID, opts)

		return WorkspaceSyncResult{
			WorkspaceID: workspaceID,
			Result:      syncResult,
			Err:         syncErr,
		}, syncErr
	}, ParallelOptions{ContinueOnError: true, AggregateErrors: true})
	if err != nil {
		return nil, err
	}

	return &BulkSyncResult{Results: ExtractValues(results)}, nil
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
