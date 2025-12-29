package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/output"
)

// WorkspaceHealthReport is the JSON output format for workspace health checks.
type WorkspaceHealthReport struct {
	Workspaces []domain.WorkspaceHealthReport `json:"workspaces"`
	Summary    string                         `json:"summary"`
	ExitCode   int                            `json:"exit_code"`
	Timestamp  time.Time                      `json:"timestamp"`
}

var doctorWorkspaceCmd = &cobra.Command{
	Use:   "workspace [ID]",
	Short: "Check workspace health and integrity",
	Long: `Perform comprehensive health checks on workspaces.

Checks performed:
  - Worktree integrity (.git file validity and back-references)
  - Metadata consistency (workspace.yaml vs disk contents)
  - Git config validity (readable and well-formed)
  - Remote URL format validation

Health scores:
  - healthy:  All checks passed
  - warning:  Non-critical issues found
  - critical: Critical issues that need attention

Exit codes:
  0 - All workspaces healthy
  1 - Warnings present
  2 - Critical issues present`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDoctorWorkspace,
}

func init() {
	doctorCmd.AddCommand(doctorWorkspaceCmd)
	doctorWorkspaceCmd.Flags().Bool("json", false, "Output results as JSON")
	doctorWorkspaceCmd.Flags().Bool("fix", false, "Attempt to auto-fix issues where possible")
}

func runDoctorWorkspace(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	fix, _ := cmd.Flags().GetBool("fix")

	app, err := getApp(cmd)
	if err != nil {
		return err
	}

	var reports []domain.WorkspaceHealthReport

	if len(args) > 0 {
		// Check specific workspace
		workspaceID := args[0]

		report, err := app.Service.CheckWorkspaceHealth(cmd.Context(), workspaceID, fix)
		if err != nil {
			return err
		}

		reports = append(reports, *report)
	} else {
		// Check all workspaces
		var err error

		reports, err = app.Service.CheckAllWorkspacesHealth(cmd.Context(), fix)
		if err != nil {
			return err
		}
	}

	fullReport := buildWorkspaceHealthReport(reports)

	if err := outputWorkspaceHealthReport(cmd.OutOrStdout(), fullReport, jsonOutput); err != nil {
		return err
	}

	if fullReport.ExitCode != 0 {
		return NewExitCodeError(fullReport.ExitCode, "")
	}

	return nil
}

// buildWorkspaceHealthReport creates a full report with summary and exit code.
func buildWorkspaceHealthReport(reports []domain.WorkspaceHealthReport) *WorkspaceHealthReport {
	result := &WorkspaceHealthReport{
		Workspaces: reports,
		Timestamp:  time.Now(),
	}

	var critical, warnings, healthy int

	for _, ws := range reports {
		switch ws.OverallStatus {
		case domain.HealthStatusCritical:
			critical++
		case domain.HealthStatusWarning:
			warnings++
		case domain.HealthStatusHealthy:
			healthy++
		}
	}

	switch {
	case critical > 0:
		result.ExitCode = 2
		result.Summary = fmt.Sprintf("%d critical, %d warning, %d healthy", critical, warnings, healthy)
	case warnings > 0:
		result.ExitCode = 1
		result.Summary = fmt.Sprintf("%d warning, %d healthy", warnings, healthy)
	default:
		result.ExitCode = 0
		result.Summary = fmt.Sprintf("All %d workspace(s) healthy", healthy)
	}

	return result
}

// outputWorkspaceHealthReport writes the report to the given writer.
func outputWorkspaceHealthReport(out io.Writer, report *WorkspaceHealthReport, jsonOutput bool) error {
	if jsonOutput {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(out, string(data))

		return nil
	}

	printWorkspaceHealthReport(out, report)

	return nil
}

// printWorkspaceHealthReport outputs the report in human-readable format.
func printWorkspaceHealthReport(out io.Writer, report *WorkspaceHealthReport) {
	_, _ = fmt.Fprintln(out, "Workspace Health Check")
	_, _ = fmt.Fprintln(out, output.SeparatorLine(output.SeparatorWidth))
	_, _ = fmt.Fprintln(out)

	if len(report.Workspaces) == 0 {
		_, _ = fmt.Fprintln(out, "  No active workspaces found.")
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, output.SeparatorLine(output.SeparatorWidth))
		_, _ = fmt.Fprintf(out, "Summary: %s\n", report.Summary)

		return
	}

	for _, ws := range report.Workspaces {
		printWorkspaceReport(out, ws)
	}

	_, _ = fmt.Fprintln(out, output.SeparatorLine(output.SeparatorWidth))
	_, _ = fmt.Fprintf(out, "Summary: %s\n", report.Summary)
}

// printWorkspaceReport outputs a single workspace's health report.
func printWorkspaceReport(out io.Writer, ws domain.WorkspaceHealthReport) {
	statusSymbol := healthStatusSymbol(ws.OverallStatus)
	style := healthStatusStyle(ws.OverallStatus)

	headerLine := fmt.Sprintf("%s Workspace: %s (%s)", statusSymbol, ws.WorkspaceID, ws.OverallStatus)
	_, _ = fmt.Fprintln(out, style(headerLine))

	// Group checks by category for cleaner output
	byCategory := groupChecksByCategory(ws.Checks)

	for _, category := range []domain.HealthCheckCategory{
		domain.HealthCategoryMetadata,
		domain.HealthCategoryWorktree,
		domain.HealthCategoryGitConfig,
		domain.HealthCategoryRemote,
	} {
		checks, ok := byCategory[category]
		if !ok || len(checks) == 0 {
			continue
		}

		_, _ = fmt.Fprintf(out, "  [%s]\n", category)

		for _, check := range checks {
			printHealthCheck(out, check)
		}
	}

	// Print fixes applied
	if len(ws.FixesApplied) > 0 {
		_, _ = fmt.Fprintln(out, "  [Fixes Applied]")
		for _, fix := range ws.FixesApplied {
			_, _ = fmt.Fprintf(out, "    ⚡ %s\n", fix)
		}
	}

	_, _ = fmt.Fprintln(out)
}

// printHealthCheck outputs a single health check result.
func printHealthCheck(out io.Writer, check domain.HealthCheck) {
	symbol := healthStatusSymbol(check.Status)
	style := healthStatusStyle(check.Status)

	if check.Status == domain.HealthStatusHealthy {
		_, _ = fmt.Fprintf(out, "    %s %s\n", symbol, check.Description)
		return
	}

	line := fmt.Sprintf("    %s %s: %s", symbol, check.Name, check.Description)
	_, _ = fmt.Fprintln(out, style(line))

	if check.Details != "" {
		_, _ = fmt.Fprintf(out, "        %s\n", check.Details)
	}

	if check.Fixable && check.FixAction != "" {
		_, _ = fmt.Fprintf(out, "        Suggestion: %s\n", check.FixAction)
	}
}

// groupChecksByCategory organizes checks by their category.
func groupChecksByCategory(checks []domain.HealthCheck) map[domain.HealthCheckCategory][]domain.HealthCheck {
	result := make(map[domain.HealthCheckCategory][]domain.HealthCheck)
	for _, check := range checks {
		result[check.Category] = append(result[check.Category], check)
	}

	return result
}

// healthStatusSymbol returns a symbol for the health status.
func healthStatusSymbol(status domain.HealthStatus) string {
	switch status {
	case domain.HealthStatusHealthy:
		return "✓"
	case domain.HealthStatusWarning:
		return "⚠"
	case domain.HealthStatusCritical:
		return "✗"
	default:
		return "?"
	}
}

// healthStatusStyle returns a style function for the health status.
func healthStatusStyle(status domain.HealthStatus) func(string) string {
	switch status {
	case domain.HealthStatusCritical:
		return func(text string) string { return output.Colorize(output.ErrorStyle, text) }
	case domain.HealthStatusWarning:
		return func(text string) string { return output.Colorize(output.WarningStyle, text) }
	case domain.HealthStatusHealthy:
		return func(text string) string { return output.Colorize(output.SuccessStyle, text) }
	default:
		return func(text string) string { return text }
	}
}
