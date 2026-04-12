package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/alexisbeaulieu97/canopy/internal/output"
)

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

	if err := output.WriteStructuredReport(cmd.OutOrStdout(), report, jsonOutput, func(outWriter io.Writer) {
		printHumanReport(outWriter, report)
	}); err != nil {
		return err
	}

	// Return ExitCodeError for non-zero exit codes
	// This allows Cobra to run cleanup while still signaling the exit code
	if report.ExitCode != 0 {
		return NewExitCodeError(report.ExitCode, "")
	}

	return nil
}
