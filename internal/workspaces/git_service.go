// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// defaultMaxParallel is the maximum number of parallel git operations.
const defaultMaxParallel = 8

// GitService defines the interface for git operations on workspaces.
type GitService interface {
	// PushWorkspace pushes all repos for a workspace.
	PushWorkspace(ctx context.Context, workspaceID string) error

	// RunGitInWorkspace executes an arbitrary git command across all repos in a workspace.
	RunGitInWorkspace(ctx context.Context, workspaceID string, args []string, opts GitRunOptions) ([]RepoGitResult, error)

	// SwitchBranch switches the branch for all repos in a workspace.
	SwitchBranch(ctx context.Context, workspaceID, branchName string, create bool) error
}

// WorkspaceFinder is the interface for finding workspaces (used to avoid circular dependencies).
type WorkspaceFinder interface {
	FindWorkspace(workspaceID string) (*domain.Workspace, string, error)
}

// WorkspaceGitService handles git operations for workspaces.
type WorkspaceGitService struct {
	config          ports.ConfigProvider
	gitEngine       ports.GitOperations
	wsEngine        ports.WorkspaceStorage
	logger          *logging.Logger
	cache           ports.WorkspaceCache
	workspaceFinder WorkspaceFinder
}

// NewGitService creates a new WorkspaceGitService.
func NewGitService(
	cfg ports.ConfigProvider,
	gitEngine ports.GitOperations,
	wsEngine ports.WorkspaceStorage,
	logger *logging.Logger,
	cache ports.WorkspaceCache,
	finder WorkspaceFinder,
) *WorkspaceGitService {
	return &WorkspaceGitService{
		config:          cfg,
		gitEngine:       gitEngine,
		wsEngine:        wsEngine,
		logger:          logger,
		cache:           cache,
		workspaceFinder: finder,
	}
}

// PushWorkspace pushes all repos for a workspace.
func (s *WorkspaceGitService) PushWorkspace(ctx context.Context, workspaceID string) error {
	targetWorkspace, dirName, err := s.workspaceFinder.FindWorkspace(workspaceID)
	if err != nil {
		return err
	}

	for _, repo := range targetWorkspace.Repos {
		// Check for context cancellation before each push
		if ctx.Err() != nil {
			return cerrors.NewContextError(ctx, "push workspace", workspaceID)
		}

		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
		branchName := targetWorkspace.BranchName

		if branchName == "" {
			if s.logger != nil {
				s.logger.Debug("Branch missing in metadata, will let git infer", "workspace", workspaceID, "repo", repo.Name)
			}
		}

		if err := s.gitEngine.Push(ctx, worktreePath, branchName); err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("push repo %s", repo.Name))
		}
	}

	return nil
}

// RunGitInWorkspace executes an arbitrary git command across all repos in a workspace.
func (s *WorkspaceGitService) RunGitInWorkspace(ctx context.Context, workspaceID string, args []string, opts GitRunOptions) ([]RepoGitResult, error) {
	targetWorkspace, dirName, err := s.workspaceFinder.FindWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	if len(targetWorkspace.Repos) == 0 {
		return nil, nil
	}

	if opts.Parallel {
		return s.runGitParallel(ctx, targetWorkspace, dirName, args, opts.ContinueOnError)
	}

	return s.runGitSequential(ctx, targetWorkspace, dirName, args, opts.ContinueOnError)
}

func (s *WorkspaceGitService) runGitSequential(ctx context.Context, workspace *domain.Workspace, dirName string, args []string, continueOnError bool) ([]RepoGitResult, error) {
	var results []RepoGitResult

	for _, repo := range workspace.Repos {
		// Check for context cancellation between iterations
		if ctx.Err() != nil {
			return results, cerrors.NewContextError(ctx, "git command", "sequential execution")
		}

		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)

		cmdResult, err := s.gitEngine.RunCommand(ctx, worktreePath, args...)

		result := RepoGitResult{
			RepoName: repo.Name,
		}

		if err != nil {
			result.Error = err
			results = append(results, result)

			if !continueOnError {
				return results, err
			}

			continue
		}

		result.Stdout = cmdResult.Stdout
		result.Stderr = cmdResult.Stderr
		result.ExitCode = cmdResult.ExitCode
		results = append(results, result)

		if cmdResult.ExitCode != 0 && !continueOnError {
			return results, cerrors.NewCommandFailed(fmt.Sprintf("git in repo %s", repo.Name), fmt.Errorf("exit code %d", cmdResult.ExitCode))
		}
	}

	return results, nil
}

func (s *WorkspaceGitService) runGitParallel(ctx context.Context, workspace *domain.Workspace, dirName string, args []string, continueOnError bool) ([]RepoGitResult, error) {
	results := make([]RepoGitResult, len(workspace.Repos))

	var wg sync.WaitGroup

	// Bounded worker pool to avoid exhausting resources for large workspaces
	sem := make(chan struct{}, defaultMaxParallel)

	// Create a cancellable context for goroutines
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track the first error for early termination
	var (
		firstErr     error
		firstErrOnce sync.Once
	)

	for i, repo := range workspace.Repos {
		wg.Add(1)

		go func(idx int, r domain.Repo) {
			defer wg.Done()

			// Check if context is cancelled before acquiring semaphore
			select {
			case <-cancelCtx.Done():
				ctxErr := cerrors.NewContextError(cancelCtx, "git command", r.Name)
				results[idx] = RepoGitResult{
					RepoName: r.Name,
					Error:    ctxErr,
				}

				// Propagate cancellation error to caller if not continuing on error
				if !continueOnError {
					firstErrOnce.Do(func() {
						firstErr = ctxErr
					})
				}

				return
			case sem <- struct{}{}:
			}

			defer func() { <-sem }()

			worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, r.Name)

			result := RepoGitResult{
				RepoName: r.Name,
			}

			cmdResult, err := s.gitEngine.RunCommand(cancelCtx, worktreePath, args...)
			if err != nil {
				result.Error = err
				results[idx] = result

				// Cancel other goroutines on first error if not continuing
				if !continueOnError {
					firstErrOnce.Do(func() {
						firstErr = err

						cancel()
					})
				}

				return
			}

			result.Stdout = cmdResult.Stdout
			result.Stderr = cmdResult.Stderr
			result.ExitCode = cmdResult.ExitCode
			results[idx] = result

			// Also cancel on non-zero exit code if not continuing
			if result.ExitCode != 0 && !continueOnError {
				firstErrOnce.Do(func() {
					firstErr = cerrors.NewCommandFailed(fmt.Sprintf("git in repo %s", r.Name), fmt.Errorf("exit code %d", result.ExitCode))

					cancel()
				})
			}
		}(i, repo)
	}

	wg.Wait()

	// Return the first error that triggered cancellation
	if firstErr != nil {
		return results, firstErr
	}

	return results, nil
}

// SwitchBranch switches the branch for all repos in a workspace.
func (s *WorkspaceGitService) SwitchBranch(_ context.Context, workspaceID, branchName string, create bool) error {
	targetWorkspace, dirName, err := s.workspaceFinder.FindWorkspace(workspaceID)
	if err != nil {
		return err
	}

	// Iterate through repos and checkout
	for _, repo := range targetWorkspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)

		if s.logger != nil {
			s.logger.Info("Switching branch", "repo", repo.Name, "branch", branchName)
		}

		if err := s.gitEngine.Checkout(worktreePath, branchName, create); err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("checkout branch %s in repo %s", branchName, repo.Name))
		}
	}

	// Update metadata
	targetWorkspace.BranchName = branchName
	if err := s.wsEngine.Save(dirName, *targetWorkspace); err != nil {
		return cerrors.NewWorkspaceMetadataError(workspaceID, "update", err)
	}

	// Invalidate cache after metadata update
	if s.cache != nil {
		s.cache.Invalidate(workspaceID)
	}

	return nil
}
