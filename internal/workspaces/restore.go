package workspaces

import (
	"context"
	"fmt"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// RestoreWorkspace recreates a workspace from the newest closed entry.
func (s *Service) RestoreWorkspace(ctx context.Context, workspaceID string, force bool) error {
	return s.withWorkspaceLock(ctx, workspaceID, true, func() error {
		archive, err := s.wsEngine.LatestClosed(ctx, workspaceID)
		if err != nil {
			return err
		}

		if _, _, err := s.findWorkspace(ctx, workspaceID); err == nil {
			if !force {
				return cerrors.NewWorkspaceExists(workspaceID).WithContext("hint", "Use --force to replace or choose a different ID")
			}

			if err := s.closeWorkspaceWithOptionsUnlocked(ctx, workspaceID, true, CloseOptions{}); err != nil {
				return cerrors.NewIOFailed("remove existing workspace", err)
			}
		}

		ws := archive.Metadata
		ws.ClosedAt = nil

		op := NewOperation(s.logger)
		op.AddStep(func() error {
			if err := s.createWorkspaceWithOptionsUnlocked(ctx, ws.ID, ws.BranchName, ws.Repos, CreateOptions{}); err != nil {
				// Preserve original error type if it's already typed
				var canopyErr *cerrors.CanopyError
				if isCanopyError(err, &canopyErr) {
					return canopyErr.WithContext("operation", fmt.Sprintf("restore workspace %s", workspaceID))
				}

				return cerrors.Wrap(cerrors.ErrIOFailed, fmt.Sprintf("failed to restore workspace %s", workspaceID), err)
			}

			return nil
		}, nil)
		op.AddStep(func() error {
			// Delete the closed entry using ID and timestamp
			closedAt := archive.ClosedAt()
			if err := s.wsEngine.DeleteClosed(ctx, workspaceID, closedAt); err != nil {
				return cerrors.NewIOFailed("remove closed entry", err)
			}

			return nil
		}, nil)

		return op.Execute()
	})
}
