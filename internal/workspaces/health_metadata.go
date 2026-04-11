package workspaces

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// checkMetadataConsistency validates workspace.yaml against actual worktrees on disk.
//
//nolint:gocyclo // Health check functions are inherently complex due to multiple validation conditions
func (s *WorkspaceHealthService) checkMetadataConsistency(ws *domain.Workspace, workspacePath string) []domain.HealthCheck {
	var checks []domain.HealthCheck

	if _, err := os.Stat(workspacePath); err != nil {
		if os.IsNotExist(err) {
			checks = append(checks, domain.HealthCheck{
				Name:        "workspace_directory",
				Category:    domain.HealthCategoryMetadata,
				Status:      domain.HealthStatusCritical,
				Description: "Workspace directory does not exist",
				Fixable:     false,
				Details:     "Directory: " + workspacePath,
			})
		} else {
			checks = append(checks, domain.HealthCheck{
				Name:        "workspace_directory",
				Category:    domain.HealthCategoryMetadata,
				Status:      domain.HealthStatusCritical,
				Description: "Cannot access workspace directory",
				Fixable:     false,
				Details:     "Error: " + err.Error(),
			})
		}

		return checks
	}

	for _, repo := range ws.Repos {
		repoPath := filepath.Join(workspacePath, repo.Name)
		if _, err := os.Stat(repoPath); err != nil {
			check := domain.HealthCheck{
				Name:     "repo_exists:" + repo.Name,
				Category: domain.HealthCategoryMetadata,
				Status:   domain.HealthStatusCritical,
				Fixable:  false,
			}

			if os.IsNotExist(err) {
				check.Description = "Repository in metadata not found on disk"
				check.Details = "Expected path: " + repoPath
			} else {
				check.Description = "Cannot access repository path from metadata"
				check.Details = "Error: " + err.Error()
			}

			checks = append(checks, check)
		}
	}

	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		checks = append(checks, domain.HealthCheck{
			Name:        "workspace_directory_scan",
			Category:    domain.HealthCategoryMetadata,
			Status:      domain.HealthStatusCritical,
			Description: "Cannot read workspace directory",
			Fixable:     false,
			Details:     "Error: " + err.Error(),
		})
	}

	repoSet := make(map[string]bool)
	for _, repo := range ws.Repos {
		repoSet[repo.Name] = true
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		if !repoSet[name] {
			checks = append(checks, domain.HealthCheck{
				Name:        "untracked_dir:" + name,
				Category:    domain.HealthCategoryMetadata,
				Status:      domain.HealthStatusWarning,
				Description: "Directory on disk not tracked in workspace metadata",
				Fixable:     false,
				Details:     "Path: " + filepath.Join(workspacePath, name),
			})
		}
	}

	if len(checks) == 0 {
		checks = append(checks, domain.HealthCheck{
			Name:        "metadata_consistency",
			Category:    domain.HealthCategoryMetadata,
			Status:      domain.HealthStatusHealthy,
			Description: "Workspace metadata matches disk contents",
		})
	}

	return checks
}

func calculateOverallStatus(checks []domain.HealthCheck) domain.HealthStatus {
	status := domain.HealthStatusHealthy

	for _, check := range checks {
		switch check.Status {
		case domain.HealthStatusCritical:
			return domain.HealthStatusCritical
		case domain.HealthStatusWarning:
			status = domain.HealthStatusWarning
		}
	}

	return status
}
