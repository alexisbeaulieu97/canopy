// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// ExportService defines the interface for workspace export/import operations.
type ExportService interface {
	// ExportWorkspace creates a portable export of a workspace definition.
	ExportWorkspace(ctx context.Context, workspaceID string) (*domain.WorkspaceExport, error)

	// ImportWorkspace creates a workspace from an exported definition.
	ImportWorkspace(ctx context.Context, export *domain.WorkspaceExport, idOverride, branchOverride string, force bool) (string, error)
}

// WorkspaceCreator is the interface for creating workspaces (used to avoid circular dependencies).
type WorkspaceCreator interface {
	CreateWorkspace(ctx context.Context, id, branchName string, repos []domain.Repo) (string, error)
	CloseWorkspace(ctx context.Context, workspaceID string, force bool) error
}

// WorkspaceExportService handles export/import operations for workspaces.
type WorkspaceExportService struct {
	config           ports.ConfigProvider
	workspaceFinder  WorkspaceFinder
	workspaceCreator WorkspaceCreator
}

// NewExportService creates a new WorkspaceExportService.
func NewExportService(
	cfg ports.ConfigProvider,
	finder WorkspaceFinder,
	creator WorkspaceCreator,
) *WorkspaceExportService {
	return &WorkspaceExportService{
		config:           cfg,
		workspaceFinder:  finder,
		workspaceCreator: creator,
	}
}

// ExportWorkspace creates a portable export of a workspace definition.
func (s *WorkspaceExportService) ExportWorkspace(_ context.Context, workspaceID string) (*domain.WorkspaceExport, error) {
	workspace, _, err := s.workspaceFinder.FindWorkspace(workspaceID)
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
		if registry := s.config.GetRegistry(); registry != nil {
			if entry, ok := registry.ResolveByURL(repo.URL); ok {
				repoExport.Alias = entry.Alias
			}
		}

		export.Repos = append(export.Repos, repoExport)
	}

	return export, nil
}

// ImportWorkspace creates a workspace from an exported definition.
func (s *WorkspaceExportService) ImportWorkspace(ctx context.Context, export *domain.WorkspaceExport, idOverride, branchOverride string, force bool) (string, error) {
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
	if err := s.prepareForImport(ctx, workspaceID, force); err != nil {
		return "", err
	}

	// Resolve repos from export
	repos, err := s.resolveExportedRepos(export.Repos, workspaceID)
	if err != nil {
		return "", err
	}

	// Create the workspace
	return s.workspaceCreator.CreateWorkspace(ctx, workspaceID, branchName, repos)
}

// resolveImportOverrides determines the final workspace ID and branch name for import.
func (s *WorkspaceExportService) resolveImportOverrides(export *domain.WorkspaceExport, idOverride, branchOverride string) (string, string) {
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
func (s *WorkspaceExportService) prepareForImport(ctx context.Context, workspaceID string, force bool) error {
	_, _, findErr := s.workspaceFinder.FindWorkspace(workspaceID)
	if findErr == nil {
		// Workspace exists
		if !force {
			return cerrors.NewWorkspaceExists(workspaceID).WithContext("hint", "Use --force to overwrite or --id to specify a different ID")
		}
		// Force mode: delete existing workspace
		if err := s.workspaceCreator.CloseWorkspace(ctx, workspaceID, true); err != nil {
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
func (s *WorkspaceExportService) resolveExportedRepos(exportedRepos []domain.RepoExport, workspaceID string) ([]domain.Repo, error) {
	repos := make([]domain.Repo, 0, len(exportedRepos))

	for _, exported := range exportedRepos {
		var repo domain.Repo

		var resolved bool

		// Try registry alias first if available.
		// When alias resolves, we use the registry's canonical name (entry.Alias) rather than
		// the exported name. This ensures consistency with the local registry and handles cases
		// where the exporting machine used a different alias for the same repo.
		if exported.Alias != "" {
			if registry := s.config.GetRegistry(); registry != nil {
				if entry, ok := registry.Resolve(exported.Alias); ok {
					repo = domain.Repo{Name: entry.Alias, URL: entry.URL}
					resolved = true
				}
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
