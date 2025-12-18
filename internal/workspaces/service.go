// Package workspaces provides the core business logic for workspace management.
//
// This package implements the central orchestration layer for all workspace
// operations including creation, closure, repository management, and status
// reporting. It follows hexagonal architecture principles, depending on
// interfaces (ports) rather than concrete implementations.
//
// # Key Operations
//
// Workspace lifecycle:
//   - CreateWorkspace: Creates a new workspace with repositories
//   - CloseWorkspace: Removes a workspace (with optional archival)
//   - ReopenWorkspace: Restores an archived workspace
//   - RenameWorkspace: Renames workspace and associated branches
//
// Repository operations:
//   - AddRepoToWorkspace: Adds a repository to an existing workspace
//   - RemoveRepoFromWorkspace: Removes a repository from a workspace
//   - ResolveRepos: Resolves repository names to URL/name pairs
//
// Status and queries:
//   - ListWorkspaces: Lists all active workspaces
//   - GetWorkspaceStatus: Returns git status for all repos in a workspace
//   - WorkspacePath: Returns the filesystem path for a workspace
//
// # Service Options
//
// The Service can be configured with functional options:
//
//	svc := workspaces.NewService(cfg, git, storage, logger,
//	    workspaces.WithHookExecutor(customExecutor),
//	    workspaces.WithCache(customCache),
//	)
//
// # Thread Safety
//
// The Service is safe for concurrent use. Individual operations acquire
// appropriate locks and the internal cache handles concurrent access.
package workspaces

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/hooks"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
	"github.com/alexisbeaulieu97/canopy/internal/validation"
)

// Service manages workspace operations
type Service struct {
	config       ports.ConfigProvider
	gitEngine    ports.GitOperations
	wsEngine     ports.WorkspaceStorage
	logger       *logging.Logger
	hookExecutor ports.HookExecutor

	// Sub-services for specific responsibilities
	resolver  *RepoResolver
	diskUsage ports.DiskUsage
	canonical *CanonicalRepoService

	// cache provides in-memory caching of workspace metadata
	cache ports.WorkspaceCache

	// Extracted sub-services
	gitService    *WorkspaceGitService
	orphanService *WorkspaceOrphanService
	exportService *WorkspaceExportService
}

// ErrNoReposConfigured indicates no repos were specified and none matched configuration.
var ErrNoReposConfigured = errors.New("no repositories specified and no patterns matched")

// HookPhase identifies which lifecycle hook set to execute.
type HookPhase string

const (
	// HookPhasePostCreate executes post_create hooks.
	HookPhasePostCreate HookPhase = "post_create"
	// HookPhasePreClose executes pre_close hooks.
	HookPhasePreClose HookPhase = "pre_close"
)

// ServiceOption is a functional option for configuring the Service.
type ServiceOption func(*serviceOptions)

// serviceOptions holds optional dependencies that can be injected.
type serviceOptions struct {
	hookExecutor ports.HookExecutor
	diskUsage    ports.DiskUsage
	cache        ports.WorkspaceCache
}

// WithHookExecutor sets a custom HookExecutor implementation.
func WithHookExecutor(h ports.HookExecutor) ServiceOption {
	return func(o *serviceOptions) {
		o.hookExecutor = h
	}
}

// WithDiskUsage sets a custom DiskUsage implementation.
func WithDiskUsage(d ports.DiskUsage) ServiceOption {
	return func(o *serviceOptions) {
		o.diskUsage = d
	}
}

// WithCache sets a custom WorkspaceCache implementation.
func WithCache(c ports.WorkspaceCache) ServiceOption {
	return func(o *serviceOptions) {
		o.cache = c
	}
}

// NewService creates a new workspace service.
// Options can be provided to override default implementations for testing.
func NewService(cfg ports.ConfigProvider, gitEngine ports.GitOperations, wsEngine ports.WorkspaceStorage, logger *logging.Logger, opts ...ServiceOption) *Service {
	// Apply all options
	options := &serviceOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Use provided hook executor or create default
	hookExecutor := options.hookExecutor
	if hookExecutor == nil {
		hookExecutor = hooks.NewExecutor(logger)
	}

	// Use provided disk usage or create default
	diskUsage := options.diskUsage
	if diskUsage == nil {
		diskUsage = DefaultDiskUsageCalculator()
	}

	// Use provided cache or create default
	cache := options.cache
	if cache == nil {
		cache = NewWorkspaceCache(DefaultCacheTTL)
	}

	svc := &Service{
		config:       cfg,
		gitEngine:    gitEngine,
		wsEngine:     wsEngine,
		logger:       logger,
		hookExecutor: hookExecutor,
		resolver:     NewRepoResolver(cfg.GetRegistry()),
		diskUsage:    diskUsage,
		canonical:    NewCanonicalRepoService(gitEngine, wsEngine, cfg.GetProjectsRoot(), logger, diskUsage),
		cache:        cache,
	}

	// Initialize sub-services with the main service as the workspace finder/creator
	svc.gitService = NewGitService(cfg, gitEngine, wsEngine, logger, cache, svc)
	svc.orphanService = NewOrphanService(cfg, gitEngine, wsEngine, logger, svc)
	svc.exportService = NewExportService(cfg, svc, svc)

	return svc
}

// FindWorkspace implements WorkspaceFinder interface for sub-services.
func (s *Service) FindWorkspace(workspaceID string) (*domain.Workspace, string, error) {
	return s.findWorkspace(context.Background(), workspaceID)
}

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

	ws := domain.Workspace{
		ID:         id,
		BranchName: branchName,
		Repos:      repos,
	}

	if err := s.wsEngine.Create(ctx, ws); err != nil {
		return "", err
	}

	// Invalidate cache for this workspace ID
	s.cache.Invalidate(id)

	// Manual cleanup helper
	cleanup := func() {
		path := filepath.Join(s.config.GetWorkspacesRoot(), id)
		if err := os.RemoveAll(path); err != nil {
			s.logger.Warn("cleanup failed", "path", path, "error", err)
		}
	}

	// Clone repositories (if any)
	if err := s.cloneWorkspaceRepos(ctx, repos, id, branchName); err != nil {
		cleanup()
		return "", err
	}

	// Run post_create hooks
	//nolint:contextcheck // Hooks manage their own timeout context per-hook
	if err := s.runPostCreateHooks(id, id, branchName, repos, opts); err != nil {
		// Hook failures don't rollback the workspace (per design.md)
		// But we return the error if not continuing on hook errors
		return id, err
	}

	return id, nil
}

// cloneWorkspaceRepos clones all repositories for a workspace.
// It runs EnsureCanonical operations in parallel (bounded by config.parallel_workers)
// for performance, then creates worktrees sequentially (as they depend on the canonical).
func (s *Service) cloneWorkspaceRepos(ctx context.Context, repos []domain.Repo, dirName, branchName string) error {
	if len(repos) == 0 {
		return nil
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
	if err := s.hookExecutor.ExecuteHooks(hooksConfig.PostCreate, hookCtx, opts.ContinueOnHookErr); err != nil {
		s.logger.Error("post_create hooks failed", "error", err)

		if !opts.ContinueOnHookErr {
			return err
		}
	}

	return nil
}

// WorkspacePath returns the absolute path for a workspace ID.
func (s *Service) WorkspacePath(workspaceID string) (string, error) {
	workspaces, err := s.wsEngine.List(context.Background())
	if err != nil {
		return "", cerrors.NewIOFailed("list workspaces", err)
	}

	for _, w := range workspaces {
		if w.ID == workspaceID {
			return filepath.Join(s.config.GetWorkspacesRoot(), w.ID), nil
		}
	}

	return "", cerrors.NewWorkspaceNotFound(workspaceID)
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
	if errors.Is(err, cerrors.WorkspaceNotFound) {
		return nil
	}

	// For any other error (IO failure, etc.), propagate it
	return err
}

// handleClosedWorkspaceError checks if the error is due to a closed workspace and returns a more helpful error message.
func (s *Service) handleClosedWorkspaceError(ctx context.Context, workspaceID string, err error) error {
	if !errors.Is(err, cerrors.WorkspaceNotFound) {
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

	isWorkspaceExists := errors.As(existingErr, &canopyErr) && canopyErr.Code == cerrors.ErrWorkspaceExists

	if !force || !isWorkspaceExists {
		return existingErr
	}

	// Force mode: delete the existing workspace
	if deleteErr := s.wsEngine.Delete(ctx, newID); deleteErr != nil {
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
	var joined error

	for _, repo := range workspace.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.RenameBranch(ctx, worktreePath, newID, oldID); err != nil {
			joined = errors.Join(joined, cerrors.WrapGitError(err, fmt.Sprintf("rollback branch rename in repo %s", repo.Name)))
		}
	}

	return joined
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
func (s *Service) renameWorkspaceDir(ctx context.Context, workspace domain.Workspace, oldID, newID string, shouldRenameBranch bool) error {
	if err := s.wsEngine.Rename(ctx, oldID, newID); err != nil {
		if shouldRenameBranch {
			s.rollbackBranchRenames(ctx, workspace, oldID, oldID, newID)
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
func (s *Service) updateBranchMetadataWithRollback(ctx context.Context, workspace domain.Workspace, oldID, newID string) error {
	if err := s.updateBranchMetadata(ctx, newID, newID); err != nil {
		var rollbackErrors []error

		// Attempt to rollback branch renames first so repo state aligns with directory rollback.
		if branchRollbackErr := s.rollbackBranchRenamesWithError(ctx, workspace, newID, oldID, newID); branchRollbackErr != nil {
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
			return errors.Join(append([]error{err}, rollbackErrors...)...)
		}

		return err
	}

	return nil
}

// executeRename performs the actual rename operations: branch rename, directory rename, and metadata update.
func (s *Service) executeRename(ctx context.Context, workspace domain.Workspace, oldID, newID string, shouldRenameBranch bool) error {
	if shouldRenameBranch {
		if err := s.renameBranchesInRepos(ctx, workspace, oldID, oldID, newID); err != nil {
			return err
		}
	}

	if err := s.renameWorkspaceDir(ctx, workspace, oldID, newID, shouldRenameBranch); err != nil {
		return err
	}

	if shouldRenameBranch {
		if err := s.updateBranchMetadataWithRollback(ctx, workspace, oldID, newID); err != nil {
			return err
		}
	}

	return nil
}

// RenameWorkspace renames a workspace to a new ID.
// If renameBranch is true and the branch name matches the old ID, it will also rename branches.
// If force is true, an existing workspace with the new ID will be deleted first.
func (s *Service) RenameWorkspace(ctx context.Context, oldID, newID string, renameBranch, force bool) error {
	workspace, _, err := s.findWorkspace(ctx, oldID)
	if err != nil {
		return s.handleClosedWorkspaceError(ctx, oldID, err)
	}

	if err := s.validateRenameInputs(newID); err != nil {
		return err
	}

	if oldID == newID {
		return cerrors.NewInvalidArgument("new_id", "cannot rename workspace to the same ID")
	}

	if err := s.ensureTargetAvailableOrDelete(ctx, newID, force); err != nil {
		return err
	}

	shouldRenameBranch := renameBranch && workspace.BranchName == oldID

	if err := s.executeRename(ctx, *workspace, oldID, newID, shouldRenameBranch); err != nil {
		return err
	}

	s.invalidateWorkspaceCache(oldID, newID)

	return nil
}

// AddRepoToWorkspace adds a repository to an existing workspace
func (s *Service) AddRepoToWorkspace(ctx context.Context, workspaceID, repoName string) error {
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

	if err := s.ensureWorkspaceWorktree(ctx, repo, workspaceID, branchName); err != nil {
		return err
	}

	if err := s.saveWorkspaceRepo(ctx, workspaceID, workspace, repo); err != nil {
		return err
	}

	s.cache.Invalidate(workspaceID)

	return nil
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
		if errors.As(err, &canopyErr) {
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
}

// CloseOptions configures workspace close behavior.
type CloseOptions struct {
	SkipHooks         bool // Skip pre_close hooks
	ContinueOnHookErr bool // Continue if hooks fail
}

// SyncOptions configures workspace sync behavior.
type SyncOptions struct {
	Timeout time.Duration
}

// CloseWorkspace removes a workspace with safety checks
func (s *Service) CloseWorkspace(_ context.Context, workspaceID string, force bool) error {
	//nolint:contextcheck // Wrapper delegates to WithOptions which handles hooks with own timeout
	return s.CloseWorkspaceWithOptions(workspaceID, force, CloseOptions{})
}

// CloseWorkspaceWithOptions removes a workspace with configurable options.
//
//nolint:contextcheck // This function manages hook contexts internally with their own timeouts
func (s *Service) CloseWorkspaceWithOptions(workspaceID string, force bool, opts CloseOptions) error {
	targetWorkspace, _, err := s.findWorkspace(context.Background(), workspaceID)
	if err != nil {
		return err
	}

	if !force {
		if err := s.ensureWorkspaceClean(targetWorkspace, workspaceID, "close"); err != nil {
			return err
		}
	}

	// Run pre_close hooks before deletion
	if !opts.SkipHooks {
		hooksConfig := s.config.GetHooks()
		if len(hooksConfig.PreClose) > 0 {
			hookCtx := domain.HookContext{
				WorkspaceID:   workspaceID,
				WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), workspaceID),
				BranchName:    targetWorkspace.BranchName,
				Repos:         targetWorkspace.Repos,
			}

			//nolint:contextcheck // Hooks manage their own timeout context per-hook
			if err := s.hookExecutor.ExecuteHooks(hooksConfig.PreClose, hookCtx, opts.ContinueOnHookErr); err != nil {
				s.logger.Error("pre_close hooks failed", "error", err)
				// Per design.md: pre_close failure aborts close operation
				if !opts.ContinueOnHookErr {
					return err
				}
			}
		}
	}

	// Remove worktrees from canonical repos before deleting workspace directory
	s.removeWorkspaceWorktrees(targetWorkspace, workspaceID)

	// Delete workspace
	if err := s.wsEngine.Delete(context.Background(), workspaceID); err != nil {
		return err
	}

	// Invalidate cache after workspace deletion
	s.cache.Invalidate(workspaceID)

	return nil
}

// CloseWorkspaceKeepMetadata moves workspace metadata to the closed store and removes the active worktree.
func (s *Service) CloseWorkspaceKeepMetadata(_ context.Context, workspaceID string, force bool) (*domain.ClosedWorkspace, error) {
	//nolint:contextcheck // Wrapper delegates to WithOptions which handles hooks with own timeout
	return s.CloseWorkspaceKeepMetadataWithOptions(workspaceID, force, CloseOptions{})
}

// CloseWorkspaceKeepMetadataWithOptions moves workspace metadata to the closed store with configurable options.
//
//nolint:contextcheck // This function manages hook contexts internally with their own timeouts
func (s *Service) CloseWorkspaceKeepMetadataWithOptions(workspaceID string, force bool, opts CloseOptions) (*domain.ClosedWorkspace, error) {
	targetWorkspace, _, err := s.findWorkspace(context.Background(), workspaceID)
	if err != nil {
		return nil, err
	}

	if !force {
		if err := s.ensureWorkspaceClean(targetWorkspace, workspaceID, "close"); err != nil {
			return nil, err
		}
	}

	// Run pre_close hooks before archiving
	if !opts.SkipHooks {
		hooksConfig := s.config.GetHooks()
		if len(hooksConfig.PreClose) > 0 {
			hookCtx := domain.HookContext{
				WorkspaceID:   workspaceID,
				WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), workspaceID),
				BranchName:    targetWorkspace.BranchName,
				Repos:         targetWorkspace.Repos,
			}

			//nolint:contextcheck // Hooks manage their own timeout context per-hook
			if err := s.hookExecutor.ExecuteHooks(hooksConfig.PreClose, hookCtx, opts.ContinueOnHookErr); err != nil {
				s.logger.Error("pre_close hooks failed", "error", err)
				// Per design.md: pre_close failure aborts close operation
				if !opts.ContinueOnHookErr {
					return nil, err
				}
			}
		}
	}

	closedAt := time.Now().UTC()

	archived, err := s.wsEngine.Close(context.Background(), workspaceID, closedAt)
	if err != nil {
		return nil, err
	}

	// Remove worktrees from canonical repos before deleting workspace directory
	s.removeWorkspaceWorktrees(targetWorkspace, workspaceID)

	if err := s.wsEngine.Delete(context.Background(), workspaceID); err != nil {
		_ = s.wsEngine.DeleteClosed(context.Background(), workspaceID, closedAt)
		return nil, cerrors.NewIOFailed("remove workspace directory", err)
	}

	// Invalidate cache after workspace deletion
	s.cache.Invalidate(workspaceID)

	return archived, nil
}

// RunHooks executes lifecycle hooks for an existing workspace without performing other actions.
//
//nolint:contextcheck // Hooks manage their own timeout context per-hook
func (s *Service) RunHooks(workspaceID string, phase HookPhase, continueOnError bool) error {
	workspace, _, err := s.findWorkspace(context.Background(), workspaceID)
	if err != nil {
		return err
	}

	hooksConfig := s.config.GetHooks()

	var selected []config.Hook

	switch phase {
	case HookPhasePostCreate:
		selected = hooksConfig.PostCreate
	case HookPhasePreClose:
		selected = hooksConfig.PreClose
	default:
		return cerrors.NewInvalidArgument("hook_phase", fmt.Sprintf("unsupported hook phase %q", phase))
	}

	if len(selected) == 0 {
		return nil
	}

	hookCtx := domain.HookContext{
		WorkspaceID:   workspaceID,
		WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), workspaceID),
		BranchName:    workspace.BranchName,
		Repos:         workspace.Repos,
	}

	if err := s.hookExecutor.ExecuteHooks(selected, hookCtx, continueOnError); err != nil {
		if s.logger != nil {
			s.logger.Error(fmt.Sprintf("%s hooks failed", phase), "error", err)
		}

		if !continueOnError {
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

// ListWorkspaces returns all active workspaces
func (s *Service) ListWorkspaces() ([]domain.Workspace, error) {
	workspaceList, err := s.wsEngine.List(context.Background())
	if err != nil {
		return nil, err
	}

	var workspaces []domain.Workspace

	for _, w := range workspaceList {
		wsPath := filepath.Join(s.config.GetWorkspacesRoot(), w.ID)

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

// ListClosedWorkspaces returns closed workspace metadata.
func (s *Service) ListClosedWorkspaces() ([]domain.ClosedWorkspace, error) {
	return s.wsEngine.ListClosed(context.Background())
}

// GetStatus returns the aggregate status of a workspace
func (s *Service) GetStatus(workspaceID string) (*domain.WorkspaceStatus, error) {
	targetWorkspace, _, err := s.findWorkspace(context.Background(), workspaceID)
	if err != nil {
		return nil, err
	}

	// 2. Check status for each repo
	var repoStatuses []domain.RepoStatus

	for _, repo := range targetWorkspace.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID, repo.Name)

		isDirty, unpushed, behind, branch, err := s.gitEngine.Status(context.Background(), worktreePath)
		if err != nil {
			repoStatuses = append(repoStatuses, domain.RepoStatus{
				Name:   repo.Name,
				Branch: "ERROR: " + err.Error(),
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

// SyncWorkspace pulls updates for all repositories in the workspace.
func (s *Service) SyncWorkspace(ctx context.Context, id string, opts SyncOptions) (*domain.SyncResult, error) {
	ws, _, err := s.findWorkspace(ctx, id)
	if err != nil {
		return nil, err
	}

	if opts.Timeout == 0 {
		opts.Timeout = 60 * time.Second // Default timeout
	}

	results := make([]domain.RepoSyncStatus, len(ws.Repos))

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	numWorkers := s.config.GetParallelWorkers()
	if numWorkers <= 0 {
		numWorkers = 1
	}

	reposChan := make(chan struct {
		index int
		repo  domain.Repo
	}, len(ws.Repos))

	for i, repo := range ws.Repos {
		reposChan <- struct {
			index int
			repo  domain.Repo
		}{i, repo}
	}

	close(reposChan)

	if numWorkers > len(ws.Repos) {
		numWorkers = len(ws.Repos)
	}

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for r := range reposChan {
				repoResult := s.syncRepo(ctx, id, r.repo, opts.Timeout)

				mu.Lock()

				results[r.index] = repoResult

				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return s.aggregateSyncResults(id, results), nil
}

func (s *Service) aggregateSyncResults(workspaceID string, results []domain.RepoSyncStatus) *domain.SyncResult {
	syncResult := &domain.SyncResult{
		WorkspaceID: workspaceID,
		Repos:       results,
	}

	for _, r := range results {
		if r.Status == domain.SyncStatusUpdated {
			syncResult.TotalUpdated += r.Updated
		}

		if r.Status == domain.SyncStatusError || r.Status == domain.SyncStatusTimeout || r.Status == domain.SyncStatusConflict {
			syncResult.TotalErrors++
		}
	}

	return syncResult
}

func (s *Service) syncRepo(ctx context.Context, wsID string, repo domain.Repo, timeout time.Duration) domain.RepoSyncStatus {
	result := domain.RepoSyncStatus{
		Name:   repo.Name,
		Status: domain.SyncStatusUpToDate,
	}

	repoCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. Fetch canonical
	if err := s.gitEngine.Fetch(repoCtx, repo.Name); err != nil {
		result.Updated = 0
		if errors.Is(err, context.DeadlineExceeded) {
			result.Status = domain.SyncStatusTimeout
			result.Error = "timeout during fetch"

			return result
		}

		result.Status = domain.SyncStatusError
		result.Error = fmt.Sprintf("fetch failed: %v", err)

		return result
	}

	worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), wsID, repo.Name)

	// 2. Get status before pull to see behind count
	_, _, behind, _, err := s.gitEngine.Status(repoCtx, worktreePath)
	if err != nil {
		result.Status = domain.SyncStatusError
		result.Error = fmt.Sprintf("status failed: %v", err)
		result.Updated = 0

		return result
	}

	result.Updated = behind

	// 3. Pull worktree only if behind remote
	if result.Updated > 0 {
		err = s.gitEngine.Pull(repoCtx, worktreePath)
		if err != nil {
			result.Updated = 0
			if errors.Is(err, context.DeadlineExceeded) {
				result.Status = domain.SyncStatusTimeout
				result.Error = "timeout during pull"

				return result
			}

			// Check for conflicts - Usually go-git returns error if pull cannot be done cleanly.
			// For now we treat it as error, but we could improve detection.
			result.Status = domain.SyncStatusError
			result.Error = fmt.Sprintf("pull failed: %v", err)

			return result
		}

		result.Status = domain.SyncStatusUpdated
	}

	return result
}

// ListCanonicalRepos returns a list of all cached repositories
func (s *Service) ListCanonicalRepos(ctx context.Context) ([]string, error) {
	return s.canonical.List(ctx)
}

// AddCanonicalRepo adds a new repository to the cache and returns the canonical name.
func (s *Service) AddCanonicalRepo(ctx context.Context, url string) (string, error) {
	return s.canonical.Add(ctx, url)
}

// RemoveCanonicalRepo removes a repository from the cache
func (s *Service) RemoveCanonicalRepo(ctx context.Context, name string, force bool) error {
	return s.canonical.Remove(ctx, name, force)
}

// PreviewRemoveCanonicalRepo returns a preview of what would happen when removing a repo.
func (s *Service) PreviewRemoveCanonicalRepo(ctx context.Context, name string) (*domain.RepoRemovePreview, error) {
	return s.canonical.PreviewRemove(ctx, name)
}

// SyncCanonicalRepo fetches updates for a cached repository
func (s *Service) SyncCanonicalRepo(ctx context.Context, name string) error {
	return s.canonical.Sync(ctx, name)
}

// PushWorkspace pushes all repos for a workspace.
func (s *Service) PushWorkspace(ctx context.Context, workspaceID string) error {
	return s.gitService.PushWorkspace(ctx, workspaceID)
}

// GitRunOptions contains options for running git commands across workspace repos.
type GitRunOptions struct {
	Parallel        bool
	ContinueOnError bool
}

// RepoGitResult holds the result of running a git command in a single repo.
type RepoGitResult struct {
	RepoName string
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// RunGitInWorkspace executes an arbitrary git command across all repos in a workspace.
func (s *Service) RunGitInWorkspace(ctx context.Context, workspaceID string, args []string, opts GitRunOptions) ([]RepoGitResult, error) {
	return s.gitService.RunGitInWorkspace(ctx, workspaceID, args, opts)
}

// SwitchBranch switches the branch for all repos in a workspace
func (s *Service) SwitchBranch(ctx context.Context, workspaceID, branchName string, create bool) error {
	return s.gitService.SwitchBranch(ctx, workspaceID, branchName, create)
}

// RestoreWorkspace recreates a workspace from the newest closed entry.
func (s *Service) RestoreWorkspace(ctx context.Context, workspaceID string, force bool) error {
	archive, err := s.wsEngine.LatestClosed(ctx, workspaceID)
	if err != nil {
		return err
	}

	if _, _, err := s.findWorkspace(ctx, workspaceID); err == nil {
		if !force {
			return cerrors.NewWorkspaceExists(workspaceID).WithContext("hint", "Use --force to replace or choose a different ID")
		}

		if err := s.CloseWorkspace(ctx, workspaceID, true); err != nil {
			return cerrors.NewIOFailed("remove existing workspace", err)
		}
	}

	ws := archive.Metadata
	ws.ClosedAt = nil

	if _, err := s.CreateWorkspace(ctx, ws.ID, ws.BranchName, ws.Repos); err != nil {
		// Preserve original error type if it's already typed
		var canopyErr *cerrors.CanopyError
		if errors.As(err, &canopyErr) {
			return canopyErr.WithContext("operation", fmt.Sprintf("restore workspace %s", workspaceID))
		}

		return cerrors.Wrap(cerrors.ErrIOFailed, fmt.Sprintf("failed to restore workspace %s", workspaceID), err)
	}

	// Delete the closed entry using ID and timestamp
	closedAt := archive.ClosedAt()
	if err := s.wsEngine.DeleteClosed(ctx, workspaceID, closedAt); err != nil {
		return cerrors.NewIOFailed("remove closed entry", err)
	}

	return nil
}

// StaleThresholdDays returns the configured stale threshold in days.
func (s *Service) StaleThresholdDays() int {
	return s.config.GetStaleThresholdDays()
}

// Keybindings returns the TUI keybindings configuration with defaults applied.
func (s *Service) Keybindings() config.Keybindings {
	return s.config.GetKeybindings()
}

// UseEmoji returns whether emoji should be used in the TUI.
func (s *Service) UseEmoji() bool {
	return s.config.GetUseEmoji()
}

// DetectOrphans finds orphaned worktrees across all workspaces.
// An orphan is a worktree reference in workspace metadata that:
// - References a canonical repo that no longer exists
// - Has a worktree directory that doesn't exist
// - Has an invalid git directory
func (s *Service) DetectOrphans() ([]domain.OrphanedWorktree, error) {
	return s.orphanService.DetectOrphans()
}

// GetWorkspacesUsingRepo returns the IDs of workspaces that use the given canonical repo.
func (s *Service) GetWorkspacesUsingRepo(ctx context.Context, repoName string) ([]string, error) {
	return s.canonical.GetWorkspacesUsingRepo(ctx, repoName)
}

// DetectOrphansForWorkspace returns orphans for a specific workspace.
// This is more efficient than DetectOrphans when only checking a single workspace.
func (s *Service) DetectOrphansForWorkspace(workspaceID string) ([]domain.OrphanedWorktree, error) {
	return s.orphanService.DetectOrphansForWorkspace(workspaceID)
}

// PruneAllWorktrees cleans up stale worktree references from all canonical repos.
// This removes worktree entries that point to non-existent directories.
func (s *Service) PruneAllWorktrees(ctx context.Context) error {
	return s.orphanService.PruneAllWorktrees(ctx)
}

func (s *Service) findWorkspace(ctx context.Context, workspaceID string) (*domain.Workspace, string, error) {
	// Check cache first
	if ws, dirName, ok := s.cache.Get(workspaceID); ok {
		return ws, dirName, nil
	}

	// Cache miss: use direct lookup via Load (ID-based)
	ws, err := s.wsEngine.Load(ctx, workspaceID)
	if err != nil {
		return nil, "", err
	}

	// Populate cache with the result (dirName is now the same as ID)
	s.cache.Set(workspaceID, ws, workspaceID)

	return ws, workspaceID, nil
}

func (s *Service) ensureWorkspaceClean(workspace *domain.Workspace, workspaceID, action string) error {
	if s.gitEngine == nil {
		return nil
	}

	for _, repo := range workspace.Repos {
		worktreePath := filepath.Join(s.config.GetWorkspacesRoot(), workspaceID, repo.Name)

		isDirty, unpushed, _, _, err := s.gitEngine.Status(context.Background(), worktreePath)
		if err != nil {
			// Log the error but continue checking other repos.
			// Status failures are non-fatal here as we're checking for uncommitted changes.
			if s.logger != nil {
				s.logger.Debug("Failed to check repo status",
					"repo", repo.Name,
					"path", worktreePath,
					"error", err)
			}

			continue
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
	}
}

// ExportWorkspace creates a portable export of a workspace definition.
func (s *Service) ExportWorkspace(ctx context.Context, workspaceID string) (*domain.WorkspaceExport, error) {
	return s.exportService.ExportWorkspace(ctx, workspaceID)
}

// ImportWorkspace creates a workspace from an exported definition.
func (s *Service) ImportWorkspace(ctx context.Context, export *domain.WorkspaceExport, idOverride, branchOverride string, force bool) (string, error) {
	return s.exportService.ImportWorkspace(ctx, export, idOverride, branchOverride, force)
}

// GetCanonicalRepoStatus returns detailed status for a single canonical repository.
func (s *Service) GetCanonicalRepoStatus(ctx context.Context, name string) (*domain.CanonicalRepoStatus, error) {
	if s.gitEngine == nil {
		return nil, cerrors.NewInternalError("git engine not initialized", nil)
	}

	usageMap, err := s.buildRepoUsageMap(ctx)
	if err != nil {
		return nil, err
	}

	return s.getCanonicalRepoStatus(ctx, name, usageMap)
}

// GetAllCanonicalRepoStatuses returns status for all canonical repositories.
func (s *Service) GetAllCanonicalRepoStatuses(ctx context.Context) ([]domain.CanonicalRepoStatus, error) {
	if s.gitEngine == nil {
		return nil, cerrors.NewInternalError("git engine not initialized", nil)
	}

	repoNames, err := s.gitEngine.List(ctx)
	if err != nil {
		return nil, cerrors.WrapGitError(err, "list canonical repos")
	}

	usageMap, err := s.buildRepoUsageMap(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]domain.CanonicalRepoStatus, 0, len(repoNames))
	for _, name := range repoNames {
		status, err := s.getCanonicalRepoStatus(ctx, name, usageMap)
		if err != nil {
			// Log error and skip this repo
			if s.logger != nil {
				s.logger.Warn("failed to get canonical repo status", "repo", name, "error", err)
			}

			continue
		}

		statuses = append(statuses, *status)
	}

	return statuses, nil
}

// getCanonicalRepoStatus is a helper that performs the status lookup with a precomputed usage map.
func (s *Service) getCanonicalRepoStatus(_ context.Context, name string, usageMap map[string][]string) (*domain.CanonicalRepoStatus, error) {
	path := filepath.Join(s.config.GetProjectsRoot(), name)

	// Check if repo exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, cerrors.NewRepoNotFound(name)
	}

	// Get disk usage
	size, err := s.gitEngine.GetRepoSize(name)
	if err != nil {
		return nil, cerrors.NewIOFailed(fmt.Sprintf("get repo size for %s", name), err)
	}

	// Get last fetch time
	lastFetch, err := s.gitEngine.LastFetchTime(name)
	if err != nil {
		return nil, cerrors.WrapGitError(err, fmt.Sprintf("get last fetch time for %s", name))
	}

	usedBy := usageMap[name]

	return &domain.CanonicalRepoStatus{
		Name:           name,
		Path:           path,
		DiskUsageBytes: size,
		LastFetchTime:  lastFetch,
		UsedByCount:    len(usedBy),
		UsedBy:         usedBy,
	}, nil
}

// buildRepoUsageMap builds a map of repository names to the IDs of workspaces that use them.
func (s *Service) buildRepoUsageMap(ctx context.Context) (map[string][]string, error) {
	workspaces, err := s.wsEngine.List(ctx)
	if err != nil {
		return nil, cerrors.NewIOFailed("list workspaces", err)
	}

	usageMap := make(map[string][]string)

	for _, ws := range workspaces {
		for _, repo := range ws.Repos {
			usageMap[repo.Name] = append(usageMap[repo.Name], ws.ID)
		}
	}

	return usageMap, nil
}
