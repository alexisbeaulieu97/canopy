// Package workspaces contains workspace-level business logic.
package workspaces

import (
	"bufio"
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

	workspacePath := filepath.Join(s.config.GetWorkspacesRoot(), dirName)

	// Run all health checks
	checks := []domain.HealthCheck{}

	// 1. Metadata consistency check
	metadataChecks := s.checkMetadataConsistency(ws, workspacePath)
	checks = append(checks, metadataChecks...)

	// 2. Worktree integrity checks for each repo
	for _, repo := range ws.Repos {
		worktreePath := filepath.Join(workspacePath, repo.Name)
		worktreeChecks := s.checkWorktreeIntegrity(ctx, repo.Name, worktreePath, ws.BranchName, fix, report)
		checks = append(checks, worktreeChecks...)
	}

	// 3. Git config validity checks for each repo
	for _, repo := range ws.Repos {
		worktreePath := filepath.Join(workspacePath, repo.Name)
		configChecks := s.checkGitConfig(repo.Name, worktreePath)
		checks = append(checks, configChecks...)
	}

	// 4. Remote URL validity checks for each repo
	for _, repo := range ws.Repos {
		worktreePath := filepath.Join(workspacePath, repo.Name)
		remoteChecks := s.checkRemoteURLValidity(repo.Name, worktreePath)
		checks = append(checks, remoteChecks...)
	}

	report.Checks = checks

	// Calculate overall status
	report.OverallStatus = calculateOverallStatus(checks)

	return report, nil
}

// checkMetadataConsistency validates workspace.yaml against actual worktrees on disk.
//
//nolint:gocyclo // Health check functions are inherently complex due to multiple validation conditions
func (s *WorkspaceHealthService) checkMetadataConsistency(ws *domain.Workspace, workspacePath string) []domain.HealthCheck {
	var checks []domain.HealthCheck

	// Check workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		checks = append(checks, domain.HealthCheck{
			Name:        "workspace_directory",
			Category:    domain.HealthCategoryMetadata,
			Status:      domain.HealthStatusCritical,
			Description: "Workspace directory does not exist",
			Fixable:     false,
			Details:     "Directory: " + workspacePath,
		})

		return checks
	}

	// Check each repo in metadata exists on disk
	for _, repo := range ws.Repos {
		repoPath := filepath.Join(workspacePath, repo.Name)
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			checks = append(checks, domain.HealthCheck{
				Name:        "repo_exists:" + repo.Name,
				Category:    domain.HealthCategoryMetadata,
				Status:      domain.HealthStatusCritical,
				Description: "Repository in metadata not found on disk",
				Fixable:     false,
				Details:     "Expected path: " + repoPath,
			})
		}
	}

	// Scan disk for directories not in metadata
	entries, err := os.ReadDir(workspacePath)
	if err == nil {
		repoSet := make(map[string]bool)
		for _, repo := range ws.Repos {
			repoSet[repo.Name] = true
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			// Skip hidden directories
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
	}

	// If no issues found, add a passing check
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

// checkWorktreeIntegrity validates that a worktree's .git file points back to the canonical repo.
//
//nolint:gocyclo // Health check functions are inherently complex due to multiple validation conditions
func (s *WorkspaceHealthService) checkWorktreeIntegrity(ctx context.Context, repoName, worktreePath, branchName string, fix bool, report *domain.WorkspaceHealthReport) []domain.HealthCheck {
	var checks []domain.HealthCheck

	gitPath := filepath.Join(worktreePath, ".git")

	// Check if .git exists
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

	// For worktrees, .git should be a file (not a directory) pointing to the canonical repo
	if info.IsDir() {
		// This is a regular git repo, not a worktree - could be an issue
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

	// Read the .git file to verify it points to the canonical repo
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

	// Parse the gitdir reference
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
	// Resolve relative paths from the worktree directory
	gitdirPath = resolveGitdirPath(gitdirPath, worktreePath)

	// Verify the gitdir path exists
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

	// Verify the gitdir points back to us (check the worktree file in the canonical repo)
	worktreeFile := filepath.Join(gitdirPath, "gitdir")
	if backRef, err := os.ReadFile(worktreeFile); err == nil { //nolint:gosec // G304: worktreeFile is constructed from git paths, not user input
		backRefPath := strings.TrimSpace(string(backRef))
		expectedPath := filepath.Join(worktreePath, ".git")

		// Normalize paths for comparison
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

	// All checks passed
	checks = append(checks, domain.HealthCheck{
		Name:        "worktree:" + repoName,
		Category:    domain.HealthCategoryWorktree,
		Status:      domain.HealthStatusHealthy,
		Description: "Worktree integrity verified",
	})

	return checks
}

// checkGitConfig validates that git config is readable for a worktree.
func (s *WorkspaceHealthService) checkGitConfig(repoName, worktreePath string) []domain.HealthCheck {
	var checks []domain.HealthCheck

	gitPath := filepath.Join(worktreePath, ".git")

	// For worktrees, .git is a file, so we need to follow the gitdir reference
	info, err := os.Stat(gitPath)
	if err != nil {
		// Already reported in worktree integrity check
		return checks
	}

	var configPath string
	if info.IsDir() {
		// Regular git repo
		configPath = filepath.Join(gitPath, "config")
	} else {
		// Worktree - read the gitdir reference
		content, err := os.ReadFile(gitPath) //nolint:gosec // G304: gitPath is constructed from workspace/repo paths, not user input
		if err != nil {
			return checks
		}

		gitdirLine := strings.TrimSpace(string(content))
		if !strings.HasPrefix(gitdirLine, "gitdir:") {
			return checks
		}

		gitdirPath := strings.TrimSpace(strings.TrimPrefix(gitdirLine, "gitdir:"))
		// Resolve relative paths from the worktree directory
		gitdirPath = resolveGitdirPath(gitdirPath, worktreePath)
		// Config is in the worktree-specific directory and in the parent canonical repo
		// For now, just verify the worktree gitdir exists
		if _, err := os.Stat(gitdirPath); err != nil { //nolint:gosec // G703: gitdirPath parsed from trusted .git worktree link file
			return checks
		}

		configPath = filepath.Join(gitdirPath, "config")
	}

	// Check if config file exists and is readable
	if _, err := os.Stat(configPath); os.IsNotExist(err) { //nolint:gosec // G703: configPath derived from trusted .git worktree link file
		checks = append(checks, domain.HealthCheck{
			Name:        "git_config:" + repoName,
			Category:    domain.HealthCategoryGitConfig,
			Status:      domain.HealthStatusWarning,
			Description: "Git config file not found",
			Fixable:     false,
			Details:     "Expected: " + configPath,
		})

		return checks
	}

	// Try to read the config file
	if _, err := os.ReadFile(configPath); err != nil { //nolint:gosec // G304: configPath is constructed from git directory paths, not user input
		checks = append(checks, domain.HealthCheck{
			Name:        "git_config:" + repoName,
			Category:    domain.HealthCategoryGitConfig,
			Status:      domain.HealthStatusCritical,
			Description: "Cannot read git config file",
			Fixable:     false,
			Details:     "Error: " + err.Error(),
		})

		return checks
	}

	checks = append(checks, domain.HealthCheck{
		Name:        "git_config:" + repoName,
		Category:    domain.HealthCategoryGitConfig,
		Status:      domain.HealthStatusHealthy,
		Description: "Git config is readable",
	})

	return checks
}

// checkRemoteURLValidity validates that remote URLs are in a valid format.
func (s *WorkspaceHealthService) checkRemoteURLValidity(repoName, worktreePath string) []domain.HealthCheck {
	var checks []domain.HealthCheck

	gitPath := filepath.Join(worktreePath, ".git")

	// For worktrees, we need to find the canonical repo's config
	info, err := os.Stat(gitPath)
	if err != nil {
		return checks
	}

	var configPath string
	if info.IsDir() {
		configPath = filepath.Join(gitPath, "config")
	} else {
		// Worktree - follow the gitdir to find the canonical repo
		content, err := os.ReadFile(gitPath) //nolint:gosec // G304: gitPath is constructed from workspace/repo paths, not user input
		if err != nil {
			return checks
		}

		gitdirLine := strings.TrimSpace(string(content))
		if !strings.HasPrefix(gitdirLine, "gitdir:") {
			return checks
		}

		gitdirPath := strings.TrimSpace(strings.TrimPrefix(gitdirLine, "gitdir:"))
		// Resolve relative paths from the worktree directory
		gitdirPath = resolveGitdirPath(gitdirPath, worktreePath)

		// Navigate from worktree gitdir to canonical repo config
		// Worktree gitdir is typically: <canonical>/.git/worktrees/<name>
		// So config is at: <canonical>/.git/config
		parentDir := filepath.Dir(gitdirPath)
		if filepath.Base(parentDir) == "worktrees" {
			canonicalGitDir := filepath.Dir(parentDir)
			configPath = filepath.Join(canonicalGitDir, "config")
		} else {
			// Fallback - use the worktree gitdir config
			configPath = filepath.Join(gitdirPath, "config")
		}
	}

	// Parse config to find remote URLs
	remoteURL, err := parseRemoteURL(configPath)
	if err != nil {
		// Config parsing failed - not a critical issue for URL validation
		return checks
	}

	if remoteURL == "" {
		checks = append(checks, domain.HealthCheck{
			Name:        "remote_url:" + repoName,
			Category:    domain.HealthCategoryRemote,
			Status:      domain.HealthStatusWarning,
			Description: "No remote URL configured",
			Fixable:     false,
		})

		return checks
	}

	// Validate URL format
	if !isValidRemoteURL(remoteURL) {
		checks = append(checks, domain.HealthCheck{
			Name:        "remote_url:" + repoName,
			Category:    domain.HealthCategoryRemote,
			Status:      domain.HealthStatusWarning,
			Description: "Remote URL format appears invalid",
			Fixable:     false,
			Details:     "URL: " + remoteURL,
		})

		return checks
	}

	checks = append(checks, domain.HealthCheck{
		Name:        "remote_url:" + repoName,
		Category:    domain.HealthCategoryRemote,
		Status:      domain.HealthStatusHealthy,
		Description: "Remote URL format is valid",
	})

	return checks
}

// fixMissingWorktree attempts to recreate a missing worktree.
func (s *WorkspaceHealthService) fixMissingWorktree(ctx context.Context, repoName, worktreePath, branchName string) error {
	// Verify canonical repo exists before attempting fix
	canonicalPath := filepath.Join(s.config.GetProjectsRoot(), repoName)
	if _, err := os.Stat(canonicalPath); os.IsNotExist(err) {
		return cerrors.NewRepoNotFound(repoName)
	}

	// Remove any leftover directory
	if err := os.RemoveAll(worktreePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Use provided branch name, default to DefaultBranchName only if empty
	if branchName == "" {
		branchName = DefaultBranchName
	}

	// Recreate the worktree using the git engine
	if err := s.gitEngine.CreateWorktree(ctx, repoName, worktreePath, branchName); err != nil {
		return err
	}

	return nil
}

// parseRemoteURL extracts the origin remote URL from a git config file.
func parseRemoteURL(configPath string) (string, error) {
	file, err := os.Open(configPath) //nolint:gosec // G304: configPath is constructed from workspace directory structure, not user input
	if err != nil {
		return "", err
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	inRemoteOrigin := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "[remote \"origin\"]" {
			inRemoteOrigin = true
			continue
		}

		if inRemoteOrigin {
			if strings.HasPrefix(line, "[") {
				// Entered a new section
				break
			}

			if strings.HasPrefix(line, "url = ") {
				return strings.TrimPrefix(line, "url = "), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

// isValidRemoteURL checks if a URL is a valid git remote URL format.
func isValidRemoteURL(rawURL string) bool {
	// Handle SSH-style URLs (git@host:path)
	if strings.HasPrefix(rawURL, "git@") && strings.Contains(rawURL, ":") {
		return true
	}

	// Handle standard URLs
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Must have a scheme and host
	validSchemes := map[string]bool{
		"https": true,
		"http":  true,
		"git":   true,
		"ssh":   true,
		"file":  true,
	}

	if !validSchemes[parsed.Scheme] {
		return false
	}

	if parsed.Host == "" && parsed.Scheme != "file" {
		return false
	}

	return true
}

// resolveGitdirPath resolves a gitdir path, handling both absolute and relative paths.
// Relative paths are resolved from the worktree directory.
func resolveGitdirPath(gitdirPath, worktreePath string) string {
	if filepath.IsAbs(gitdirPath) {
		return gitdirPath
	}
	// Relative path - resolve from worktree directory
	return filepath.Join(worktreePath, gitdirPath)
}

// calculateOverallStatus determines the worst status from a set of checks.
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
