// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alexisbeaulieu97/yard/internal/config"
	"github.com/alexisbeaulieu97/yard/internal/domain"
	"github.com/alexisbeaulieu97/yard/internal/gitx"
	"github.com/alexisbeaulieu97/yard/internal/logging"
	"github.com/alexisbeaulieu97/yard/internal/workspace"
)

// Service manages workspace operations
type Service struct {
	config    *config.Config
	gitEngine *gitx.GitEngine
	wsEngine  *workspace.Engine
	logger    *logging.Logger
	registry  *config.RepoRegistry
}

// ErrNoReposConfigured indicates no repos were specified and none matched configuration.
var ErrNoReposConfigured = errors.New("no repositories specified and no patterns matched")

// NewService creates a new workspace service
func NewService(cfg *config.Config, gitEngine *gitx.GitEngine, wsEngine *workspace.Engine, logger *logging.Logger) *Service {
	return &Service{
		config:    cfg,
		gitEngine: gitEngine,
		wsEngine:  wsEngine,
		logger:    logger,
		registry:  cfg.Registry,
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
func (s *Service) CreateWorkspace(id, slug, branchName string, repos []domain.Repo) (string, error) {
	// 1. Determine directory name
	// Simplified naming: Use ID as-is.
	// If slug is provided (legacy/optional), append it.
	dirName := id

	if slug != "" {
		// Keep the template logic if slug is provided, for backward compatibility or advanced usage
		var err error

		dirName, err = s.renderWorkspaceDirName(id, slug)
		if err != nil {
			return "", fmt.Errorf("failed to render workspace directory name: %w", err)
		}
	}

	// 2. Determine branch name
	if branchName == "" {
		branchName = id // Default branch name is the workspace ID
	}

	if err := s.wsEngine.Create(dirName, id, slug, branchName, repos); err != nil {
		return "", err
	}

	// Manual cleanup helper
	cleanup := func() {
		path := fmt.Sprintf("%s/%s", s.config.WorkspacesRoot, dirName)
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
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, dirName, repo.Name)
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
			return filepath.Join(s.config.WorkspacesRoot, dir), nil
		}
	}

	return "", fmt.Errorf("workspace %s not found", workspaceID)
}

func (s *Service) renderWorkspaceDirName(id, slug string) (string, error) {
	// Default to ID if no naming pattern
	pattern := s.config.WorkspaceNaming
	if pattern == "" {
		return id, nil
	}

	tmpl, err := template.New("workspace_dir").Parse(pattern)
	if err != nil {
		return "", err
	}

	data := struct {
		ID   string
		Slug string
	}{
		ID:   id,
		Slug: slug,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	// Clean up result: remove trailing separators if slug was empty
	result := buf.String()
	result = strings.TrimSuffix(result, "__")
	result = strings.TrimSuffix(result, "_")
	result = strings.TrimSuffix(result, "-")

	return result, nil
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
			return fmt.Errorf("repository %s already exists in workspace %s", repoName, workspaceID)
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
		branchName = workspace.ID // Fallback for legacy workspaces
	}

	worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, dirName, repo.Name)
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
		return fmt.Errorf("repository %s not found in workspace %s", repoName, workspaceID)
	}

	// 3. Remove worktree directory
	worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, dirName, repoName)
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
		for _, repo := range targetWorkspace.Repos {
			worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, dirName, repo.Name)

			isDirty, _, _, err := s.gitEngine.Status(worktreePath)
			if err != nil {
				continue
			}

			if isDirty {
				return fmt.Errorf("repo %s has uncommitted changes. Use --force to close", repo.Name)
			}
		}
	}

	// 2. Delete workspace
	return s.wsEngine.Delete(dirName)
}

// ListWorkspaces returns all active workspaces
func (s *Service) ListWorkspaces() ([]domain.Workspace, error) {
	workspaceMap, err := s.wsEngine.List()
	if err != nil {
		return nil, err
	}

	var workspaces []domain.Workspace
	for _, w := range workspaceMap {
		workspaces = append(workspaces, w)
	}

	return workspaces, nil
}

// GetStatus returns the aggregate status of a workspace
func (s *Service) GetStatus(workspaceID string) (*domain.WorkspaceStatus, error) {
	// 1. Load workspace metadata
	// We assume workspaceID maps to dirName for now, or we need to find dirName if they differ.
	// In current implementation, dirName is derived from ID (and slug), so we might not know it directly if we only have ID.
	// However, List() returns map[dirName]Workspace.
	// If we want to support looking up by ID without knowing dirName, we still need to List() or have an index.
	// But wait, CreateWorkspace returns dirName.
	// If we change GetStatus to take dirName, it changes the API.
	// Let's stick to List() for lookup if we don't know dirName, OR assume dirName == ID for simple cases.
	// But we support custom naming.
	// So we MUST List() to find the workspace by ID unless we enforce dirName == ID.
	// Let's keep the List() logic for finding the workspace, but use the found workspace object directly.
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	// 2. Check status for each repo
	var repoStatuses []domain.RepoStatus

	for _, repo := range targetWorkspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, dirName, repo.Name)

		isDirty, unpushed, branch, err := s.gitEngine.Status(worktreePath)
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
	// 1. Check if repo is used by any ticket
	tickets, err := s.wsEngine.List()
	if err != nil {
		return fmt.Errorf("failed to list tickets: %w", err)
	}

	var usedBy []string

	for _, t := range tickets {
		for _, r := range t.Repos {
			if r.Name == name {
				usedBy = append(usedBy, t.ID)
				break
			}
		}
	}

	if len(usedBy) > 0 && !force {
		return fmt.Errorf("repository %s is used by tickets: %s. Use --force to remove", name, strings.Join(usedBy, ", "))
	}

	// 2. Remove repo
	path := fmt.Sprintf("%s/%s", s.config.ProjectsRoot, name)
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

// SyncWorkspace pulls latest changes for all repos in a workspace
func (s *Service) SyncWorkspace(workspaceID string) error {
	targetWorkspace, dirName, err := s.findWorkspace(workspaceID)
	if err != nil {
		return err
	}

	// 2. Iterate through repos and pull
	for _, repo := range targetWorkspace.Repos {
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, dirName, repo.Name)
		s.logger.Info("Syncing repo", "repo", repo.Name)
		s.logger.Debug("Pulling changes", "path", worktreePath)

		if err := s.gitEngine.Pull(worktreePath); err != nil {
			// Log error but continue? Or fail?
			// For now, let's return error to alert user.
			return fmt.Errorf("failed to sync repo %s: %w", repo.Name, err)
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
		worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, dirName, repo.Name)
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

	return nil, "", fmt.Errorf("workspace %s not found", workspaceID)
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
		return domain.Repo{}, false, fmt.Errorf("unknown repository '%s'. Register it first: yard repo register %s <repository-url>", val, val)
	}

	// Legacy fallback for pattern-based names without explicit URLs.
	return domain.Repo{
		Name: val,
		URL:  fmt.Sprintf("https://github.com/example/%s", val),
	}, true, nil
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
