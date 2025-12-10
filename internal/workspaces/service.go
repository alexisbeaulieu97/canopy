// Package workspaces contains workspace-level business logic.
package workspaces

import (
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
)

// Service manages workspace operations
type Service struct {
	config       ports.ConfigProvider
	gitEngine    ports.GitOperations
	wsEngine     ports.WorkspaceStorage
	logger       *logging.Logger
	hookExecutor *hooks.Executor

	// Sub-services for specific responsibilities
	resolver  *RepoResolver
	diskUsage *DiskUsageCalculator
	canonical *CanonicalRepoService
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

// NewService creates a new workspace service
func NewService(cfg ports.ConfigProvider, gitEngine ports.GitOperations, wsEngine ports.WorkspaceStorage, logger *logging.Logger) *Service {
	diskUsage := DefaultDiskUsageCalculator()

	return &Service{
		config:       cfg,
		gitEngine:    gitEngine,
		wsEngine:     wsEngine,
		logger:       logger,
		hookExecutor: hooks.NewExecutor(logger),
		resolver:     NewRepoResolver(cfg.GetRegistry()),
		diskUsage:    diskUsage,
		canonical:    NewCanonicalRepoService(gitEngine, wsEngine, cfg.GetProjectsRoot(), logger, diskUsage),
	}
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
func (s *Service) CreateWorkspace(id, branchName string, repos []domain.Repo) (string, error) {
	return s.CreateWorkspaceWithOptions(id, branchName, repos, CreateOptions{})
}

// CreateWorkspaceWithOptions creates a new workspace with configurable options.
func (s *Service) CreateWorkspaceWithOptions(id, branchName string, repos []domain.Repo, opts CreateOptions) (string, error) {
	dirName := id

	// Default branch name is the workspace ID
	if branchName == "" {
		branchName = id
	}

	if err := s.wsEngine.Create(dirName, id, branchName, repos); err != nil {
		return "", err
	}

	// Manual cleanup helper
	cleanup := func() {
		path := fmt.Sprintf("%s/%s", s.config.GetWorkspacesRoot(), dirName)
		_ = os.RemoveAll(path)
	}

	// 3. Clone repositories (if any)
	for _, repo := range repos {
		// Ensure canonical exists
		_, err := s.gitEngine.EnsureCanonical(repo.URL, repo.Name)
		if err != nil {
			cleanup()
			return "", cerrors.WrapGitError(err, fmt.Sprintf("ensure canonical for %s", repo.Name))
		}

		// Create worktree
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.CreateWorktree(repo.Name, worktreePath, branchName); err != nil {
			cleanup()
			return "", cerrors.WrapGitError(err, fmt.Sprintf("create worktree for %s", repo.Name))
		}
	}

	// Run post_create hooks
	if !opts.SkipHooks {
		hooksConfig := s.config.GetHooks()
		if len(hooksConfig.PostCreate) > 0 {
			hookCtx := hooks.HookContext{
				WorkspaceID:   id,
				WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
				BranchName:    branchName,
				Repos:         repos,
			}

			if err := s.hookExecutor.ExecuteHooks(hooksConfig.PostCreate, hookCtx, opts.ContinueOnHookErr); err != nil {
				s.logger.Error("post_create hooks failed", "error", err)
				// Hook failures don't rollback the workspace (per design.md)
				// But we log the error and can optionally fail the operation
				if !opts.ContinueOnHookErr {
					return dirName, err
				}
			}
		}
	}

	return dirName, nil
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

// AddRepoToWorkspace adds a repository to an existing workspace
func (s *Service) AddRepoToWorkspace(workspaceID, repoName string) error {
	workspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return err
	}

	// 2. Check if repo already exists in workspace
	for _, r := range workspace.Repos {
		if r.Name == repoName {
			return cerrors.NewRepoAlreadyExists(repoName, workspaceID)
		}
	}

	// 3. Resolve repo URL
	repos, err := s.ResolveRepos(workspaceID, []string{repoName})
	if err != nil {
		// Preserve original error type if it's already typed
		var canopyErr *cerrors.CanopyError
		if errors.As(err, &canopyErr) {
			return canopyErr.WithContext("operation", fmt.Sprintf("resolve repo %s", repoName))
		}

		return cerrors.Wrap(cerrors.ErrUnknownRepository, fmt.Sprintf("failed to resolve repo %s", repoName), err)
	}

	repo := repos[0]

	// 4. Clone repo
	// Ensure canonical exists
	_, err = s.gitEngine.EnsureCanonical(repo.URL, repo.Name)
	if err != nil {
		return cerrors.WrapGitError(err, fmt.Sprintf("ensure canonical for %s", repo.Name))
	}

	// Create worktree
	branchName := workspace.BranchName
	if branchName == "" {
		return cerrors.NewMissingBranchConfig(workspaceID)
	}

	worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
	if err := s.gitEngine.CreateWorktree(repo.Name, worktreePath, branchName); err != nil {
		return cerrors.WrapGitError(err, fmt.Sprintf("create worktree for %s", repo.Name))
	}

	// 5. Update metadata
	workspace.Repos = append(workspace.Repos, repo)
	if err := s.wsEngine.Save(dirName, *workspace); err != nil {
		return cerrors.NewWorkspaceMetadataError(workspaceID, "update", err)
	}

	return nil
}

// RemoveRepoFromWorkspace removes a repository from an existing workspace
func (s *Service) RemoveRepoFromWorkspace(workspaceID, repoName string) error {
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

	return nil
}

// CloseOptions configures workspace close behavior.
type CloseOptions struct {
	SkipHooks         bool // Skip pre_close hooks
	ContinueOnHookErr bool // Continue if hooks fail
}

// CloseWorkspace removes a workspace with safety checks
func (s *Service) CloseWorkspace(workspaceID string, force bool) error {
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
			hookCtx := hooks.HookContext{
				WorkspaceID:   workspaceID,
				WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
				BranchName:    targetWorkspace.BranchName,
				Repos:         targetWorkspace.Repos,
			}

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
	return s.wsEngine.Delete(dirName)
}

// CloseWorkspaceKeepMetadata moves workspace metadata to the closed store and removes the active worktree.
func (s *Service) CloseWorkspaceKeepMetadata(workspaceID string, force bool) (*domain.ClosedWorkspace, error) {
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
			hookCtx := hooks.HookContext{
				WorkspaceID:   workspaceID,
				WorkspacePath: filepath.Join(s.config.GetWorkspacesRoot(), dirName),
				BranchName:    targetWorkspace.BranchName,
				Repos:         targetWorkspace.Repos,
			}

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

	hookCtx := hooks.HookContext{
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
func (s *Service) AddCanonicalRepo(url string) (string, error) {
	return s.canonical.Add(url)
}

// RemoveCanonicalRepo removes a repository from the cache
func (s *Service) RemoveCanonicalRepo(name string, force bool) error {
	return s.canonical.Remove(name, force)
}

// PreviewRemoveCanonicalRepo returns a preview of what would happen when removing a repo.
func (s *Service) PreviewRemoveCanonicalRepo(name string) (*domain.RepoRemovePreview, error) {
	return s.canonical.PreviewRemove(name)
}

// SyncCanonicalRepo fetches updates for a cached repository
func (s *Service) SyncCanonicalRepo(name string) error {
	return s.canonical.Sync(name)
}

// PushWorkspace pushes all repos for a workspace.
func (s *Service) PushWorkspace(workspaceID string) error {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return err
	}

	for _, repo := range targetWorkspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
		branchName := targetWorkspace.BranchName

		if branchName == "" {
			if s.logger != nil {
				s.logger.Debug("Branch missing in metadata, will let git infer", "workspace", workspaceID, "repo", repo.Name)
			}
		}

		if err := s.gitEngine.Push(worktreePath, branchName); err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("push repo %s", repo.Name))
		}
	}

	return nil
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
func (s *Service) RunGitInWorkspace(workspaceID string, args []string, opts GitRunOptions) ([]RepoGitResult, error) {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	if len(targetWorkspace.Repos) == 0 {
		return nil, nil
	}

	if opts.Parallel {
		return s.runGitParallel(targetWorkspace, dirName, args, opts.ContinueOnError)
	}

	return s.runGitSequential(targetWorkspace, dirName, args, opts.ContinueOnError)
}

func (s *Service) runGitSequential(workspace *domain.Workspace, dirName string, args []string, continueOnError bool) ([]RepoGitResult, error) {
	var results []RepoGitResult

	for _, repo := range workspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)

		cmdResult, err := s.gitEngine.RunCommand(worktreePath, args...)

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

const defaultMaxParallel = 10

func (s *Service) runGitParallel(workspace *domain.Workspace, dirName string, args []string, continueOnError bool) ([]RepoGitResult, error) {
	results := make([]RepoGitResult, len(workspace.Repos))

	var wg sync.WaitGroup

	// Bounded worker pool to avoid exhausting resources for large workspaces
	sem := make(chan struct{}, defaultMaxParallel)

	for i, repo := range workspace.Repos {
		wg.Add(1)

		go func(idx int, r domain.Repo) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}

			defer func() { <-sem }()

			worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, r.Name)

			result := RepoGitResult{
				RepoName: r.Name,
			}

			cmdResult, err := s.gitEngine.RunCommand(worktreePath, args...)
			if err != nil {
				result.Error = err
				results[idx] = result

				return
			}

			result.Stdout = cmdResult.Stdout
			result.Stderr = cmdResult.Stderr
			result.ExitCode = cmdResult.ExitCode
			results[idx] = result
		}(i, repo)
	}

	wg.Wait()

	// Check for errors if not continuing on error
	if !continueOnError {
		for _, r := range results {
			if r.Error != nil {
				return results, r.Error
			}

			if r.ExitCode != 0 {
				return results, cerrors.NewCommandFailed(fmt.Sprintf("git in repo %s", r.RepoName), fmt.Errorf("exit code %d", r.ExitCode))
			}
		}
	}

	return results, nil
}

// SwitchBranch switches the branch for all repos in a workspace
func (s *Service) SwitchBranch(workspaceID, branchName string, create bool) error {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return err
	}

	// 2. Iterate through repos and checkout
	for _, repo := range targetWorkspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
		s.logger.Info("Switching branch", "repo", repo.Name, "branch", branchName)

		if err := s.gitEngine.Checkout(worktreePath, branchName, create); err != nil {
			return cerrors.WrapGitError(err, fmt.Sprintf("checkout branch %s in repo %s", branchName, repo.Name))
		}
	}

	// 3. Update metadata
	targetWorkspace.BranchName = branchName
	if err := s.wsEngine.Save(dirName, *targetWorkspace); err != nil {
		return cerrors.NewWorkspaceMetadataError(workspaceID, "update", err)
	}

	return nil
}

// RestoreWorkspace recreates a workspace from the newest closed entry.
func (s *Service) RestoreWorkspace(workspaceID string, force bool) error {
	archive, err := s.wsEngine.LatestClosed(workspaceID)
	if err != nil {
		return err
	}

	if _, _, err := s.findWorkspace(workspaceID); err == nil {
		if !force {
			return cerrors.NewWorkspaceExists(workspaceID).WithContext("hint", "Use --force to replace or choose a different ID")
		}

		if err := s.CloseWorkspace(workspaceID, true); err != nil {
			return cerrors.NewIOFailed("remove existing workspace", err)
		}
	}

	ws := archive.Metadata
	ws.ClosedAt = nil

	if _, err := s.CreateWorkspace(ws.ID, ws.BranchName, ws.Repos); err != nil {
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
	workspaceMap, err := s.wsEngine.List()
	if err != nil {
		return nil, cerrors.NewIOFailed("list workspaces", err)
	}

	canonicalSet, err := s.buildCanonicalRepoSet()
	if err != nil {
		return nil, err
	}

	var orphans []domain.OrphanedWorktree

	for dir, ws := range workspaceMap {
		wsOrphans := s.checkWorkspaceForOrphans(ws, dir, canonicalSet)
		orphans = append(orphans, wsOrphans...)
	}

	return orphans, nil
}

// buildCanonicalRepoSet returns a set of canonical repo names.
func (s *Service) buildCanonicalRepoSet() (map[string]bool, error) {
	canonicalRepos, err := s.gitEngine.List()
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
func (s *Service) checkWorkspaceForOrphans(
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
func (s *Service) checkRepoForOrphan(
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
		s.logStatError("worktree directory", workspaceID, repoName, worktreePath, err)

		return &domain.OrphanedWorktree{
			WorkspaceID:  workspaceID,
			RepoName:     repoName,
			WorktreePath: worktreePath,
			Reason:       domain.OrphanReasonDirectoryMissing,
		}
	}

	// Check 3: Valid git directory
	gitDir := filepath.Join(worktreePath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		s.logStatError(".git directory", workspaceID, repoName, gitDir, err)

		return &domain.OrphanedWorktree{
			WorkspaceID:  workspaceID,
			RepoName:     repoName,
			WorktreePath: worktreePath,
			Reason:       domain.OrphanReasonInvalidGitDir,
		}
	}

	return nil
}

// logStatError logs stat errors if they are not IsNotExist errors.
func (s *Service) logStatError(itemType, workspaceID, repoName, path string, err error) {
	if !os.IsNotExist(err) && s.logger != nil {
		s.logger.Debug("Failed to stat "+itemType,
			"workspace", workspaceID, "repo", repoName, "path", path, "error", err)
	}
}

// GetWorkspacesUsingRepo returns the IDs of workspaces that use the given canonical repo.
func (s *Service) GetWorkspacesUsingRepo(repoName string) ([]string, error) {
	return s.canonical.GetWorkspacesUsingRepo(repoName)
}

// DetectOrphansForWorkspace returns orphans for a specific workspace.
// This is more efficient than DetectOrphans when only checking a single workspace.
func (s *Service) DetectOrphansForWorkspace(workspaceID string) ([]domain.OrphanedWorktree, error) {
	ws, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	canonicalSet, err := s.buildCanonicalRepoSet()
	if err != nil {
		return nil, err
	}

	return s.checkWorkspaceForOrphans(*ws, dirName, canonicalSet), nil
}

func (s *Service) findWorkspace(workspaceID string) (*domain.Workspace, string, error) {
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return nil, "", cerrors.NewIOFailed("list workspaces", err)
	}

	for dir, w := range workspaces {
		if w.ID == workspaceID {
			return &w, dir, nil
		}
	}

	return nil, "", cerrors.NewWorkspaceNotFound(workspaceID)
}

func (s *Service) ensureWorkspaceClean(workspace *domain.Workspace, dirName, action string) error {
	if s.gitEngine == nil {
		return nil
	}

	for _, repo := range workspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)

		isDirty, _, _, _, err := s.gitEngine.Status(worktreePath)
		if err != nil {
			continue
		}

		if isDirty {
			return cerrors.NewRepoNotClean(repo.Name, action)
		}
	}

	return nil
}

// ExportWorkspace creates a portable export of a workspace definition.
func (s *Service) ExportWorkspace(workspaceID string) (*domain.WorkspaceExport, error) {
	workspace, _, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	export := &domain.WorkspaceExport{
		Version:    "1",
		ID:         workspace.ID,
		Branch:     workspace.BranchName,
		ExportedAt: time.Now().UTC(),
		Repos:      make([]domain.RepoExport, 0, len(workspace.Repos)),
	}

	for _, repo := range workspace.Repos {
		repoExport := domain.RepoExport{
			Name: repo.Name,
			URL:  repo.URL,
		}

		// Try to find registry alias for this URL
		if s.resolver != nil && s.resolver.registry != nil {
			if entry, ok := s.resolver.registry.ResolveByURL(repo.URL); ok {
				repoExport.Alias = entry.Alias
			}
		}

		export.Repos = append(export.Repos, repoExport)
	}

	return export, nil
}

// ImportWorkspace creates a workspace from an exported definition.
func (s *Service) ImportWorkspace(export *domain.WorkspaceExport, idOverride, branchOverride string, force bool) (string, error) {
	if export == nil {
		return "", cerrors.NewInvalidArgument("export", "export definition is nil")
	}

	// Validate version
	if export.Version != "1" {
		return "", cerrors.NewInvalidArgument("version", fmt.Sprintf("unsupported export version: %s", export.Version))
	}

	// Resolve final workspace ID and branch name
	workspaceID, branchName := s.resolveImportOverrides(export, idOverride, branchOverride)

	// Handle existing workspace
	if err := s.prepareForImport(workspaceID, force); err != nil {
		return "", err
	}

	// Resolve repos from export
	repos, err := s.resolveExportedRepos(export.Repos, workspaceID)
	if err != nil {
		return "", err
	}

	// Create the workspace
	return s.CreateWorkspace(workspaceID, branchName, repos)
}

// resolveImportOverrides determines the final workspace ID and branch name for import.
func (s *Service) resolveImportOverrides(export *domain.WorkspaceExport, idOverride, branchOverride string) (string, string) {
	workspaceID := export.ID
	if idOverride != "" {
		workspaceID = idOverride
	}

	branchName := export.Branch
	if branchOverride != "" {
		branchName = branchOverride
	}

	// Default branch to workspace ID if not specified (consistent with workspace new)
	if branchName == "" {
		branchName = workspaceID
	}

	return workspaceID, branchName
}

// prepareForImport checks for existing workspace and removes it if force is set.
func (s *Service) prepareForImport(workspaceID string, force bool) error {
	_, _, findErr := s.findWorkspace(workspaceID)
	if findErr == nil {
		// Workspace exists
		if !force {
			return cerrors.NewWorkspaceExists(workspaceID).WithContext("hint", "Use --force to overwrite or --id to specify a different ID")
		}
		// Force mode: delete existing workspace
		if err := s.CloseWorkspace(workspaceID, true); err != nil {
			return cerrors.NewIOFailed("remove existing workspace", err)
		}

		return nil
	}

	if !errors.Is(findErr, cerrors.WorkspaceNotFound) {
		// Unexpected error (IO failure, etc.) - propagate it
		return findErr
	}

	// Workspace not found, proceed with import
	return nil
}

// resolveExportedRepos converts exported repo definitions to domain.Repo objects.
func (s *Service) resolveExportedRepos(exportedRepos []domain.RepoExport, workspaceID string) ([]domain.Repo, error) {
	repos := make([]domain.Repo, 0, len(exportedRepos))

	for _, exported := range exportedRepos {
		var repo domain.Repo

		var resolved bool

		// Try registry alias first if available.
		// When alias resolves, we use the registry's canonical name (entry.Alias) rather than
		// the exported name. This ensures consistency with the local registry and handles cases
		// where the exporting machine used a different alias for the same repo.
		if exported.Alias != "" && s.resolver != nil && s.resolver.registry != nil {
			if entry, ok := s.resolver.registry.Resolve(exported.Alias); ok {
				repo = domain.Repo{Name: entry.Alias, URL: entry.URL}
				resolved = true
			}
		}

		// Fall back to URL
		if !resolved && exported.URL != "" {
			repo = domain.Repo{Name: exported.Name, URL: exported.URL}
			resolved = true
		}

		if !resolved {
			return nil, cerrors.NewUnknownRepository(exported.Name, true).WithContext("workspace_id", workspaceID)
		}

		repos = append(repos, repo)
	}

	return repos, nil
}
