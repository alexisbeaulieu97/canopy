package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/output"
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

	statusPass  = "pass"
	statusFail  = "fail"
	statusFixed = "fixed"
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

// buildDoctorReport runs all checks and returns the report.
func buildDoctorReport(ctx context.Context, fix bool) *DoctorReport {
	report := &DoctorReport{
		Checks:    []CheckResult{},
		Timestamp: time.Now(),
	}

	cfg, configErr := loadConfigForDoctor()

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
	var errorsCount, warningsCount int

	for _, check := range report.Checks {
		if check.Status != statusFail {
			continue
		}

		switch check.Severity {
		case SeverityError:
			errorsCount++
		case SeverityWarning:
			warningsCount++
		}
	}

	switch {
	case errorsCount > 0:
		report.ExitCode = 2
		report.Summary = fmt.Sprintf("%d error(s), %d warning(s)", errorsCount, warningsCount)
	case warningsCount > 0:
		report.ExitCode = 1
		report.Summary = fmt.Sprintf("%d warning(s)", warningsCount)
	default:
		report.ExitCode = 0
		report.Summary = "All checks passed"
	}
}

// printHumanReport outputs the doctor report in human-readable format.
// Write errors are intentionally ignored as this is CLI output with no recovery path.
func printHumanReport(out io.Writer, report *DoctorReport) {
	output.WriteReportHeader(out, "Canopy Doctor", output.SeparatorWidth)

	for _, check := range report.Checks {
		style := severityStyle(check.Severity)
		symbol := statusSymbol(check.Status)

		if check.Status == statusPass {
			_, _ = fmt.Fprintf(out, "  %s %s: %s\n", symbol, check.Name, check.Message)
			continue
		}

		line := fmt.Sprintf("  %s %s: %s", symbol, check.Name, check.Message)
		_, _ = fmt.Fprintln(out, style(line))

		if check.Details != "" {
			_, _ = fmt.Fprintf(out, "      %s\n", check.Details)
		}
	}

	output.WriteReportSummary(out, report.Summary, output.SeparatorWidth)
}

// loadConfigForDoctor attempts to load config without failing on missing files.
func loadConfigForDoctor() (*config.Config, error) {
	cfg, err := config.Load("")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// statusSymbol returns a symbol for the check status.
func statusSymbol(status string) string {
	switch status {
	case statusPass:
		return "✓"
	case statusFail:
		return "✗"
	case statusFixed:
		return "⚡"
	default:
		return "?"
	}
}

func severityStyle(sev CheckSeverity) func(string) string {
	switch sev {
	case SeverityError:
		return func(text string) string { return output.Colorize(output.ErrorStyle, text) }
	case SeverityWarning:
		return func(text string) string { return output.Colorize(output.WarningStyle, text) }
	case SeverityInfo:
		return func(text string) string { return output.Colorize(output.InfoStyle, text) }
	default:
		return func(text string) string { return text }
	}
}
