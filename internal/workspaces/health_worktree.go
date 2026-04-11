package workspaces

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// checkWorktreeIntegrity validates that a worktree's .git file points back to the canonical repo.
//
//nolint:gocyclo // Health check functions are inherently complex due to multiple validation conditions
func (s *WorkspaceHealthService) checkWorktreeIntegrity(ctx context.Context, repoName, worktreePath, branchName string, fix bool, report *domain.WorkspaceHealthReport) []domain.HealthCheck {
	var checks []domain.HealthCheck

	gitPath := filepath.Join(worktreePath, ".git")

	info, err := os.Stat(gitPath)
	if os.IsNotExist(err) {
		check := domain.HealthCheck{
			Name:        "worktree_git:" + repoName,
			Category:    domain.HealthCategoryWorktree,
			Status:      domain.HealthStatusCritical,
			Description: "Missing .git file/directory in worktree",
			Fixable:     true,
			FixAction:   "Re-create worktree from canonical repository",
			Details:     "Path: " + gitPath,
		}

		if fix {
			if fixErr := s.fixMissingWorktree(ctx, repoName, worktreePath, branchName); fixErr == nil {
				check.Status = domain.HealthStatusHealthy
				check.Description = "Worktree recreated successfully"

				report.FixesApplied = append(report.FixesApplied, "recreated worktree for "+repoName)
			} else {
				check.Details += "; Fix failed: " + fixErr.Error()
			}
		}

		checks = append(checks, check)

		return checks
	}

	if err != nil {
		checks = append(checks, domain.HealthCheck{
			Name:        "worktree_git:" + repoName,
			Category:    domain.HealthCategoryWorktree,
			Status:      domain.HealthStatusCritical,
			Description: "Cannot access .git file/directory",
			Fixable:     false,
			Details:     "Error: " + err.Error(),
		})

		return checks
	}

	if info.IsDir() {
		checks = append(checks, domain.HealthCheck{
			Name:        "worktree_type:" + repoName,
			Category:    domain.HealthCategoryWorktree,
			Status:      domain.HealthStatusWarning,
			Description: ".git is a directory instead of a worktree link file",
			Fixable:     false,
			Details:     "Expected worktree link, found regular git directory",
		})

		return checks
	}

	content, err := os.ReadFile(gitPath) //nolint:gosec // G304: gitPath is constructed from workspace/repo paths, not user input
	if err != nil {
		checks = append(checks, domain.HealthCheck{
			Name:        "worktree_link:" + repoName,
			Category:    domain.HealthCategoryWorktree,
			Status:      domain.HealthStatusCritical,
			Description: "Cannot read .git worktree link file",
			Fixable:     false,
			Details:     "Error: " + err.Error(),
		})

		return checks
	}

	gitdirLine := strings.TrimSpace(string(content))
	if !strings.HasPrefix(gitdirLine, "gitdir:") {
		checks = append(checks, domain.HealthCheck{
			Name:        "worktree_format:" + repoName,
			Category:    domain.HealthCategoryWorktree,
			Status:      domain.HealthStatusCritical,
			Description: "Invalid .git worktree link format",
			Fixable:     false,
			Details:     "Expected 'gitdir: <path>', got: " + gitdirLine,
		})

		return checks
	}

	gitdirPath := strings.TrimSpace(strings.TrimPrefix(gitdirLine, "gitdir:"))
	gitdirPath = resolveGitdirPath(gitdirPath, worktreePath)

	if _, err := os.Stat(gitdirPath); os.IsNotExist(err) { //nolint:gosec // G703: gitdirPath parsed from trusted .git worktree link file
		check := domain.HealthCheck{
			Name:        "worktree_ref:" + repoName,
			Category:    domain.HealthCategoryWorktree,
			Status:      domain.HealthStatusCritical,
			Description: "Worktree link points to non-existent git directory",
			Fixable:     true,
			FixAction:   "Re-create worktree from canonical repository",
			Details:     "Missing: " + gitdirPath,
		}

		if fix {
			if fixErr := s.fixMissingWorktree(ctx, repoName, worktreePath, branchName); fixErr == nil {
				check.Status = domain.HealthStatusHealthy
				check.Description = "Worktree recreated successfully"

				report.FixesApplied = append(report.FixesApplied, "recreated worktree for "+repoName)
			} else {
				check.Details += "; Fix failed: " + fixErr.Error()
			}
		}

		checks = append(checks, check)

		return checks
	}

	worktreeFile := filepath.Join(gitdirPath, "gitdir")
	if backRef, err := os.ReadFile(worktreeFile); err == nil { //nolint:gosec // G304: worktreeFile is constructed from git paths, not user input
		backRefPath := strings.TrimSpace(string(backRef))
		expectedPath := filepath.Join(worktreePath, ".git")

		backRefPath = filepath.Clean(backRefPath)
		expectedPath = filepath.Clean(expectedPath)

		if backRefPath != expectedPath {
			checks = append(checks, domain.HealthCheck{
				Name:        "worktree_backref:" + repoName,
				Category:    domain.HealthCategoryWorktree,
				Status:      domain.HealthStatusWarning,
				Description: "Worktree back-reference mismatch",
				Fixable:     false,
				Details:     "Expected: " + expectedPath + ", Got: " + backRefPath,
			})

			return checks
		}
	}

	checks = append(checks, domain.HealthCheck{
		Name:        "worktree:" + repoName,
		Category:    domain.HealthCategoryWorktree,
		Status:      domain.HealthStatusHealthy,
		Description: "Worktree integrity verified",
	})

	return checks
}

func (s *WorkspaceHealthService) fixMissingWorktree(ctx context.Context, repoName, worktreePath, branchName string) error {
	canonicalPath := filepath.Join(s.config.GetProjectsRoot(), repoName)
	if _, err := os.Stat(canonicalPath); os.IsNotExist(err) {
		return cerrors.NewRepoNotFound(repoName)
	}

	if err := os.RemoveAll(worktreePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	if branchName == "" {
		branchName = DefaultBranchName
	}

	if err := s.gitEngine.CreateWorktree(ctx, repoName, worktreePath, branchName); err != nil {
		return err
	}

	return nil
}

func resolveGitdirPath(gitdirPath, worktreePath string) string {
	if filepath.IsAbs(gitdirPath) {
		return gitdirPath
	}

	return filepath.Join(worktreePath, gitdirPath)
}
