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

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/hooks"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
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

	lockManager *LockManager

	// Extracted sub-services
	gitService    *WorkspaceGitService
	orphanService *WorkspaceOrphanService
	exportService *WorkspaceExportService
}

// ErrNoReposConfigured indicates no repos were specified and none matched configuration.
var ErrNoReposConfigured = errors.New("no repositories specified and no patterns matched")

// ServiceOption is a functional option for configuring the Service.
type ServiceOption func(*serviceOptions)

// serviceOptions holds optional dependencies that can be injected.
type serviceOptions struct {
	hookExecutor ports.HookExecutor
	diskUsage    ports.DiskUsage
	cache        ports.WorkspaceCache
	lockManager  *LockManager
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

// WithLockManager sets a custom LockManager implementation.
func WithLockManager(l *LockManager) ServiceOption {
	return func(o *serviceOptions) {
		o.lockManager = l
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

	lockManager := options.lockManager
	if lockManager == nil {
		lockTimeout := cfg.GetLockTimeout()
		if lockTimeout > 0 {
			lockManager = NewLockManager(cfg.GetWorkspacesRoot(), lockTimeout, cfg.GetLockStaleThreshold(), logger, cfg.ComputeWorkspaceDir)
		}
	}

	svc := &Service{
		config:       cfg,
		gitEngine:    gitEngine,
		wsEngine:     wsEngine,
		logger:       logger,
		hookExecutor: hookExecutor,
		resolver:     NewRepoResolver(cfg.GetRegistry()),
		diskUsage:    diskUsage,
		canonical:    NewCanonicalRepoService(gitEngine, wsEngine, cfg.GetProjectsRoot(), logger, diskUsage, cfg.GetRegistry()),
		cache:        cache,
		lockManager:  lockManager,
	}

	// Initialize sub-services with the main service as the workspace finder/creator
	svc.gitService = NewGitService(cfg, gitEngine, wsEngine, logger, cache, svc)
	svc.orphanService = NewOrphanService(cfg, gitEngine, wsEngine, logger, svc)
	svc.exportService = NewExportService(cfg, svc, svc)

	return svc
}

// FindWorkspace implements WorkspaceFinder interface for sub-services.
func (s *Service) FindWorkspace(ctx context.Context, workspaceID string) (*domain.Workspace, string, error) {
	return s.findWorkspace(ctx, workspaceID)
}

func (s *Service) withWorkspaceLock(ctx context.Context, workspaceID string, createDir bool, fn func() error) error {
	if s.lockManager == nil {
		return fn()
	}

	handle, err := s.lockManager.Acquire(ctx, workspaceID, createDir)
	if err != nil {
		return err
	}

	runErr := fn()

	releaseErr := handle.Release()
	if releaseErr != nil && s.logger != nil {
		s.logger.Warn("workspace lock release failed", "workspace_id", workspaceID, "error", releaseErr)
	}

	if runErr != nil {
		return runErr
	}

	// Only log release failures; successful operations shouldn't be marked as failed.
	return nil
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

	dirName := ws.DirName
	if dirName == "" {
		dirName, err = s.config.ComputeWorkspaceDir(workspaceID)
		if err != nil {
			return nil, "", err
		}
	}

	// Populate cache with the result
	s.cache.Set(workspaceID, ws, dirName)

	return ws, dirName, nil
}

// Canonical repository operations - delegated to CanonicalRepoService

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

// Git operations - delegated to WorkspaceGitService

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

// Orphan detection - delegated to WorkspaceOrphanService

// DetectOrphans finds orphaned worktrees across all workspaces.
// An orphan is a worktree reference in workspace metadata that:
// - References a canonical repo that no longer exists
// - Has a worktree directory that doesn't exist
// - Has an invalid git directory
func (s *Service) DetectOrphans(ctx context.Context) ([]domain.OrphanedWorktree, error) {
	return s.orphanService.DetectOrphans(ctx)
}

// GetWorkspacesUsingRepo returns the IDs of workspaces that use the given canonical repo.
func (s *Service) GetWorkspacesUsingRepo(ctx context.Context, repoName string) ([]string, error) {
	return s.canonical.GetWorkspacesUsingRepo(ctx, repoName)
}

// DetectOrphansForWorkspace returns orphans for a specific workspace.
// This is more efficient than DetectOrphans when only checking a single workspace.
func (s *Service) DetectOrphansForWorkspace(ctx context.Context, workspaceID string) ([]domain.OrphanedWorktree, error) {
	return s.orphanService.DetectOrphansForWorkspace(ctx, workspaceID)
}

// PruneAllWorktrees cleans up stale worktree references from all canonical repos.
// This removes worktree entries that point to non-existent directories.
func (s *Service) PruneAllWorktrees(ctx context.Context) error {
	return s.orphanService.PruneAllWorktrees(ctx)
}

// Export/Import - delegated to WorkspaceExportService

// ExportWorkspace creates a portable export of a workspace definition.
func (s *Service) ExportWorkspace(ctx context.Context, workspaceID string) (*domain.WorkspaceExport, error) {
	return s.exportService.ExportWorkspace(ctx, workspaceID)
}

// ImportWorkspace creates a workspace from an exported definition.
func (s *Service) ImportWorkspace(ctx context.Context, export *domain.WorkspaceExport, idOverride, branchOverride string, force bool) (string, error) {
	return s.exportService.ImportWorkspace(ctx, export, idOverride, branchOverride, force)
}

// Configuration accessors

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

// Canonical repository status methods

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

	executor := NewParallelExecutor(s.config.GetParallelWorkers())

	results, err := ParallelMap(ctx, executor, len(repoNames), func(runCtx context.Context, index int) (*domain.CanonicalRepoStatus, error) {
		if runCtx.Err() != nil {
			return nil, runCtx.Err()
		}

		return s.getCanonicalRepoStatus(runCtx, repoNames[index], usageMap)
	}, ParallelOptions{ContinueOnError: true})
	if err != nil {
		return nil, err
	}

	statuses := make([]domain.CanonicalRepoStatus, 0, len(repoNames))

	for i, result := range results {
		if result.Err != nil {
			if s.logger != nil {
				s.logger.Warn("failed to get canonical repo status", "repo", repoNames[i], "error", result.Err)
			}

			continue
		}

		if result.Value != nil {
			statuses = append(statuses, *result.Value)
		}
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
