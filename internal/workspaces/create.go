package workspaces

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// CreateOptions configures workspace creation behavior.
type CreateOptions struct {
	SkipHooks         bool // Skip post_create hooks
	ContinueOnHookErr bool // Continue if hooks fail
}

// CreateWorkspace creates a new workspace directory and returns the directory name
func (s *Service) CreateWorkspace(ctx context.Context, id, branchName string, repos []domain.Repo) (string, error) {
	return s.CreateWorkspaceWithOptions(ctx, id, branchName, repos, CreateOptions{})
}

// CreateWorkspaceWithOptions creates a new workspace with configurable options.
func (s *Service) CreateWorkspaceWithOptions(ctx context.Context, id, branchName string, repos []domain.Repo, opts CreateOptions) (string, error) {
	branchName, err := s.resolveCreateBranchName(id, branchName)
	if err != nil {
		return "", err
	}

	if err := s.withWorkspaceLock(ctx, id, true, func() error {
		return s.createWorkspaceWithOptionsUnlocked(ctx, id, branchName, repos, opts)
	}); err != nil {
		return id, err
	}

	return id, nil
}

func (s *Service) createWorkspaceWithOptionsUnlocked(ctx context.Context, id, branchName string, repos []domain.Repo, opts CreateOptions) error {
	if err := s.ensureWorkspaceAvailable(id); err != nil {
		return err
	}

	ws := domain.Workspace{
		ID:         id,
		BranchName: branchName,
		Repos:      repos,
	}

	if err := s.executeWorkspaceCreate(ctx, ws, repos); err != nil {
		return err
	}

	// Invalidate cache for this workspace ID
	s.cache.Invalidate(id)

	// Run post_create hooks
	//nolint:contextcheck // Hooks manage their own timeout context per-hook
	if err := s.runPostCreateHooks(id, id, branchName, repos, opts); err != nil {
		// Hook failures don't rollback the workspace (per design.md)
		// But we return the error if not continuing on hook errors
		return err
	}

	return nil
}

func (s *Service) resolveCreateBranchName(id, branchName string) (string, error) {
	// Validate inputs
	if err := validation.ValidateWorkspaceID(id); err != nil {
		return "", err
	}

	if err := validation.ValidateBranchName(branchName); err != nil {
		return "", err
	}

	// Default branch name is the workspace ID
	if branchName == "" {
		branchName = id
		// Validate the derived branch name (workspace IDs may contain chars invalid for git refs)
		if err := validation.ValidateBranchName(branchName); err != nil {
			return "", cerrors.NewInvalidArgument("workspace-id", "cannot be used as default branch name: "+err.Error())
		}
	}

	return branchName, nil
}

func (s *Service) executeWorkspaceCreate(ctx context.Context, ws domain.Workspace, repos []domain.Repo) error {
	workspacePath := filepath.Join(s.config.GetWorkspacesRoot(), ws.ID)
	op := NewOperation(s.logger)
	op.AddStep(func() error {
		if err := os.Mkdir(workspacePath, 0o750); err != nil {
			if os.IsExist(err) {
				entries, readErr := os.ReadDir(workspacePath)
				if readErr != nil {
					return cerrors.NewIOFailed("read workspace directory", readErr)
				}

				if len(entries) == 0 {
					return nil
				}

				if len(entries) == 1 && entries[0].Name() == lockFileName {
					return nil
				}

				return cerrors.NewWorkspaceExists(ws.ID)
			}

			return cerrors.NewIOFailed("create workspace directory", err)
		}

		return nil
	}, func() error {
		if err := os.RemoveAll(workspacePath); err != nil {
			return cerrors.NewIOFailed("remove workspace directory", err)
		}

		return nil
	})
	op.AddStep(func() error {
		return s.cloneWorkspaceRepos(ctx, repos, ws.ID, ws.BranchName)
	}, func() error {
		return s.removeWorkspaceRepoWorktrees(ctx, ws.ID, repos)
	})
	op.AddStep(func() error {
		return s.wsEngine.Create(ctx, ws)
	}, nil)

	return op.Execute()
}

func (s *Service) ensureWorkspaceAvailable(workspaceID string) error {
	metaPath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID, "workspace.yaml")
	if _, err := os.Stat(metaPath); err == nil {
		return cerrors.NewWorkspaceExists(workspaceID)
	} else if err != nil && !os.IsNotExist(err) {
		return cerrors.NewIOFailed("check workspace metadata", err)
	}

	return nil
}

func (s *Service) removeWorkspaceRepoWorktrees(ctx context.Context, workspaceID string, repos []domain.Repo) error {
	if s.gitEngine == nil {
		return cerrors.NewInternalError("git engine not initialized", nil)
	}

	var errs []error

	for _, repo := range repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID, repo.Name)
		if err := s.gitEngine.RemoveWorktree(ctx, repo.Name, worktreePath); err != nil {
			errs = append(errs, err)
		}
	}

	return joinErrors(errs...)
}

// cloneWorkspaceRepos clones all repositories for a workspace.
// It runs EnsureCanonical operations in parallel (bounded by config.parallel_workers)
// for performance, then creates worktrees sequentially (as they depend on the canonical).
func (s *Service) cloneWorkspaceRepos(ctx context.Context, repos []domain.Repo, dirName, branchName string) error {
	if len(repos) == 0 {
		return nil
	}

	if s.gitEngine == nil {
		return cerrors.NewInternalError("git engine not initialized", nil)
	}

	// Check for context cancellation before starting
	if ctx.Err() != nil {
		return cerrors.NewContextError(ctx, "create workspace", dirName)
	}

	// Run EnsureCanonical operations in parallel
	err := s.runParallelCanonical(ctx, repos, parallelCanonicalOptions{
		workers: s.config.GetParallelWorkers(),
	})
	if err != nil {
		return err
	}

	// Create worktrees sequentially (they depend on the canonical being ready)
	for _, repo := range repos {
		// Check for context cancellation before each operation
		if ctx.Err() != nil {
			return cerrors.NewContextError(ctx, "create workspace", dirName)
		}

		// Create worktree
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.CreateWorktree(ctx, repo.Name, worktreePath, branchName); err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("create worktree for %s", repo.Name))
		}
	}

	return nil
}

// runPostCreateHooks runs post_create hooks if configured and not skipped.
// Returns nil if hooks are skipped or succeed, error otherwise.
func (s *Service) runPostCreateHooks(id, dirName, branchName string, repos []domain.Repo, opts CreateOptions) error {
	if opts.SkipHooks {
		return nil
	}

	hooksConfig := s.config.GetHooks()
	if len(hooksConfig.PostCreate) == 0 {
		return nil
	}

	hookCtx := domain.HookContext{
		WorkspaceID:   id,
		WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
		BranchName:    branchName,
		Repos:         repos,
	}

	//nolint:contextcheck // Hooks manage their own timeout context per-hook
	if _, err := s.hookExecutor.ExecuteHooks(hooksConfig.PostCreate, hookCtx, ports.HookExecuteOptions{
		ContinueOnError: opts.ContinueOnHookErr,
	}); err != nil {
		s.logger.Error("post_create hooks failed", "error", err)

		if !opts.ContinueOnHookErr {
			return err
		}
	}

	return nil
}
