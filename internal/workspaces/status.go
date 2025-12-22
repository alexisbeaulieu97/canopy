package workspaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// WorkspacePath returns the absolute path for a workspace ID.
func (s *Service) WorkspacePath(ctx context.Context, workspaceID string) (string, error) {
	// Use Load instead of List to avoid O(n) scan of all workspaces
	_, dirName, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		return "", err
	}

	return filepath.Join(s.config.GetWorkspacesRoot(), dirName), nil
}

// ListWorkspaces returns all active workspaces
func (s *Service) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	workspaceList, err := s.wsEngine.List(ctx)
	if err != nil {
		return nil, err
	}

	var workspaces []domain.Workspace

	for _, w := range workspaceList {
		dirName := w.DirName
		if dirName == "" {
			var dirErr error

			dirName, dirErr = s.config.ComputeWorkspaceDir(w.ID)
			if dirErr != nil {
				return nil, dirErr
			}
		}

		wsPath := filepath.Join(s.config.GetWorkspacesRoot(), dirName)

		usage, latest, sizeErr := s.diskUsage.CachedUsage(wsPath)
		if sizeErr != nil {
			if s.logger != nil {
				s.logger.Debug("Failed to calculate workspace stats", "workspace", w.ID, "error", sizeErr)
			}
		}

		if usage > 0 {
			w.DiskUsageBytes = usage
		}

		if !latest.IsZero() {
			w.LastModified = latest
		} else if info, statErr := os.Stat(wsPath); statErr == nil {
			w.LastModified = info.ModTime()
		}

		workspaces = append(workspaces, w)
	}

	return workspaces, nil
}

// WorkspaceLocked reports whether a workspace currently has an active lock.
func (s *Service) WorkspaceLocked(workspaceID string) (bool, error) {
	if s.lockManager == nil {
		return false, nil
	}

	return s.lockManager.IsLocked(workspaceID)
}

// ListClosedWorkspaces returns closed workspace metadata.
func (s *Service) ListClosedWorkspaces(ctx context.Context) ([]domain.ClosedWorkspace, error) {
	return s.wsEngine.ListClosed(ctx)
}

// GetStatus returns the aggregate status of a workspace
func (s *Service) GetStatus(ctx context.Context, workspaceID string) (*domain.WorkspaceStatus, error) {
	targetWorkspace, dirName, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// 2. Check status for each repo
	var repoStatuses []domain.RepoStatus

	for _, repo := range targetWorkspace.Repos {
		// Check for context cancellation before each repo
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)

		isDirty, unpushed, behind, branch, err := s.gitEngine.Status(ctx, worktreePath)
		if err != nil {
			statusErr := domain.StatusError(err.Error())
			if errors.Is(err, context.DeadlineExceeded) {
				statusErr = domain.StatusErrorTimeout
			}

			repoStatuses = append(repoStatuses, domain.RepoStatus{
				Name:  repo.Name,
				Error: statusErr,
			})

			continue
		}

		repoStatuses = append(repoStatuses, domain.RepoStatus{
			Name:            repo.Name,
			IsDirty:         isDirty,
			UnpushedCommits: unpushed,
			BehindRemote:    behind,
			Branch:          branch,
		})
	}

	return &domain.WorkspaceStatus{
		ID:         workspaceID,
		BranchName: targetWorkspace.BranchName,
		Repos:      repoStatuses,
	}, nil
}
