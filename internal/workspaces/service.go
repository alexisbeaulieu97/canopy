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
//   - RestoreWorkspace: Restores an archived workspace
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
package workspaces

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	return s.findWorkspace(workspaceID)
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

	dirName := id

	// Default branch name is the workspace ID
	if branchName == "" {
		branchName = id
		// Validate the derived branch name (workspace IDs may contain chars invalid for git refs)
		if err := validation.ValidateBranchName(branchName); err != nil {
			return "", cerrors.NewInvalidArgument("workspace-id", "cannot be used as default branch name: "+err.Error())
		}
	}

	if err := s.wsEngine.Create(dirName, id, branchName, repos); err != nil {
		return "", err
	}

	// Invalidate cache for this workspace ID
	s.cache.Invalidate(id)

	// Manual cleanup helper
	cleanup := func() {
		path := fmt.Sprintf("%s/%s", s.config.GetWorkspacesRoot(), dirName)
		_ = os.RemoveAll(path)
	}

	// Clone repositories (if any)
	if err := s.cloneWorkspaceRepos(ctx, repos, dirName, branchName); err != nil {
		cleanup()
		return "", err
	}

	// Run post_create hooks
	//nolint:contextcheck // Hooks manage their own timeout context per-hook
	if err := s.runPostCreateHooks(id, dirName, branchName, repos, opts); err != nil {
		// Hook failures don't rollback the workspace (per design.md)
		// But we return the error if not continuing on hook errors
		return dirName, err
	}

	return dirName, nil
}

// cloneWorkspaceRepos clones all repositories for a workspace, checking for context cancellation.
func (s *Service) cloneWorkspaceRepos(ctx context.Context, repos []domain.Repo, dirName, branchName string) error {
	for _, repo := range repos {
		// Check for context cancellation before each operation
		if ctx.Err() != nil {
			return cerrors.NewContextError(ctx, "create workspace", dirName)
		}

		// Ensure canonical exists
		_, err := s.gitEngine.EnsureCanonical(ctx, repo.URL, repo.Name)
		if err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("ensure canonical for %s", repo.Name))
		}

		// Create worktree
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.CreateWorktree(repo.Name, worktreePath, branchName); err != nil {
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
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return "", cerrors.NewIOFailed("list workspaces", err)
	}

	for dir, w := range workspaces {
		if w.ID == workspaceID {
			return filepath.Join(s.config.GetWorkspacesRoot(), dir), nil
		}
	}

	return "", cerrors.NewWorkspaceNotFound(workspaceID)
}

// validateRenameInputs validates the inputs for renaming a workspace.
func (s *Service) validateRenameInputs(newID string) error {
	return validation.ValidateWorkspaceID(newID)
}

// ensureNewIDAvailable checks that the new workspace ID doesn't already exist.
func (s *Service) ensureNewIDAvailable(newID string) error {
	_, _, err := s.wsEngine.LoadByID(newID)
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
func (s *Service) updateBranchMetadata(dirName, newID string) error {
	ws, err := s.wsEngine.Load(dirName)
	if err != nil {
		return cerrors.NewWorkspaceMetadataError(newID, "load", err)
	}

	ws.BranchName = newID
	if err := s.wsEngine.Save(dirName, *ws); err != nil {
		return cerrors.NewWorkspaceMetadataError(newID, "save", err)
	}

	return nil
}

// renameWorkspaceDir renames the workspace directory and handles rollback on failure.
func (s *Service) renameWorkspaceDir(ctx context.Context, workspace domain.Workspace, oldDirName, newDirName, oldID, newID string, shouldRenameBranch bool) error {
	if err := s.wsEngine.Rename(oldDirName, newDirName, newID); err != nil {
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
func (s *Service) updateBranchMetadataWithRollback(ctx context.Context, workspace domain.Workspace, oldDirName, newDirName, oldID, newID string) error {
	if err := s.updateBranchMetadata(newDirName, newID); err != nil {
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
		if dirRollbackErr := s.wsEngine.Rename(newDirName, oldDirName, oldID); dirRollbackErr != nil {
			if s.logger != nil {
				s.logger.Error("failed to rollback workspace rename after metadata update error",
					"error", dirRollbackErr,
					"from", newDirName,
					"to", oldDirName,
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

// RenameWorkspace renames a workspace to a new ID.
// If renameBranch is true and the branch name matches the old ID, it will also rename branches.
func (s *Service) RenameWorkspace(ctx context.Context, oldID, newID string, renameBranch bool) error {
	workspace, oldDirName, err := s.findWorkspace(oldID)
	if err != nil {
		return err
	}

	if err := s.validateRenameInputs(newID); err != nil {
		return err
	}

	if err := s.ensureNewIDAvailable(newID); err != nil {
		return err
	}

	shouldRenameBranch := renameBranch && workspace.BranchName == oldID
	newDirName := newID

	if shouldRenameBranch {
		if err := s.renameBranchesInRepos(ctx, *workspace, oldDirName, oldID, newID); err != nil {
			return err
		}
	}

	if err := s.renameWorkspaceDir(ctx, *workspace, oldDirName, newDirName, oldID, newID, shouldRenameBranch); err != nil {
		return err
	}

	if shouldRenameBranch {
		if err := s.updateBranchMetadataWithRollback(ctx, *workspace, oldDirName, newDirName, oldID, newID); err != nil {
			return err
		}
	}

	s.invalidateWorkspaceCache(oldID, newID)

	return nil
}

// AddRepoToWorkspace adds a repository to an existing workspace
func (s *Service) AddRepoToWorkspace(ctx context.Context, workspaceID, repoName string) error {
	if err := validateAddRepoInputs(workspaceID, repoName); err != nil {
		return err
	}

	workspace, dirName, err := s.findWorkspace(workspaceID)
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

	if err := s.ensureWorkspaceWorktree(ctx, repo, dirName, branchName); err != nil {
		return err
	}

	if err := s.saveWorkspaceRepo(dirName, workspaceID, workspace, repo); err != nil {
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

	worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
	if err := s.gitEngine.CreateWorktree(repo.Name, worktreePath, branchName); err != nil {
		return cerrors.WrapGitError(err, fmt.Sprintf("create worktree for %s", repo.Name))
	}

	return nil
}

func (s *Service) saveWorkspaceRepo(dirName, workspaceID string, workspace *domain.Workspace, repo domain.Repo) error {
	workspace.Repos = append(workspace.Repos, repo)
	if err := s.wsEngine.Save(dirName, *workspace); err != nil {
		return cerrors.NewWorkspaceMetadataError(workspaceID, "update", err)
	}

	return nil
}

// RemoveRepoFromWorkspace removes a repository from an existing workspace
func (s *Service) RemoveRepoFromWorkspace(_ context.Context, workspaceID, repoName string) error {
	workspace, dirName, err := s.findWorkspace(workspaceID)
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
	worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repoName)
	if err := os.RemoveAll(worktreePath); err != nil {
		return cerrors.NewIOFailed(fmt.Sprintf("remove worktree %s", worktreePath), err)
	}

	// 4. Update metadata
	workspace.Repos = append(workspace.Repos[:repoIndex], workspace.Repos[repoIndex+1:]...)
	if err := s.wsEngine.Save(dirName, *workspace); err != nil {
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

// CloseWorkspace removes a workspace with safety checks
func (s *Service) CloseWorkspace(_ context.Context, workspaceID string, force bool) error {
	//nolint:contextcheck // Wrapper delegates to WithOptions which handles hooks with own timeout
	return s.CloseWorkspaceWithOptions(workspaceID, force, CloseOptions{})
}

// CloseWorkspaceWithOptions removes a workspace with configurable options.
func (s *Service) CloseWorkspaceWithOptions(workspaceID string, force bool, opts CloseOptions) error {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return err
	}

	if !force {
		if err := s.ensureWorkspaceClean(targetWorkspace, dirName, "close"); err != nil {
			return err
		}
	}

	// Run pre_close hooks before deletion
	if !opts.SkipHooks {
		hooksConfig := s.config.GetHooks()
		if len(hooksConfig.PreClose) > 0 {
			hookCtx := domain.HookContext{
				WorkspaceID:   workspaceID,
				WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
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

	// Delete workspace
	if err := s.wsEngine.Delete(dirName); err != nil {
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
func (s *Service) CloseWorkspaceKeepMetadataWithOptions(workspaceID string, force bool, opts CloseOptions) (*domain.ClosedWorkspace, error) {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	if !force {
		if err := s.ensureWorkspaceClean(targetWorkspace, dirName, "close"); err != nil {
			return nil, err
		}
	}

	// Run pre_close hooks before archiving
	if !opts.SkipHooks {
		hooksConfig := s.config.GetHooks()
		if len(hooksConfig.PreClose) > 0 {
			hookCtx := domain.HookContext{
				WorkspaceID:   workspaceID,
				WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
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

	archived, err := s.wsEngine.Close(dirName, *targetWorkspace, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if err := s.wsEngine.Delete(dirName); err != nil {
		_ = s.wsEngine.DeleteClosed(archived.Path)
		return nil, cerrors.NewIOFailed("remove workspace directory", err)
	}

	// Invalidate cache after workspace deletion
	s.cache.Invalidate(workspaceID)

	return archived, nil
}

// RunHooks executes lifecycle hooks for an existing workspace without performing other actions.
func (s *Service) RunHooks(workspaceID string, phase HookPhase, continueOnError bool) error {
	workspace, dirName, err := s.findWorkspace(workspaceID)
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
		WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
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
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	wsPath := filepath.Join(s.config.GetWorkspacesRoot(), dirName)

	repoNames := []string{}
	for _, r := range targetWorkspace.Repos {
		repoNames = append(repoNames, r.Name)
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
		DiskUsageBytes: usage,
		KeepMetadata:   keepMetadata,
	}, nil
}

// ListWorkspaces returns all active workspaces
func (s *Service) ListWorkspaces() ([]domain.Workspace, error) {
	workspaceMap, err := s.wsEngine.List()
	if err != nil {
		return nil, err
	}

	var workspaces []domain.Workspace

	for dir, w := range workspaceMap {
		wsPath := filepath.Join(s.config.GetWorkspacesRoot(), dir)

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

// CalculateDiskUsage sums file sizes under the provided root and returns latest mtime.
// Note: .git directories are skipped so LastModified reflects working tree activity.
//
// Deprecated: Use DiskUsageCalculator.Calculate directly. This method delegates to DiskUsageCalculator.
func (s *Service) CalculateDiskUsage(root string) (int64, time.Time, error) {
	return s.diskUsage.Calculate(root)
}

// ListClosedWorkspaces returns closed workspace metadata.
func (s *Service) ListClosedWorkspaces() ([]domain.ClosedWorkspace, error) {
	return s.wsEngine.ListClosed()
}

// GetStatus returns the aggregate status of a workspace
func (s *Service) GetStatus(workspaceID string) (*domain.WorkspaceStatus, error) {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	// 2. Check status for each repo
	var repoStatuses []domain.RepoStatus

	for _, repo := range targetWorkspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)

		isDirty, unpushed, behind, branch, err := s.gitEngine.Status(worktreePath)
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

// ListCanonicalRepos returns a list of all cached repositories
func (s *Service) ListCanonicalRepos() ([]string, error) {
	return s.canonical.List()
}

// AddCanonicalRepo adds a new repository to the cache and returns the canonical name.
func (s *Service) AddCanonicalRepo(ctx context.Context, url string) (string, error) {
	return s.canonical.Add(ctx, url)
}

// RemoveCanonicalRepo removes a repository from the cache
func (s *Service) RemoveCanonicalRepo(_ context.Context, name string, force bool) error {
	return s.canonical.Remove(name, force)
}

// PreviewRemoveCanonicalRepo returns a preview of what would happen when removing a repo.
func (s *Service) PreviewRemoveCanonicalRepo(name string) (*domain.RepoRemovePreview, error) {
	return s.canonical.PreviewRemove(name)
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
	archive, err := s.wsEngine.LatestClosed(workspaceID)
	if err != nil {
		return err
	}

	if _, _, err := s.findWorkspace(workspaceID); err == nil {
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

	if err := s.wsEngine.DeleteClosed(archive.Path); err != nil {
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

// DetectOrphans finds orphaned worktrees across all workspaces.
// An orphan is a worktree reference in workspace metadata that:
// - References a canonical repo that no longer exists
// - Has a worktree directory that doesn't exist
// - Has an invalid git directory
func (s *Service) DetectOrphans() ([]domain.OrphanedWorktree, error) {
	return s.orphanService.DetectOrphans()
}

// GetWorkspacesUsingRepo returns the IDs of workspaces that use the given canonical repo.
func (s *Service) GetWorkspacesUsingRepo(repoName string) ([]string, error) {
	return s.canonical.GetWorkspacesUsingRepo(repoName)
}

// DetectOrphansForWorkspace returns orphans for a specific workspace.
// This is more efficient than DetectOrphans when only checking a single workspace.
func (s *Service) DetectOrphansForWorkspace(workspaceID string) ([]domain.OrphanedWorktree, error) {
	return s.orphanService.DetectOrphansForWorkspace(workspaceID)
}

func (s *Service) findWorkspace(workspaceID string) (*domain.Workspace, string, error) {
	// Check cache first
	if ws, dirName, ok := s.cache.Get(workspaceID); ok {
		return ws, dirName, nil
	}

	// Cache miss: use direct lookup via LoadByID
	ws, dirName, err := s.wsEngine.LoadByID(workspaceID)
	if err != nil {
		return nil, "", err
	}

	// Populate cache with the result
	s.cache.Set(workspaceID, ws, dirName)

	return ws, dirName, nil
}

func (s *Service) ensureWorkspaceClean(workspace *domain.Workspace, dirName, action string) error {
	if s.gitEngine == nil {
		return nil
	}

	for _, repo := range workspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)

		isDirty, _, _, _, err := s.gitEngine.Status(worktreePath)
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
	}

	return nil
}

// ExportWorkspace creates a portable export of a workspace definition.
func (s *Service) ExportWorkspace(ctx context.Context, workspaceID string) (*domain.WorkspaceExport, error) {
	return s.exportService.ExportWorkspace(ctx, workspaceID)
}

// ImportWorkspace creates a workspace from an exported definition.
func (s *Service) ImportWorkspace(ctx context.Context, export *domain.WorkspaceExport, idOverride, branchOverride string, force bool) (string, error) {
	return s.exportService.ImportWorkspace(ctx, export, idOverride, branchOverride, force)
}
