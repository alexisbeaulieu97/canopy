package workspaces

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// ResolveRepos determines which repos should be part of the workspace
func (s *Service) ResolveRepos(workspaceID string, requestedRepos []string) ([]domain.Repo, error) {
	var repoNames []string

	userRequested := len(requestedRepos) > 0

	// 1. Use requested repos if provided
	if userRequested {
		repoNames = requestedRepos
	} else {
		// 2. Fallback to config patterns
		repoNames = s.config.GetReposForWorkspace(workspaceID)
	}

	if len(repoNames) == 0 {
		return nil, cerrors.NewNoReposConfigured(workspaceID)
	}

	var repos []domain.Repo

	for _, raw := range repoNames {
		repo, ok, err := s.resolver.Resolve(raw, userRequested)
		if err != nil {
			return nil, err
		}

		if ok {
			repos = append(repos, repo)
		}
	}

	return repos, nil
}

// AddRepoToWorkspace adds a repository to an existing workspace
func (s *Service) AddRepoToWorkspace(ctx context.Context, workspaceID, repoName string) error {
	return s.withWorkspaceLock(ctx, workspaceID, false, func() error {
		if err := validateAddRepoInputs(workspaceID, repoName); err != nil {
			return err
		}

		workspace, _, err := s.findWorkspace(ctx, workspaceID)
		if err != nil {
			return err
		}

		if repoExistsInWorkspace(workspace.Repos, repoName) {
			return cerrors.NewRepoAlreadyExists(repoName, workspaceID)
		}

		repo, err := s.resolveWorkspaceRepo(workspaceID, repoName)
		if err != nil {
			return err
		}

		branchName, err := s.workspaceBranchName(workspaceID, workspace.BranchName)
		if err != nil {
			return err
		}

		op := NewOperation(s.logger)
		op.AddStep(func() error {
			return s.ensureWorkspaceWorktree(ctx, repo, workspaceID, branchName)
		}, func() error {
			return s.removeWorkspaceRepoWorktrees(ctx, workspaceID, []domain.Repo{repo})
		})
		op.AddStep(func() error {
			return s.saveWorkspaceRepo(ctx, workspaceID, workspace, repo)
		}, nil)

		if err := op.Execute(); err != nil {
			return err
		}

		s.cache.Invalidate(workspaceID)

		return nil
	})
}

func validateAddRepoInputs(workspaceID, repoName string) error {
	if err := validation.ValidateWorkspaceID(workspaceID); err != nil {
		return err
	}

	return validation.ValidateRepoName(repoName)
}

func repoExistsInWorkspace(repos []domain.Repo, repoName string) bool {
	for _, r := range repos {
		if r.Name == repoName {
			return true
		}
	}

	return false
}

func (s *Service) resolveWorkspaceRepo(workspaceID, repoName string) (domain.Repo, error) {
	repos, err := s.ResolveRepos(workspaceID, []string{repoName})
	if err != nil {
		var canopyErr *cerrors.CanopyError
		if isCanopyError(err, &canopyErr) {
			return domain.Repo{}, canopyErr.WithContext("operation", fmt.Sprintf("resolve repo %s", repoName))
		}

		return domain.Repo{}, cerrors.Wrap(cerrors.ErrUnknownRepository, fmt.Sprintf("failed to resolve repo %s", repoName), err)
	}

	return repos[0], nil
}

func (s *Service) workspaceBranchName(workspaceID, branchName string) (string, error) {
	if branchName == "" {
		return "", cerrors.NewMissingBranchConfig(workspaceID)
	}

	return branchName, nil
}

func (s *Service) ensureWorkspaceWorktree(ctx context.Context, repo domain.Repo, dirName, branchName string) error {
	if _, err := s.gitEngine.EnsureCanonical(ctx, repo.URL, repo.Name); err != nil {
		return cerrors.WrapGitError(err, fmt.Sprintf("ensure canonical for %s", repo.Name))
	}

	worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)
	if err := s.gitEngine.CreateWorktree(ctx, repo.Name, worktreePath, branchName); err != nil {
		return cerrors.WrapGitError(err, fmt.Sprintf("create worktree for %s", repo.Name))
	}

	return nil
}

func (s *Service) saveWorkspaceRepo(ctx context.Context, workspaceID string, workspace *domain.Workspace, repo domain.Repo) error {
	workspace.Repos = append(workspace.Repos, repo)
	if err := s.wsEngine.Save(ctx, *workspace); err != nil {
		return cerrors.NewWorkspaceMetadataError(workspaceID, "update", err)
	}

	return nil
}

// RemoveRepoFromWorkspace removes a repository from an existing workspace
func (s *Service) RemoveRepoFromWorkspace(ctx context.Context, workspaceID, repoName string) error {
	return s.withWorkspaceLock(ctx, workspaceID, false, func() error {
		workspace, _, err := s.findWorkspace(ctx, workspaceID)
		if err != nil {
			return err
		}

		// 2. Check if repo exists in workspace
		repoIndex := -1

		for i, r := range workspace.Repos {
			if r.Name == repoName {
				repoIndex = i
				break
			}
		}

		if repoIndex == -1 {
			return cerrors.NewRepoNotFound(repoName).WithContext("workspace_id", workspaceID)
		}

		// 3. Remove worktree directory
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID, repoName)
		if err := os.RemoveAll(worktreePath); err != nil {
			return cerrors.NewIOFailed(fmt.Sprintf("remove worktree %s", worktreePath), err)
		}

		// 4. Update metadata
		workspace.Repos = append(workspace.Repos[:repoIndex], workspace.Repos[repoIndex+1:]...)
		if err := s.wsEngine.Save(ctx, *workspace); err != nil {
			return cerrors.NewWorkspaceMetadataError(workspaceID, "update", err)
		}

		// Invalidate cache after metadata update
		s.cache.Invalidate(workspaceID)

		return nil
	})
}
