package workspaces

import (
	"context"
	"path/filepath"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// CloseOptions configures workspace close behavior.
type CloseOptions struct {
	SkipHooks         bool // Skip pre_close hooks
	ContinueOnHookErr bool // Continue if hooks fail
}

// CloseWorkspace removes a workspace with safety checks
func (s *Service) CloseWorkspace(ctx context.Context, workspaceID string, force bool) error {
	//nolint:contextcheck // Wrapper delegates to WithOptions which handles hooks with own timeout
	return s.CloseWorkspaceWithOptions(ctx, workspaceID, force, CloseOptions{})
}

// CloseWorkspaceWithOptions removes a workspace with configurable options.
//
//nolint:contextcheck // This function manages hook contexts internally with their own timeouts
func (s *Service) CloseWorkspaceWithOptions(ctx context.Context, workspaceID string, force bool, opts CloseOptions) error {
	return s.withWorkspaceLock(ctx, workspaceID, false, func() error {
		return s.closeWorkspaceWithOptionsUnlocked(ctx, workspaceID, force, opts)
	})
}

//nolint:contextcheck // This function manages hook contexts internally with their own timeouts
func (s *Service) closeWorkspaceWithOptionsUnlocked(ctx context.Context, workspaceID string, force bool, opts CloseOptions) error {
	targetWorkspace, _, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	if !force {
		if err := s.ensureWorkspaceClean(ctx, targetWorkspace, workspaceID, "close"); err != nil {
			return err
		}
	}

	// Run pre_close hooks before deletion
	if err := s.executePreCloseHooks(targetWorkspace, workspaceID, opts); err != nil {
		return err
	}

	// Delete workspace first, then clean up worktrees
	// This ensures metadata is consistent - if Delete fails, worktrees are still intact
	if err := s.wsEngine.Delete(ctx, workspaceID); err != nil {
		return err
	}

	// Remove worktrees from canonical repos after successful deletion
	s.removeWorkspaceWorktrees(targetWorkspace, workspaceID)

	// Invalidate cache after workspace deletion
	s.cache.Invalidate(workspaceID)

	return nil
}

// CloseWorkspaceKeepMetadata moves workspace metadata to the closed store and removes the active worktree.
func (s *Service) CloseWorkspaceKeepMetadata(ctx context.Context, workspaceID string, force bool) (*domain.ClosedWorkspace, error) {
	//nolint:contextcheck // Wrapper delegates to WithOptions which handles hooks with own timeout
	return s.CloseWorkspaceKeepMetadataWithOptions(ctx, workspaceID, force, CloseOptions{})
}

// CloseWorkspaceKeepMetadataWithOptions moves workspace metadata to the closed store with configurable options.
//
//nolint:contextcheck // This function manages hook contexts internally with their own timeouts
func (s *Service) CloseWorkspaceKeepMetadataWithOptions(ctx context.Context, workspaceID string, force bool, opts CloseOptions) (*domain.ClosedWorkspace, error) {
	var closed *domain.ClosedWorkspace

	if err := s.withWorkspaceLock(ctx, workspaceID, false, func() error {
		var err error
		closed, err = s.closeWorkspaceKeepMetadataWithOptionsUnlocked(ctx, workspaceID, force, opts)
		return err
	}); err != nil {
		return nil, err
	}

	return closed, nil
}

//nolint:contextcheck // This function manages hook contexts internally with their own timeouts
func (s *Service) closeWorkspaceKeepMetadataWithOptionsUnlocked(ctx context.Context, workspaceID string, force bool, opts CloseOptions) (*domain.ClosedWorkspace, error) {
	targetWorkspace, _, err := s.findWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if !force {
		if err := s.ensureWorkspaceClean(ctx, targetWorkspace, workspaceID, "close"); err != nil {
			return nil, err
		}
	}

	// Run pre_close hooks before archiving
	if err := s.executePreCloseHooks(targetWorkspace, workspaceID, opts); err != nil {
		return nil, err
	}

	closedAt := time.Now().UTC()

	archived, err := s.wsEngine.Close(ctx, workspaceID, closedAt)
	if err != nil {
		return nil, err
	}

	// Delete workspace first, then clean up worktrees
	if err := s.wsEngine.Delete(ctx, workspaceID); err != nil {
		rollbackErr := s.wsEngine.DeleteClosed(ctx, workspaceID, closedAt)
		if rollbackErr != nil {
			return nil, joinErrors(
				cerrors.NewIOFailed("remove workspace directory", err),
				cerrors.NewIOFailed("rollback closed entry", rollbackErr),
			)
		}

		return nil, cerrors.NewIOFailed("remove workspace directory", err)
	}

	// Remove worktrees from canonical repos after successful deletion
	s.removeWorkspaceWorktrees(targetWorkspace, workspaceID)

	// Invalidate cache after workspace deletion
	s.cache.Invalidate(workspaceID)

	return archived, nil
}

// executePreCloseHooks runs pre_close hooks if configured and not skipped.
// Returns nil if hooks are skipped, succeed, or ContinueOnHookErr is set.
//
//nolint:contextcheck // Hooks manage their own timeout context per-hook
func (s *Service) executePreCloseHooks(workspace *domain.Workspace, workspaceID string, opts CloseOptions) error {
	if opts.SkipHooks {
		return nil
	}

	hooksConfig := s.config.GetHooks()
	if len(hooksConfig.PreClose) == 0 {
		return nil
	}

	hookCtx := domain.HookContext{
		WorkspaceID:   workspaceID,
		WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), workspaceID),
		BranchName:    workspace.BranchName,
		Repos:         workspace.Repos,
	}

	if _, err := s.hookExecutor.ExecuteHooks(hooksConfig.PreClose, hookCtx, ports.HookExecuteOptions{
		ContinueOnError: opts.ContinueOnHookErr,
	}); err != nil {
		s.logger.Error("pre_close hooks failed", "error", err)
		// Per design.md: pre_close failure aborts close operation
		if !opts.ContinueOnHookErr {
			return err
		}
	}

	return nil
}

// PreviewCloseWorkspace returns a preview of what would happen when closing a workspace.
func (s *Service) PreviewCloseWorkspace(workspaceID string, keepMetadata bool) (*domain.WorkspaceClosePreview, error) {
	targetWorkspace, _, err := s.findWorkspace(context.Background(), workspaceID)
	if err != nil {
		return nil, err
	}

	wsPath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID)

	repoNames := []string{}
	repoStatuses := []domain.RepoCloseStatus{}

	for _, r := range targetWorkspace.Repos {
		repoNames = append(repoNames, r.Name)

		// Check repo status if git engine is available
		if s.gitEngine != nil {
			worktreePath := filepath.Join(wsPath, r.Name)

			isDirty, unpushed, _, _, statusErr := s.gitEngine.Status(context.Background(), worktreePath)
			if statusErr != nil {
				if s.logger != nil {
					s.logger.Debug("Failed to check repo status for preview",
						"repo", r.Name,
						"path", worktreePath,
						"error", statusErr)
				}
				// Include repo with unknown status (zeros)
				repoStatuses = append(repoStatuses, domain.RepoCloseStatus{
					Name: r.Name,
				})
			} else {
				repoStatuses = append(repoStatuses, domain.RepoCloseStatus{
					Name:          r.Name,
					IsDirty:       isDirty,
					UnpushedCount: unpushed,
				})
			}
		}
	}

	usage, _, sizeErr := s.diskUsage.CachedUsage(wsPath)
	if sizeErr != nil && s.logger != nil {
		s.logger.Debug("Failed to calculate workspace usage for preview", "workspace", workspaceID, "error", sizeErr)
	}

	return &domain.WorkspaceClosePreview{
		WorkspaceID:    workspaceID,
		WorkspacePath:  wsPath,
		BranchName:     targetWorkspace.BranchName,
		ReposAffected:  repoNames,
		RepoStatuses:   repoStatuses,
		DiskUsageBytes: usage,
		KeepMetadata:   keepMetadata,
	}, nil
}

func (s *Service) ensureWorkspaceClean(ctx context.Context, workspace *domain.Workspace, workspaceID, action string) error {
	if s.gitEngine == nil {
		return cerrors.NewInternalError("git engine not initialized", nil)
	}

	for _, repo := range workspace.Repos {
		// Check for context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID, repo.Name)

		isDirty, unpushed, _, _, err := s.gitEngine.Status(ctx, worktreePath)
		if err != nil {
			// Status errors should abort the operation - we can't verify cleanliness
			return cerrors.NewIOFailed("check repo status for "+repo.Name, err)
		}

		if isDirty {
			return cerrors.NewRepoNotClean(repo.Name, action)
		}

		if unpushed > 0 {
			return cerrors.NewRepoHasUnpushedCommits(repo.Name, unpushed, action)
		}
	}

	return nil
}

// removeWorkspaceWorktrees removes all worktrees from canonical repos for a workspace.
// This is called during workspace close to properly clean up git worktree references.
// Errors are logged but not returned since the workspace is being deleted anyway.
func (s *Service) removeWorkspaceWorktrees(workspace *domain.Workspace, workspaceID string) {
	if s.gitEngine == nil {
		return
	}

	for _, repo := range workspace.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID, repo.Name)

		//nolint:contextcheck // Using background context since this is cleanup during close
		if err := s.gitEngine.RemoveWorktree(context.Background(), repo.Name, worktreePath); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to remove worktree from canonical repo",
					"repo", repo.Name,
					"path", worktreePath,
					"error", err)
			}
		}

		// Prune stale references in case the worktree directory was already removed.
		//nolint:contextcheck // Using background context since this is cleanup during close
		if err := s.gitEngine.PruneWorktrees(context.Background(), repo.Name); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to prune worktrees for canonical repo",
					"repo", repo.Name,
					"error", err)
			}
		}
	}
}
