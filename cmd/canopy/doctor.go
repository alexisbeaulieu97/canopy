package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/config"
)

// CheckSeverity indicates the severity level of a check result.
type CheckSeverity string

const (
	// SeverityInfo is for informational messages.
	SeverityInfo CheckSeverity = "info"
	// SeverityWarning indicates a non-critical issue.
	SeverityWarning CheckSeverity = "warning"
	// SeverityError indicates a critical issue.
	SeverityError CheckSeverity = "error"
)

// CheckResult represents the outcome of a single diagnostic check.
type CheckResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"` // "pass", "fail", "fixed"
	Severity CheckSeverity `json:"severity"`
	Message  string        `json:"message"`
	Details  string        `json:"details,omitempty"`
}

// DoctorReport holds all check results.
type DoctorReport struct {
	Checks    []CheckResult `json:"checks"`
	Summary   string        `json:"summary"`
	ExitCode  int           `json:"exit_code"`
	Timestamp time.Time     `json:"timestamp"`
}

// statusSymbol returns a symbol for the check status.
func statusSymbol(status string) string {
	switch status {
	case "pass":
		return "✓"
	case "fail":
		return "✗"
	case "fixed":
		return "⚡"
	default:
		return "?"
	}
}

// severityColor returns ANSI color code for severity.
func severityColor(sev CheckSeverity) string {
	switch sev {
	case SeverityError:
		return "\033[31m" // red
	case SeverityWarning:
		return "\033[33m" // yellow
	case SeverityInfo:
		return "\033[36m" // cyan
	default:
		return "\033[0m" // reset
	}
}

// colorResetDoctor avoids redeclaration with repo.go.
const colorResetDoctor = "\033[0m"

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose environment and configuration issues",
	Long: `Validate Canopy environment and configuration, reporting issues with actionable guidance.

Checks performed:
  - Git installation and version
  - Configuration file validity
  - Directory existence and permissions
  - Canonical repository health

Exit codes:
  0 - All checks pass
  1 - Warnings present (non-critical issues)
  2 - Errors present (critical issues)`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().Bool("json", false, "Output results as JSON")
	doctorCmd.Flags().Bool("fix", false, "Attempt to auto-fix simple issues")
}

func runDoctor(cmd *cobra.Command, _ []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	fix, _ := cmd.Flags().GetBool("fix")

	report := buildDoctorReport(cmd.Context(), fix)
	calculateReportSummary(report)

	// Output results
	if err := outputDoctorReport(cmd.OutOrStdout(), report, jsonOutput); err != nil {
		return err
	}

	// Return ExitCodeError for non-zero exit codes
	// This allows Cobra to run cleanup while still signaling the exit code
	if report.ExitCode != 0 {
		return NewExitCodeError(report.ExitCode, "")
	}

	return nil
}

// buildDoctorReport runs all checks and returns the report.
func buildDoctorReport(ctx context.Context, fix bool) *DoctorReport {
	report := &DoctorReport{
		Checks:    []CheckResult{},
		Timestamp: time.Now(),
	}

	// Load config for checks (using lenient loading that doesn't fail on missing config)
	cfg, configErr := loadConfigForDoctor()

	// Run all checks
	report.Checks = append(report.Checks, checkGitInstalled())
	report.Checks = append(report.Checks, checkConfigFile(configErr))

	if cfg != nil {
		report.Checks = append(report.Checks, checkDirectory("projects_root", cfg.GetProjectsRoot(), fix)...)
		report.Checks = append(report.Checks, checkDirectory("workspaces_root", cfg.GetWorkspacesRoot(), fix)...)
		report.Checks = append(report.Checks, checkDirectory("closed_root", cfg.GetClosedRoot(), fix)...)
		report.Checks = append(report.Checks, checkCanonicalRepos(ctx, cfg)...)
	}

	return report
}

// calculateReportSummary sets the exit code and summary based on check results.
func calculateReportSummary(report *DoctorReport) {
	var errors, warnings int

	for _, c := range report.Checks {
		if c.Status == "fail" {
			switch c.Severity {
			case SeverityError:
				errors++
			case SeverityWarning:
				warnings++
			}
		}
	}

	switch {
	case errors > 0:
		report.ExitCode = 2
		report.Summary = fmt.Sprintf("%d error(s), %d warning(s)", errors, warnings)
	case warnings > 0:
		report.ExitCode = 1
		report.Summary = fmt.Sprintf("%d warning(s)", warnings)
	default:
		report.ExitCode = 0
		report.Summary = "All checks passed"
	}
}

// outputDoctorReport writes the report to the given writer.
func outputDoctorReport(out io.Writer, report *DoctorReport, jsonOutput bool) error {
	if jsonOutput {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(out, string(data))
	} else {
		printHumanReport(out, report)
	}

	return nil
}

// printHumanReport outputs the doctor report in human-readable format.
// Write errors are intentionally ignored as this is CLI output with no recovery path.
func printHumanReport(out io.Writer, report *DoctorReport) {
	_, _ = fmt.Fprintln(out, "Canopy Doctor")
	_, _ = fmt.Fprintln(out, strings.Repeat("=", 50))
	_, _ = fmt.Fprintln(out)

	for _, c := range report.Checks {
		color := severityColor(c.Severity)
		symbol := statusSymbol(c.Status)

		if c.Status == "pass" {
			_, _ = fmt.Fprintf(out, "  %s %s: %s\n", symbol, c.Name, c.Message)
		} else {
			_, _ = fmt.Fprintf(out, "  %s%s %s: %s%s\n", color, symbol, c.Name, c.Message, colorResetDoctor)

			if c.Details != "" {
				_, _ = fmt.Fprintf(out, "      %s\n", c.Details)
			}
		}
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, strings.Repeat("-", 50))
	_, _ = fmt.Fprintf(out, "Summary: %s\n", report.Summary)
}

// loadConfigForDoctor attempts to load config without failing on missing files.
func loadConfigForDoctor() (*config.Config, error) {
	cfg, err := config.Load("")
	if err != nil {
		return nil, err
	}
	// Skip validation - we'll report validation errors as check results
	return cfg, nil
}

// checkGitInstalled verifies git is installed and returns version info.
func checkGitInstalled() CheckResult {
	result := CheckResult{
		Name:     "Git Installation",
		Severity: SeverityError,
	}

	cmd := exec.Command("git", "--version")

	output, err := cmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = "git is not installed or not in PATH"
		result.Details = "Install git: https://git-scm.com/downloads"

		return result
	}

	version := strings.TrimSpace(string(output))
	result.Status = "pass"
	result.Message = version
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
		result.Status = "fail"
		result.Message = "configuration error"
		result.Details = configErr.Error()

		return result
	}

	result.Status = "pass"
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
			// Attempt to create the directory
			// User workspace directories need 0755 for proper access
			if mkErr := os.MkdirAll(path, 0o755); mkErr != nil { //nolint:gosec // G301: 0755 is intentional for user workspace directories
				result.Status = "fail"
				result.Message = fmt.Sprintf("directory does not exist: %s", path)
				result.Details = fmt.Sprintf("Failed to create: %v", mkErr)
			} else {
				result.Status = "fixed"
				result.Message = fmt.Sprintf("created directory: %s", path)
				result.Severity = SeverityInfo
			}
		} else {
			result.Status = "fail"
			result.Message = fmt.Sprintf("directory does not exist: %s", path)
			result.Details = "Run with --fix to create it"
		}

		results = append(results, result)

		return results
	}

	if err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("cannot access directory: %s", path)
		result.Details = err.Error()
		results = append(results, result)

		return results
	}

	if !info.IsDir() {
		result.Status = "fail"
		result.Message = fmt.Sprintf("path is not a directory: %s", path)
		results = append(results, result)

		return results
	}

	// Check write permission by attempting to create a temp file
	testFile := filepath.Join(path, ".canopy_doctor_test")

	f, err := os.Create(testFile) //nolint:gosec // G304: testFile is constructed from validated path parameter
	if err != nil {
		result.Status = "fail"
		result.Severity = SeverityError
		result.Message = fmt.Sprintf("directory not writable: %s", path)
		result.Details = err.Error()
		results = append(results, result)

		return results
	}

	_ = f.Close()
	_ = os.Remove(testFile)

	result.Status = "pass"
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
		// Directory doesn't exist or can't be read - already checked above
		return results
	}

	staleThreshold := cfg.GetStaleThresholdDays()
	if staleThreshold <= 0 {
		staleThreshold = 30 // default to 30 days for doctor check
	}

	staleCutoff := time.Now().AddDate(0, 0, -staleThreshold)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(projectsRoot, entry.Name())

		// Check if it's a git repo (bare or non-bare)
		// Bare repos have HEAD at root, non-bare repos have .git/HEAD
		gitDir := repoPath
		bareHead := filepath.Join(repoPath, "HEAD")
		nonBareHead := filepath.Join(repoPath, ".git", "HEAD")

		if _, err := os.Stat(bareHead); os.IsNotExist(err) {
			// Not a bare repo, check for non-bare
			if _, err := os.Stat(nonBareHead); os.IsNotExist(err) {
				continue // Not a git repo, skip
			}

			gitDir = filepath.Join(repoPath, ".git")
		}

		result := CheckResult{
			Name:     fmt.Sprintf("Repo: %s", entry.Name()),
			Severity: SeverityWarning,
		}

		// Check last fetch time
		fetchHead := filepath.Join(gitDir, "FETCH_HEAD")

		info, err := os.Stat(fetchHead)
		if err != nil {
			// FETCH_HEAD doesn't exist - repo may never have been fetched
			result.Status = "fail"
			result.Message = "never fetched"
			result.Details = "Run: canopy repo sync " + entry.Name()
			results = append(results, result)

			continue
		}

		if info.ModTime().Before(staleCutoff) {
			result.Status = "fail"
			result.Message = fmt.Sprintf("stale (last fetch: %s)", info.ModTime().Format("2006-01-02"))
			result.Details = "Run: canopy repo sync " + entry.Name()
			results = append(results, result)

			continue
		}

		result.Status = "pass"
		result.Message = fmt.Sprintf("healthy (last fetch: %s)", info.ModTime().Format("2006-01-02"))
		result.Severity = SeverityInfo
		results = append(results, result)
	}

	return results
}
