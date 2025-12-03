// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// Service manages workspace operations
type Service struct {
	config     ports.ConfigProvider
	gitEngine  ports.GitOperations
	wsEngine   ports.WorkspaceStorage
	logger     *logging.Logger
	registry   *config.RepoRegistry
	usageCache map[string]usageEntry
	usageMu    sync.Mutex
}

type usageEntry struct {
	usage     int64
	lastMod   time.Time
	scannedAt time.Time
	err       error
}

// ErrNoReposConfigured indicates no repos were specified and none matched configuration.
var ErrNoReposConfigured = errors.New("no repositories specified and no patterns matched")

// NewService creates a new workspace service
func NewService(cfg ports.ConfigProvider, gitEngine ports.GitOperations, wsEngine ports.WorkspaceStorage, logger *logging.Logger) *Service {
	return &Service{
		config:     cfg,
		gitEngine:  gitEngine,
		wsEngine:   wsEngine,
		logger:     logger,
		registry:   cfg.GetRegistry(),
		usageCache: make(map[string]usageEntry),
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
		return nil, fmt.Errorf("%w for %s", ErrNoReposConfigured, workspaceID)
	}

	var repos []domain.Repo

	for _, raw := range repoNames {
		repo, ok, err := s.resolveRepoIdentifier(raw, userRequested)
		if err != nil {
			return nil, err
		}

		if ok {
			repos = append(repos, repo)
		}
	}

	return repos, nil
}

// CreateWorkspace creates a new workspace directory and returns the directory name
func (s *Service) CreateWorkspace(id, branchName string, repos []domain.Repo) (string, error) {
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
			return "", fmt.Errorf("failed to ensure canonical for %s: %w", repo.Name, err)
		}

		// Create worktree
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
		if err := s.gitEngine.CreateWorktree(repo.Name, worktreePath, branchName); err != nil {
			cleanup()
			return "", fmt.Errorf("failed to create worktree for %s: %w", repo.Name, err)
		}
	}

	return dirName, nil
}

// WorkspacePath returns the absolute path for a workspace ID.
func (s *Service) WorkspacePath(workspaceID string) (string, error) {
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return "", fmt.Errorf("failed to list workspaces: %w", err)
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
		return fmt.Errorf("failed to resolve repo %s: %w", repoName, err)
	}

	repo := repos[0]

	// 4. Clone repo
	// Ensure canonical exists
	_, err = s.gitEngine.EnsureCanonical(repo.URL, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to ensure canonical for %s: %w", repo.Name, err)
	}

	// Create worktree
	branchName := workspace.BranchName
	if branchName == "" {
		return fmt.Errorf("workspace %s has no branch set in metadata", workspaceID)
	}

	worktreePath := fmt.Sprintf("%s/%s/%s", s.config.GetWorkspacesRoot(), dirName, repo.Name)
	if err := s.gitEngine.CreateWorktree(repo.Name, worktreePath, branchName); err != nil {
		return fmt.Errorf("failed to create worktree for %s: %w", repo.Name, err)
	}

	// 5. Update metadata
	workspace.Repos = append(workspace.Repos, repo)
	if err := s.wsEngine.Save(dirName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace metadata: %w", err)
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
		return fmt.Errorf("failed to remove worktree %s: %w", worktreePath, err)
	}

	// 4. Update metadata
	workspace.Repos = append(workspace.Repos[:repoIndex], workspace.Repos[repoIndex+1:]...)
	if err := s.wsEngine.Save(dirName, *workspace); err != nil {
		return fmt.Errorf("failed to update workspace metadata: %w", err)
	}

	return nil
}

// CloseWorkspace removes a workspace with safety checks
func (s *Service) CloseWorkspace(workspaceID string, force bool) error {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return err
	}

	if !force {
		if err := s.ensureWorkspaceClean(targetWorkspace, dirName, "close"); err != nil {
			return err
		}
	}

	// 2. Delete workspace
	return s.wsEngine.Delete(dirName)
}

// CloseWorkspaceKeepMetadata moves workspace metadata to the closed store and removes the active worktree.
func (s *Service) CloseWorkspaceKeepMetadata(workspaceID string, force bool) (*domain.ClosedWorkspace, error) {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	if !force {
		if err := s.ensureWorkspaceClean(targetWorkspace, dirName, "close"); err != nil {
			return nil, err
		}
	}

	archived, err := s.wsEngine.Close(dirName, *targetWorkspace, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if err := s.wsEngine.Delete(dirName); err != nil {
		_ = s.wsEngine.DeleteClosed(archived.Path)
		return nil, fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	return archived, nil
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

		usage, latest, sizeErr := s.cachedWorkspaceUsage(wsPath)
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

// cachedWorkspaceUsage returns cached workspace usage/mtime with a short TTL to avoid repeated scans.
func (s *Service) cachedWorkspaceUsage(root string) (int64, time.Time, error) {
	const ttl = time.Minute

	s.usageMu.Lock()

	entry, ok := s.usageCache[root]
	if ok && time.Since(entry.scannedAt) < ttl {
		s.usageMu.Unlock()
		return entry.usage, entry.lastMod, entry.err
	}

	s.usageMu.Unlock()

	usage, latest, err := s.CalculateDiskUsage(root)

	s.usageMu.Lock()
	s.usageCache[root] = usageEntry{
		usage:     usage,
		lastMod:   latest,
		scannedAt: time.Now(),
		err:       err,
	}
	s.usageMu.Unlock()

	return usage, latest, err
}

// CalculateDiskUsage sums file sizes under the provided root and returns latest mtime.
// Note: .git directories are skipped so LastModified reflects working tree activity.
func (s *Service) CalculateDiskUsage(root string) (int64, time.Time, error) {
	var (
		total   int64
		latest  time.Time
		skipDir = map[string]struct{}{".git": {}}
	)

	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if _, ok := skipDir[d.Name()]; ok {
				return fs.SkipDir
			}

			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}

		total += info.Size()

		if mod := info.ModTime(); mod.After(latest) {
			latest = mod
		}

		return nil
	})
	if err != nil {
		return 0, time.Time{}, err
	}

	return total, latest, nil
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
	return s.gitEngine.List()
}

// AddCanonicalRepo adds a new repository to the cache and returns the canonical name.
func (s *Service) AddCanonicalRepo(url string) (string, error) {
	name := repoNameFromURL(url)
	if name == "" {
		return "", fmt.Errorf("could not determine repo name from URL: %s", url)
	}

	return name, s.gitEngine.Clone(url, name)
}

// RemoveCanonicalRepo removes a repository from the cache
func (s *Service) RemoveCanonicalRepo(name string, force bool) error {
	// 1. Check if repo is used by any workspace
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	var usedBy []string

	for _, ws := range workspaces {
		for _, r := range ws.Repos {
			if r.Name == name {
				usedBy = append(usedBy, ws.ID)
				break
			}
		}
	}

	if len(usedBy) > 0 && !force {
		return fmt.Errorf("repository %s is used by workspaces: %s. Use --force to remove", name, strings.Join(usedBy, ", "))
	}

	// 2. Remove repo
	path := fmt.Sprintf("%s/%s", s.config.GetProjectsRoot(), name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("repository %s does not exist", name)
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove repo %s: %w", name, err)
	}

	return nil
}

// SyncCanonicalRepo fetches updates for a cached repository
func (s *Service) SyncCanonicalRepo(name string) error {
	return s.gitEngine.Fetch(name)
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
			return fmt.Errorf("failed to push repo %s: %w", repo.Name, err)
		}
	}

	return nil
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
			return fmt.Errorf("failed to checkout branch %s in repo %s: %w", branchName, repo.Name, err)
		}
	}

	// 3. Update metadata
	targetWorkspace.BranchName = branchName
	if err := s.wsEngine.Save(dirName, *targetWorkspace); err != nil {
		return fmt.Errorf("failed to update workspace metadata: %w", err)
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
			return fmt.Errorf("workspace %s already exists. Use --force to replace or choose a different ID", workspaceID)
		}

		if err := s.CloseWorkspace(workspaceID, true); err != nil {
			return fmt.Errorf("failed to remove existing workspace: %w", err)
		}
	}

	ws := archive.Metadata
	ws.ClosedAt = nil

	if _, err := s.CreateWorkspace(ws.ID, ws.BranchName, ws.Repos); err != nil {
		return fmt.Errorf("failed to restore workspace %s: %w", workspaceID, err)
	}

	if err := s.wsEngine.DeleteClosed(archive.Path); err != nil {
		return fmt.Errorf("failed to remove closed entry: %w", err)
	}

	return nil
}

// StaleThresholdDays returns the configured stale threshold in days.
func (s *Service) StaleThresholdDays() int {
	return s.config.GetStaleThresholdDays()
}

func (s *Service) findWorkspace(workspaceID string) (*domain.Workspace, string, error) {
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list workspaces: %w", err)
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

func isLikelyURL(val string) bool {
	return strings.HasPrefix(val, "http://") ||
		strings.HasPrefix(val, "https://") ||
		strings.HasPrefix(val, "ssh://") ||
		strings.HasPrefix(val, "git://") ||
		strings.HasPrefix(val, "git@") ||
		strings.HasPrefix(val, "file://")
}

func (s *Service) resolveRepoIdentifier(raw string, userRequested bool) (domain.Repo, bool, error) {
	val := strings.TrimSpace(raw)
	if val == "" {
		return domain.Repo{}, false, nil
	}

	if isLikelyURL(val) {
		if s.registry != nil {
			if entry, ok := s.registry.ResolveByURL(val); ok {
				return domain.Repo{Name: entry.Alias, URL: entry.URL}, true, nil
			}
		}

		return domain.Repo{Name: repoNameFromURL(val), URL: val}, true, nil
	}

	if s.registry != nil {
		if entry, ok := s.registry.Resolve(val); ok {
			return domain.Repo{Name: entry.Alias, URL: entry.URL}, true, nil
		}
	}

	if strings.Count(val, "/") == 1 {
		parts := strings.Split(val, "/")
		url := "https://github.com/" + val

		return domain.Repo{Name: parts[1], URL: url}, true, nil
	}

	if userRequested {
		return domain.Repo{}, false, cerrors.NewUnknownRepository(val, true)
	}

	return domain.Repo{}, false, cerrors.NewUnknownRepository(val, false)
}

func repoNameFromURL(url string) string {
	// Strip scp-like prefix if present
	if strings.Contains(url, ":") && !strings.HasPrefix(url, "http") {
		parts := strings.Split(url, ":")
		url = parts[len(parts)-1]
	}

	parts := strings.Split(url, "/")

	var name string

	for i := len(parts) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(parts[i]); trimmed != "" {
			name = trimmed
			break
		}
	}

	if name == "" {
		return ""
	}

	return strings.TrimSuffix(name, ".git")
}
