package workspaces

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// RenameWorkspace renames a workspace to a new ID.
// If renameBranch is true and the branch name matches the old ID, it will also rename branches.
// If force is true, an existing workspace with the new ID will be deleted first.
func (s *Service) RenameWorkspace(ctx context.Context, oldID, newID string, renameBranch, force bool) error {
	if s.lockManager == nil {
		_, err := s.renameWorkspaceUnlocked(ctx, oldID, newID, renameBranch, force)
		return err
	}

	handle, err := s.lockManager.Acquire(ctx, oldID, false)
	if err != nil {
		return s.handleClosedWorkspaceError(ctx, oldID, err)
	}

	newDirName, renameErr := s.renameWorkspaceUnlocked(ctx, oldID, newID, renameBranch, force)
	if renameErr == nil && newDirName != "" {
		handle.UpdateLocation(newID, filepath.Join(s.config.GetWorkspacesRoot(), newDirName, lockFileName))
	}

	releaseErr := handle.Release()
	if releaseErr != nil && s.logger != nil {
		s.logger.Warn("workspace lock release failed", "workspace_id", oldID, "error", releaseErr)
	}

	if renameErr != nil {
		return renameErr
	}

	return nil
}

func (s *Service) renameWorkspaceUnlocked(ctx context.Context, oldID, newID string, renameBranch, force bool) (string, error) {
	workspace, dirName, err := s.findWorkspace(ctx, oldID)
	if err != nil {
		return "", s.handleClosedWorkspaceError(ctx, oldID, err)
	}

	if err := s.validateRenameInputs(newID); err != nil {
		return "", err
	}

	if oldID == newID {
		return "", cerrors.NewInvalidArgument("new_id", "cannot rename workspace to the same ID")
	}

	if err := s.ensureTargetAvailableOrDelete(ctx, newID, force); err != nil {
		return "", err
	}

	shouldRenameBranch := renameBranch && workspace.BranchName == oldID

	newDirName, err := s.config.ComputeWorkspaceDir(newID)
	if err != nil {
		return "", err
	}

	if err := s.executeRename(ctx, *workspace, oldID, newID, dirName, newDirName, shouldRenameBranch); err != nil {
		return "", err
	}

	s.invalidateWorkspaceCache(oldID, newID)

	return newDirName, nil
}

// validateRenameInputs validates the inputs for renaming a workspace.
func (s *Service) validateRenameInputs(newID string) error {
	return validation.ValidateWorkspaceID(newID)
}

// ensureNewIDAvailable checks that the new workspace ID doesn't already exist.
func (s *Service) ensureNewIDAvailable(ctx context.Context, newID string) error {
	_, err := s.wsEngine.Load(ctx, newID)
	if err == nil {
		return cerrors.NewWorkspaceExists(newID)
	}

	// If the error is "workspace not found", that's what we want - the ID is available
	if isWorkspaceNotFound(err) {
		return nil
	}

	// For any other error (IO failure, etc.), propagate it
	return err
}

// handleClosedWorkspaceError checks if the error is due to a closed workspace and returns a more helpful error message.
func (s *Service) handleClosedWorkspaceError(ctx context.Context, workspaceID string, err error) error {
	if !isWorkspaceNotFound(err) {
		return err
	}

	closed, closedErr := s.wsEngine.LatestClosed(ctx, workspaceID)
	if closedErr != nil || closed == nil {
		return err
	}

	return cerrors.NewInvalidArgument("workspace", "cannot rename closed workspace; reopen first with 'workspace open'")
}

// ensureTargetAvailableOrDelete checks if the target ID is available, optionally force-deleting an existing workspace.
func (s *Service) ensureTargetAvailableOrDelete(ctx context.Context, newID string, force bool) error {
	existingErr := s.ensureNewIDAvailable(ctx, newID)
	if existingErr == nil {
		return nil
	}

	// Only force-delete when the error is specifically "workspace exists"
	var canopyErr *cerrors.CanopyError

	isWorkspaceExists := isCanopyError(existingErr, &canopyErr) && canopyErr.Code == cerrors.ErrWorkspaceExists

	if !force || !isWorkspaceExists {
		return existingErr
	}

	return s.forceDeleteWorkspace(ctx, newID)
}

func (s *Service) forceDeleteWorkspace(ctx context.Context, workspaceID string) error {
	if s.lockManager != nil {
		handle, err := s.lockManager.Acquire(ctx, workspaceID, false)
		if err != nil {
			if isWorkspaceNotFound(err) {
				return nil
			}

			return err
		}

		defer func() {
			if releaseErr := handle.Release(); releaseErr != nil && s.logger != nil {
				s.logger.Warn("workspace lock release failed", "workspace_id", workspaceID, "error", releaseErr)
			}
		}()
	}

	if deleteErr := s.wsEngine.Delete(ctx, workspaceID); deleteErr != nil {
		return cerrors.NewInternalError("failed to delete existing workspace for force rename", deleteErr)
	}

	return nil
}

// renameBranchesInRepos renames branches in all repos and returns the list of repos that were renamed.
func (s *Service) renameBranchesInRepos(ctx context.Context, workspace domain.Workspace, dirName, oldID, newID string) error {
	for _, repo := range workspace.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)

		if err := s.gitEngine.RenameBranch(ctx, worktreePath, oldID, newID); err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("rename branch in repo %s", repo.Name))
		}
	}

	return nil
}

// rollbackBranchRenames attempts to rollback branch renames on failure (best effort, ignores errors).
func (s *Service) rollbackBranchRenames(ctx context.Context, workspace domain.Workspace, dirName, oldID, newID string) {
	for _, repo := range workspace.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		_ = s.gitEngine.RenameBranch(ctx, worktreePath, newID, oldID) // best effort rollback
	}
}

// rollbackBranchRenamesWithError attempts to rollback branch renames and reports errors.
func (s *Service) rollbackBranchRenamesWithError(ctx context.Context, workspace domain.Workspace, dirName, oldID, newID string) error {
	var errs []error

	for _, repo := range workspace.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.RenameBranch(ctx, worktreePath, newID, oldID); err != nil {
			errs = append(errs, cerrors.WrapGitError(err, fmt.Sprintf("rollback branch rename in repo %s", repo.Name)))
		}
	}

	return joinErrors(errs...)
}

// updateBranchMetadata loads the workspace and updates the branch name metadata.
func (s *Service) updateBranchMetadata(ctx context.Context, workspaceID, newBranchName string) error {
	ws, err := s.wsEngine.Load(ctx, workspaceID)
	if err != nil {
		return cerrors.NewWorkspaceMetadataError(workspaceID, "load", err)
	}

	ws.BranchName = newBranchName
	if err := s.wsEngine.Save(ctx, *ws); err != nil {
		return cerrors.NewWorkspaceMetadataError(workspaceID, "save", err)
	}

	return nil
}

// renameWorkspaceDir renames the workspace directory and handles rollback on failure.
func (s *Service) renameWorkspaceDir(ctx context.Context, workspace domain.Workspace, oldID, newID, oldDirName string, shouldRenameBranch bool) error {
	if err := s.wsEngine.Rename(ctx, oldID, newID); err != nil {
		if shouldRenameBranch {
			s.rollbackBranchRenames(ctx, workspace, oldDirName, oldID, newID)
		}

		return err
	}

	return nil
}

// invalidateWorkspaceCache invalidates cache entries for the given workspace IDs.
func (s *Service) invalidateWorkspaceCache(ids ...string) {
	if s.cache != nil {
		for _, id := range ids {
			s.cache.Invalidate(id)
		}
	}
}

// updateBranchMetadataWithRollback updates workspace metadata and rolls back branch and directory renames on failure.
func (s *Service) updateBranchMetadataWithRollback(ctx context.Context, workspace domain.Workspace, oldID, newID, newDirName string) error {
	if err := s.updateBranchMetadata(ctx, newID, newID); err != nil {
		var rollbackErrors []error

		// Attempt to rollback branch renames first so repo state aligns with directory rollback.
		if branchRollbackErr := s.rollbackBranchRenamesWithError(ctx, workspace, newDirName, oldID, newID); branchRollbackErr != nil {
			if s.logger != nil {
				s.logger.Error("failed to rollback branch renames after metadata update error",
					"error", branchRollbackErr,
					"from", newID,
					"to", oldID,
				)
			}

			rollbackErrors = append(rollbackErrors, cerrors.NewInternalError("branch rollback failed", branchRollbackErr))
		}

		// Then rollback directory rename.
		if dirRollbackErr := s.wsEngine.Rename(ctx, newID, oldID); dirRollbackErr != nil {
			if s.logger != nil {
				s.logger.Error("failed to rollback workspace rename after metadata update error",
					"error", dirRollbackErr,
					"from", newID,
					"to", oldID,
				)
			}

			rollbackErrors = append(rollbackErrors, cerrors.NewInternalError("workspace rename rollback failed", dirRollbackErr))
		}

		if len(rollbackErrors) > 0 {
			return joinErrors(append([]error{err}, rollbackErrors...)...)
		}

		return err
	}

	return nil
}

// executeRename performs the actual rename operations: branch rename, directory rename, and metadata update.
func (s *Service) executeRename(ctx context.Context, workspace domain.Workspace, oldID, newID, oldDirName, newDirName string, shouldRenameBranch bool) error {
	if shouldRenameBranch {
		if err := s.renameBranchesInRepos(ctx, workspace, oldDirName, oldID, newID); err != nil {
			return err
		}
	}

	if err := s.renameWorkspaceDir(ctx, workspace, oldID, newID, oldDirName, shouldRenameBranch); err != nil {
		return err
	}

	if shouldRenameBranch {
		if err := s.updateBranchMetadataWithRollback(ctx, workspace, oldID, newID, newDirName); err != nil {
			return err
		}
	}

	return nil
}
