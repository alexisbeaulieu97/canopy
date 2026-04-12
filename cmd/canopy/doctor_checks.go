package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// checkGitInstalled verifies git is installed and returns version info.
func checkGitInstalled() CheckResult {
	result := CheckResult{
		Name:     "Git Installation",
		Severity: SeverityError,
	}

	cmd := exec.Command("git", "--version")

	outputBytes, err := cmd.Output()
	if err != nil {
		result.Status = statusFail
		result.Message = "git is not installed or not in PATH"
		result.Details = "Install git: https://git-scm.com/downloads"

		return result
	}

	result.Status = statusPass
	result.Message = strings.TrimSpace(string(outputBytes))
	result.Severity = SeverityInfo

	return result
}

// checkConfigFile verifies the config file is valid.
func checkConfigFile(configErr error) CheckResult {
	result := CheckResult{
		Name:     "Configuration",
		Severity: SeverityError,
	}

	if configErr != nil {
		result.Status = statusFail
		result.Message = "configuration error"
		result.Details = configErr.Error()

		return result
	}

	result.Status = statusPass
	result.Message = "configuration is valid"
	result.Severity = SeverityInfo

	return result
}

// checkDirectory verifies a directory exists and is writable.
func checkDirectory(name, path string, fix bool) []CheckResult {
	var results []CheckResult

	result := CheckResult{
		Name:     fmt.Sprintf("Directory: %s", name),
		Severity: SeverityError,
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if fix {
			if mkErr := os.MkdirAll(path, 0o755); mkErr != nil { //nolint:gosec // G301: 0755 is intentional for user workspace directories
				result.Status = statusFail
				result.Message = fmt.Sprintf("directory does not exist: %s", path)
				result.Details = fmt.Sprintf("Failed to create: %v", mkErr)
			} else {
				result.Status = statusFixed
				result.Message = fmt.Sprintf("created directory: %s", path)
				result.Severity = SeverityInfo
			}
		} else {
			result.Status = statusFail
			result.Message = fmt.Sprintf("directory does not exist: %s", path)
			result.Details = "Run with --fix to create it"
		}

		results = append(results, result)

		return results
	}

	if err != nil {
		result.Status = statusFail
		result.Message = fmt.Sprintf("cannot access directory: %s", path)
		result.Details = err.Error()
		results = append(results, result)

		return results
	}

	if !info.IsDir() {
		result.Status = statusFail
		result.Message = fmt.Sprintf("path is not a directory: %s", path)
		results = append(results, result)

		return results
	}

	testFile := filepath.Join(path, ".canopy_doctor_test")

	file, err := os.Create(testFile) //nolint:gosec // G304: testFile is constructed from validated path parameter
	if err != nil {
		result.Status = statusFail
		result.Severity = SeverityError
		result.Message = fmt.Sprintf("directory not writable: %s", path)
		result.Details = err.Error()
		results = append(results, result)

		return results
	}

	_ = file.Close()
	_ = os.Remove(testFile)

	result.Status = statusPass
	result.Message = fmt.Sprintf("directory exists and is writable: %s", path)
	result.Severity = SeverityInfo
	results = append(results, result)

	return results
}

// doctorConfig is the interface needed by doctor checks.
type doctorConfig interface {
	GetProjectsRoot() string
	GetStaleThresholdDays() int
}

// checkCanonicalRepos verifies health of canonical repositories.
func checkCanonicalRepos(_ context.Context, cfg doctorConfig) []CheckResult {
	var results []CheckResult

	projectsRoot := cfg.GetProjectsRoot()

	entries, err := os.ReadDir(projectsRoot)
	if err != nil {
		return results
	}

	staleThreshold := cfg.GetStaleThresholdDays()
	if staleThreshold <= 0 {
		staleThreshold = 30
	}

	staleCutoff := time.Now().AddDate(0, 0, -staleThreshold)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(projectsRoot, entry.Name())
		gitDir := repoPath
		bareHead := filepath.Join(repoPath, "HEAD")
		nonBareHead := filepath.Join(repoPath, ".git", "HEAD")

		if _, err := os.Stat(bareHead); os.IsNotExist(err) {
			if _, err := os.Stat(nonBareHead); os.IsNotExist(err) {
				continue
			}

			gitDir = filepath.Join(repoPath, ".git")
		}

		result := CheckResult{
			Name:     fmt.Sprintf("Repo: %s", entry.Name()),
			Severity: SeverityWarning,
		}

		fetchHead := filepath.Join(gitDir, "FETCH_HEAD")

		info, err := os.Stat(fetchHead)
		if err != nil {
			result.Status = statusFail
			result.Message = "never fetched"
			result.Details = "Run: canopy repo sync " + entry.Name()
			results = append(results, result)

			continue
		}

		if info.ModTime().Before(staleCutoff) {
			result.Status = statusFail
			result.Message = fmt.Sprintf("stale (last fetch: %s)", info.ModTime().Format("2006-01-02"))
			result.Details = "Run: canopy repo sync " + entry.Name()
			results = append(results, result)

			continue
		}

		result.Status = statusPass
		result.Message = fmt.Sprintf("healthy (last fetch: %s)", info.ModTime().Format("2006-01-02"))
		result.Severity = SeverityInfo
		results = append(results, result)
	}

	return results
}
