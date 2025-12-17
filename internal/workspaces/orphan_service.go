// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// OrphanService defines the interface for orphan detection operations.
type OrphanService interface {
	// DetectOrphans finds orphaned worktrees across all workspaces.
	DetectOrphans() ([]domain.OrphanedWorktree, error)

	// DetectOrphansForWorkspace returns orphans for a specific workspace.
	DetectOrphansForWorkspace(workspaceID string) ([]domain.OrphanedWorktree, error)

	// PruneAllWorktrees cleans up stale worktree references from all canonical repos.
	PruneAllWorktrees(ctx context.Context) error
}

// WorkspaceOrphanService handles orphan detection for workspaces.
type WorkspaceOrphanService struct {
	config          ports.ConfigProvider
	gitEngine       ports.GitOperations
	wsEngine        ports.WorkspaceStorage
	logger          *logging.Logger
	workspaceFinder WorkspaceFinder
}

// NewOrphanService creates a new WorkspaceOrphanService.
func NewOrphanService(
	cfg ports.ConfigProvider,
	gitEngine ports.GitOperations,
	wsEngine ports.WorkspaceStorage,
	logger *logging.Logger,
	finder WorkspaceFinder,
) *WorkspaceOrphanService {
	return &WorkspaceOrphanService{
		config:          cfg,
		gitEngine:       gitEngine,
		wsEngine:        wsEngine,
		logger:          logger,
		workspaceFinder: finder,
	}
}

// DetectOrphans finds orphaned worktrees across all workspaces.
// An orphan is a worktree reference in workspace metadata that:
// - References a canonical repo that no longer exists
// - Has a worktree directory that doesn't exist
// - Has an invalid git directory
func (s *WorkspaceOrphanService) DetectOrphans() ([]domain.OrphanedWorktree, error) {
	workspaceList, err := s.wsEngine.List(context.Background())
	if err != nil {
		return nil, cerrors.NewIOFailed("list workspaces", err)
	}

	canonicalSet, err := s.buildCanonicalRepoSet()
	if err != nil {
		return nil, err
	}

	var orphans []domain.OrphanedWorktree

	for _, ws := range workspaceList {
		wsOrphans := s.checkWorkspaceForOrphans(ws, ws.ID, canonicalSet)
		orphans = append(orphans, wsOrphans...)
	}

	return orphans, nil
}

// DetectOrphansForWorkspace returns orphans for a specific workspace.
// This is more efficient than DetectOrphans when only checking a single workspace.
func (s *WorkspaceOrphanService) DetectOrphansForWorkspace(workspaceID string) ([]domain.OrphanedWorktree, error) {
	ws, _, err := s.workspaceFinder.FindWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	canonicalSet, err := s.buildCanonicalRepoSet()
	if err != nil {
		return nil, err
	}

	return s.checkWorkspaceForOrphans(*ws, workspaceID, canonicalSet), nil
}

// buildCanonicalRepoSet returns a set of canonical repo names.
func (s *WorkspaceOrphanService) buildCanonicalRepoSet() (map[string]bool, error) {
	canonicalRepos, err := s.gitEngine.List(context.Background())
	if err != nil {
		return nil, cerrors.NewIOFailed("list canonical repos", err)
	}

	canonicalSet := make(map[string]bool)
	for _, r := range canonicalRepos {
		canonicalSet[r] = true
	}

	return canonicalSet, nil
}

// checkWorkspaceForOrphans checks a single workspace for orphaned worktrees.
func (s *WorkspaceOrphanService) checkWorkspaceForOrphans(
	ws domain.Workspace,
	dirName string,
	canonicalSet map[string]bool,
) []domain.OrphanedWorktree {
	var orphans []domain.OrphanedWorktree

	for _, repo := range ws.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)

		if orphan := s.checkRepoForOrphan(ws.ID, repo.Name, worktreePath, canonicalSet); orphan != nil {
			orphans = append(orphans, *orphan)
		}
	}

	return orphans
}

// checkRepoForOrphan checks if a single repo is orphaned. Returns nil if not orphaned.
// For non-existence errors (permission issues, I/O errors, etc.), it logs the error
// and returns nil rather than marking the repo as orphaned.
func (s *WorkspaceOrphanService) checkRepoForOrphan(
	workspaceID, repoName, worktreePath string,
	canonicalSet map[string]bool,
) *domain.OrphanedWorktree {
	// Check 1: Canonical repo exists
	if !canonicalSet[repoName] {
		return &domain.OrphanedWorktree{
			WorkspaceID:  workspaceID,
			RepoName:     repoName,
			WorktreePath: worktreePath,
			Reason:       domain.OrphanReasonCanonicalMissing,
		}
	}

	// Check 2: Worktree directory exists
	if _, err := os.Stat(worktreePath); err != nil {
		if os.IsNotExist(err) {
			return &domain.OrphanedWorktree{
				WorkspaceID:  workspaceID,
				RepoName:     repoName,
				WorktreePath: worktreePath,
				Reason:       domain.OrphanReasonDirectoryMissing,
			}
		}
		// Non-existence error (permission, I/O, etc.) - log and skip
		if s.logger != nil {
			s.logger.Warn("Unexpected error checking worktree directory",
				"workspace", workspaceID, "repo", repoName, "path", worktreePath, "error", err)
		}

		return nil
	}

	// Check 3: Valid git directory
	gitDir := filepath.Join(worktreePath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		if os.IsNotExist(err) {
			return &domain.OrphanedWorktree{
				WorkspaceID:  workspaceID,
				RepoName:     repoName,
				WorktreePath: worktreePath,
				Reason:       domain.OrphanReasonInvalidGitDir,
			}
		}
		// Non-existence error (permission, I/O, etc.) - log and skip
		if s.logger != nil {
			s.logger.Warn("Unexpected error checking .git directory",
				"workspace", workspaceID, "repo", repoName, "path", gitDir, "error", err)
		}

		return nil
	}

	return nil
}

// PruneAllWorktrees cleans up stale worktree references from all canonical repos.
// This removes worktree entries that point to non-existent directories.
func (s *WorkspaceOrphanService) PruneAllWorktrees(ctx context.Context) error {
	repos, err := s.gitEngine.List(ctx)
	if err != nil {
		return cerrors.NewIOFailed("list canonical repos", err)
	}

	var pruneErrors []error

	for _, repoName := range repos {
		if err := s.gitEngine.PruneWorktrees(ctx, repoName); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to prune worktrees",
					"repo", repoName,
					"error", err)
			}

			pruneErrors = append(pruneErrors, err)
		}
	}

	if len(pruneErrors) > 0 {
		return cerrors.NewInternalError("some worktree prune operations failed", pruneErrors[0])
	}

	return nil
}
