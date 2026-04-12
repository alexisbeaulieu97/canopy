// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"context"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

const (
	// DefaultBranchName is the fallback branch name when none is specified.
	DefaultBranchName = "main"
)

// WorkspaceHealthService handles health checks for workspaces.
type WorkspaceHealthService struct {
	config          ports.ConfigProvider
	gitEngine       ports.GitOperations
	wsEngine        ports.WorkspaceStorage
	logger          *logging.Logger
	workspaceFinder WorkspaceFinder
}

// NewHealthService creates a new WorkspaceHealthService.
func NewHealthService(
	cfg ports.ConfigProvider,
	gitEngine ports.GitOperations,
	wsEngine ports.WorkspaceStorage,
	logger *logging.Logger,
	finder WorkspaceFinder,
) *WorkspaceHealthService {
	return &WorkspaceHealthService{
		config:          cfg,
		gitEngine:       gitEngine,
		wsEngine:        wsEngine,
		logger:          logger,
		workspaceFinder: finder,
	}
}

// CheckWorkspace performs health checks on a specific workspace.
func (s *WorkspaceHealthService) CheckWorkspace(ctx context.Context, workspaceID string, fix bool) (*domain.WorkspaceHealthReport, error) {
	ws, dirName, err := s.workspaceFinder.FindWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return s.checkWorkspaceHealth(ctx, ws, dirName, fix)
}

// CheckAllWorkspaces performs health checks on all active workspaces.
func (s *WorkspaceHealthService) CheckAllWorkspaces(ctx context.Context, fix bool) ([]domain.WorkspaceHealthReport, error) {
	workspaces, err := s.wsEngine.List(ctx)
	if err != nil {
		return nil, cerrors.NewIOFailed("list workspaces", err)
	}

	var reports []domain.WorkspaceHealthReport

	for _, ws := range workspaces {
		if ctx.Err() != nil {
			return nil, cerrors.NewContextError(ctx, "health check", ws.ID)
		}

		dirName := ws.DirName
		if dirName == "" {
			dirName, err = s.config.ComputeWorkspaceDir(ws.ID)
			if err != nil {
				// Include error report for workspaces we couldn't check
				reports = append(reports, domain.WorkspaceHealthReport{
					WorkspaceID:   ws.ID,
					OverallStatus: domain.HealthStatusCritical,
					Checks: []domain.HealthCheck{{
						Name:        "workspace_access",
						Category:    domain.HealthCategoryMetadata,
						Status:      domain.HealthStatusCritical,
						Description: "Cannot determine workspace directory",
						Details:     err.Error(),
					}},
				})

				continue
			}
		}

		report, err := s.checkWorkspaceHealth(ctx, &ws, dirName, fix)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("health check failed for workspace", "workspace", ws.ID, "error", err)
			}

			// Include error report for workspaces that failed health check
			reports = append(reports, domain.WorkspaceHealthReport{
				WorkspaceID:   ws.ID,
				OverallStatus: domain.HealthStatusCritical,
				Checks: []domain.HealthCheck{{
					Name:        "health_check_error",
					Category:    domain.HealthCategoryMetadata,
					Status:      domain.HealthStatusCritical,
					Description: "Health check failed",
					Details:     err.Error(),
				}},
			})

			continue
		}

		reports = append(reports, *report)
	}

	return reports, nil
}

// checkWorkspaceHealth performs all health checks on a single workspace.
func (s *WorkspaceHealthService) checkWorkspaceHealth(ctx context.Context, ws *domain.Workspace, dirName string, fix bool) (*domain.WorkspaceHealthReport, error) {
	report := &domain.WorkspaceHealthReport{
		WorkspaceID:   ws.ID,
		OverallStatus: domain.HealthStatusHealthy,
		Checks:        []domain.HealthCheck{},
	}

	workspacePath := workspacePath(s.config.GetWorkspacesRoot(), dirName)

	// Run all health checks
	checks := []domain.HealthCheck{}

	// 1. Metadata consistency check
	metadataChecks := s.checkMetadataConsistency(ws, workspacePath)
	checks = append(checks, metadataChecks...)

	// 2. Worktree integrity checks for each repo
	for _, repo := range ws.Repos {
		worktreePath := workspaceRepoPath(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		worktreeChecks := s.checkWorktreeIntegrity(ctx, repo.Name, worktreePath, ws.BranchName, fix, report)
		checks = append(checks, worktreeChecks...)
	}

	// 3. Git config validity checks for each repo
	for _, repo := range ws.Repos {
		worktreePath := workspaceRepoPath(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		configChecks := s.checkGitConfig(repo.Name, worktreePath)
		checks = append(checks, configChecks...)
	}

	// 4. Remote URL validity checks for each repo
	for _, repo := range ws.Repos {
		worktreePath := workspaceRepoPath(s.config.GetWorkspacesRoot(), dirName, repo.Name)
		remoteChecks := s.checkRemoteURLValidity(repo.Name, worktreePath)
		checks = append(checks, remoteChecks...)
	}

	report.Checks = checks

	// Calculate overall status
	report.OverallStatus = calculateOverallStatus(checks)

	return report, nil
}
