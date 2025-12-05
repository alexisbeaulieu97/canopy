// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// CanonicalRepoService manages canonical (bare) repository operations.
type CanonicalRepoService struct {
	gitEngine    ports.GitOperations
	wsStorage    ports.WorkspaceStorage
	projectsRoot string
	logger       *logging.Logger
	diskUsage    *DiskUsageCalculator
}

// NewCanonicalRepoService creates a new CanonicalRepoService.
func NewCanonicalRepoService(
	gitEngine ports.GitOperations,
	wsStorage ports.WorkspaceStorage,
	projectsRoot string,
	logger *logging.Logger,
	diskUsage *DiskUsageCalculator,
) *CanonicalRepoService {
	return &CanonicalRepoService{
		gitEngine:    gitEngine,
		wsStorage:    wsStorage,
		projectsRoot: projectsRoot,
		logger:       logger,
		diskUsage:    diskUsage,
	}
}

// List returns a list of all canonical repository names.
func (c *CanonicalRepoService) List() ([]string, error) {
	return c.gitEngine.List()
}

// Add clones a new repository to the canonical store and returns the canonical name.
func (c *CanonicalRepoService) Add(url string) (string, error) {
	name := repoNameFromURL(url)
	if name == "" {
		return "", cerrors.NewInvalidArgument("url", fmt.Sprintf("could not determine repo name from URL: %s", url))
	}

	return name, c.gitEngine.Clone(url, name)
}

// Remove removes a repository from the canonical store.
func (c *CanonicalRepoService) Remove(name string, force bool) error {
	// 1. Check if repo is used by any workspace
	usedBy, err := c.GetWorkspacesUsingRepo(name)
	if err != nil {
		return err
	}

	if len(usedBy) > 0 {
		if !force {
			return cerrors.NewRepoInUse(name, usedBy)
		}

		// Log warning when force removing an in-use repo
		if c.logger != nil {
			c.logger.Warn("Force removing repository that is in use",
				"repo", name,
				"workspaces", usedBy,
				"warning", "These workspaces will have orphaned worktrees")
		}
	}

	// 2. Remove repo
	path := filepath.Join(c.projectsRoot, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cerrors.NewRepoNotFound(name)
	}

	if err := os.RemoveAll(path); err != nil {
		return cerrors.NewIOFailed(fmt.Sprintf("remove repo %s", name), err)
	}

	return nil
}

// PreviewRemove returns a preview of what would happen when removing a repo.
func (c *CanonicalRepoService) PreviewRemove(name string) (*domain.RepoRemovePreview, error) {
	path := filepath.Join(c.projectsRoot, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, cerrors.NewRepoNotFound(name)
	}

	// Find workspaces using this repo
	usedBy, err := c.GetWorkspacesUsingRepo(name)
	if err != nil {
		return nil, err
	}

	// Calculate disk usage
	var usage int64

	if c.diskUsage != nil {
		usage, _, _ = c.diskUsage.Calculate(path)
	}

	return &domain.RepoRemovePreview{
		RepoName:           name,
		RepoPath:           path,
		DiskUsageBytes:     usage,
		WorkspacesAffected: usedBy,
	}, nil
}

// Sync fetches updates for a canonical repository.
func (c *CanonicalRepoService) Sync(name string) error {
	return c.gitEngine.Fetch(name)
}

// GetWorkspacesUsingRepo returns the IDs of workspaces that use the given canonical repo.
func (c *CanonicalRepoService) GetWorkspacesUsingRepo(repoName string) ([]string, error) {
	workspaceMap, err := c.wsStorage.List()
	if err != nil {
		return nil, cerrors.NewIOFailed("list workspaces", err)
	}

	var usingRepo []string

	for _, ws := range workspaceMap {
		for _, repo := range ws.Repos {
			if repo.Name == repoName {
				usingRepo = append(usingRepo, ws.ID)
				break
			}
		}
	}

	return usingRepo, nil
}
