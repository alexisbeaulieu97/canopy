package workspaces

import (
	"bufio"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func (s *WorkspaceHealthService) checkGitConfig(repoName, worktreePath string) []domain.HealthCheck {
	var checks []domain.HealthCheck

	gitPath := filepath.Join(worktreePath, ".git")

	info, err := os.Stat(gitPath)
	if err != nil {
		return checks
	}

	var configPath string
	if info.IsDir() {
		configPath = filepath.Join(gitPath, "config")
	} else {
		content, err := os.ReadFile(gitPath) //nolint:gosec // G304: gitPath is constructed from workspace/repo paths, not user input
		if err != nil {
			return checks
		}

		gitdirLine := strings.TrimSpace(string(content))
		if !strings.HasPrefix(gitdirLine, "gitdir:") {
			return checks
		}

		gitdirPath := strings.TrimSpace(strings.TrimPrefix(gitdirLine, "gitdir:"))
		gitdirPath = resolveGitdirPath(gitdirPath, worktreePath)

		if _, err := os.Stat(gitdirPath); err != nil { //nolint:gosec // G703: gitdirPath parsed from trusted .git worktree link file
			return checks
		}

		configPath = resolveGitConfigPath(gitdirPath)
	}

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

func (s *WorkspaceHealthService) checkRemoteURLValidity(repoName, worktreePath string) []domain.HealthCheck {
	var checks []domain.HealthCheck

	gitPath := filepath.Join(worktreePath, ".git")

	info, err := os.Stat(gitPath)
	if err != nil {
		return checks
	}

	var configPath string
	if info.IsDir() {
		configPath = filepath.Join(gitPath, "config")
	} else {
		content, err := os.ReadFile(gitPath) //nolint:gosec // G304: gitPath is constructed from workspace/repo paths, not user input
		if err != nil {
			return checks
		}

		gitdirLine := strings.TrimSpace(string(content))
		if !strings.HasPrefix(gitdirLine, "gitdir:") {
			return checks
		}

		gitdirPath := strings.TrimSpace(strings.TrimPrefix(gitdirLine, "gitdir:"))
		gitdirPath = resolveGitdirPath(gitdirPath, worktreePath)

		configPath = resolveGitConfigPath(gitdirPath)
	}

	remoteURL, err := parseRemoteURL(configPath)
	if err != nil {
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
				break
			}

			key, value, ok := strings.Cut(line, "=")
			if ok && strings.TrimSpace(key) == "url" {
				return trimMatchingQuotes(strings.TrimSpace(value)), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

func isValidRemoteURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	if isLocalRemotePath(rawURL) || isSCPStyleRemote(rawURL) {
		return true
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

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

func resolveGitConfigPath(gitdirPath string) string {
	parentDir := filepath.Dir(gitdirPath)
	if filepath.Base(parentDir) == "worktrees" {
		return filepath.Join(filepath.Dir(parentDir), "config")
	}

	return filepath.Join(gitdirPath, "config")
}

func trimMatchingQuotes(value string) string {
	if len(value) < 2 {
		return value
	}

	if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
		return value[1 : len(value)-1]
	}

	return value
}

func isLocalRemotePath(rawURL string) bool {
	return filepath.IsAbs(rawURL) ||
		strings.HasPrefix(rawURL, "~/") ||
		strings.HasPrefix(rawURL, "./") ||
		strings.HasPrefix(rawURL, "../")
}

func isSCPStyleRemote(rawURL string) bool {
	if strings.Contains(rawURL, "://") || strings.Contains(rawURL, " ") {
		return false
	}

	hostPart, pathPart, ok := strings.Cut(rawURL, ":")
	if !ok || hostPart == "" || pathPart == "" {
		return false
	}

	if strings.Contains(hostPart, "/") {
		return false
	}

	return true
}
