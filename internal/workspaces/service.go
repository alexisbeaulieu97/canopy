package workspaces

import (
	"bytes"
	"fmt"
	"os"
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
}

// NewService creates a new workspace service
func NewService(cfg *config.Config, gitEngine *gitx.GitEngine, wsEngine *workspace.Engine, logger *logging.Logger) *Service {
	return &Service{
		config:    cfg,
		gitEngine: gitEngine,
		wsEngine:  wsEngine,
		logger:    logger,
	}
}

// ResolveRepos determines which repos should be part of the workspace
func (s *Service) ResolveRepos(workspaceID string, requestedRepos []string) ([]domain.Repo, error) {
	var repoNames []string

	// 1. Use requested repos if provided
	if len(requestedRepos) > 0 {
		repoNames = requestedRepos
	} else {
		// 2. Fallback to config patterns
		repoNames = s.config.GetReposForWorkspace(workspaceID)
	}

	if len(repoNames) == 0 {
		return nil, fmt.Errorf("no repositories specified and no patterns matched for %s", workspaceID)
	}

	var repos []domain.Repo
	for _, val := range repoNames {
		if val == "" {
			continue
		}

		name := val
		url := "https://github.com/example/" + name

		// Check if it is a full URL
		if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") || strings.HasPrefix(val, "git@") {
			url = val
			// Extract name from URL
			parts := strings.Split(val, "/")
			if len(parts) > 0 {
				name = parts[len(parts)-1]
				name = strings.TrimSuffix(name, ".git")
			}
		} else if strings.Contains(val, "/") {
			// If input is like "owner/repo", use that for URL and repo name
			parts := strings.Split(val, "/")
			if len(parts) == 2 {
				name = parts[1]
				url = "https://github.com/" + val
			}
		}

		repos = append(repos, domain.Repo{Name: name, URL: url})
	}

	return repos, nil
}

// CreateWorkspace creates a new workspace directory and returns the directory name
func (s *Service) CreateWorkspace(id string, slug string, branchName string, repos []domain.Repo) (string, error) {
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

	// Setup cleanup on failure
	createdDir := false
	defer func() {
		if createdDir {
			// Manual cleanup handled below
		}
	}()

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
	// 1. Get workspace details to find directory and branch
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	var workspace *domain.Workspace
	var dirName string
	for dir, w := range workspaces {
		if w.ID == workspaceID {
			workspace = &w
			dirName = dir
			break
		}
	}
	if workspace == nil {
		return fmt.Errorf("workspace not found: %s", workspaceID)
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
	// 1. Get workspace details
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	var workspace *domain.Workspace
	var dirName string
	for dir, w := range workspaces {
		if w.ID == workspaceID {
			workspace = &w
			dirName = dir
			break
		}
	}
	if workspace == nil {
		return fmt.Errorf("workspace not found: %s", workspaceID)
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
	// 1. Check for uncommitted/unpushed changes
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return err
	}

	var targetWorkspace *domain.Workspace
	for _, w := range workspaces {
		if w.ID == workspaceID {
			targetWorkspace = &w
			break
		}
	}

	if targetWorkspace == nil {
		return fmt.Errorf("workspace %s not found", workspaceID)
	}

	if !force {
		for _, repo := range targetWorkspace.Repos {
			worktreePath := fmt.Sprintf("%s/%s/%s", s.config.WorkspacesRoot, workspaceID, repo.Name)
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
	return s.wsEngine.Delete(workspaceID)
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

	workspaces, err := s.wsEngine.List()
	if err != nil {
		return nil, err
	}

	var targetWorkspace *domain.Workspace
	var dirName string
	for dir, w := range workspaces {
		if w.ID == workspaceID {
			targetWorkspace = &w
			dirName = dir
			break
		}
	}

	if targetWorkspace == nil {
		return nil, fmt.Errorf("workspace %s not found", workspaceID)
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

// AddCanonicalRepo adds a new repository to the cache
func (s *Service) AddCanonicalRepo(url string) error {
	// Extract name from URL
	// Assuming URL ends with /name.git or /name
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid URL: %s", url)
	}
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".git")

	if name == "" {
		return fmt.Errorf("could not determine repo name from URL: %s", url)
	}

	return s.gitEngine.Clone(url, name)
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
	// 1. Get workspace details
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	var targetWorkspace *domain.Workspace
	var dirName string
	for dir, w := range workspaces {
		if w.ID == workspaceID {
			targetWorkspace = &w
			dirName = dir
			break
		}
	}

	if targetWorkspace == nil {
		return fmt.Errorf("workspace %s not found", workspaceID)
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
	// 1. Get workspace details
	workspaces, err := s.wsEngine.List()
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	var targetWorkspace *domain.Workspace
	var dirName string
	for dir, w := range workspaces {
		if w.ID == workspaceID {
			targetWorkspace = &w
			dirName = dir
			break
		}
	}

	if targetWorkspace == nil {
		return fmt.Errorf("workspace %s not found", workspaceID)
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
