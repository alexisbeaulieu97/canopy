package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Note: These tests cannot run in parallel because they modify package-level
// version variables that are shared across tests.

func TestVersionCommand_TextOutput(t *testing.T) {
	// Reset version vars for test
	originalVersion := version
	originalCommit := commit
	originalBuildDate := buildDate

	defer func() {
		version = originalVersion
		commit = originalCommit
		buildDate = originalBuildDate
	}()

	version = "v1.2.3"
	commit = "abc1234"
	buildDate = "2025-01-15T10:30:00Z"

	// Execute version command
	output := executeVersionCmd(t, false)

	// Verify format
	if !strings.Contains(output, "canopy version v1.2.3") {
		t.Errorf("expected version line, got: %s", output)
	}

	if !strings.Contains(output, "commit: abc1234") {
		t.Errorf("expected commit line, got: %s", output)
	}

	if !strings.Contains(output, "built: 2025-01-15T10:30:00Z") {
		t.Errorf("expected built line, got: %s", output)
	}

	if !strings.Contains(output, "go: go") {
		t.Errorf("expected go version line, got: %s", output)
	}
}

func TestVersionCommand_JSONOutput(t *testing.T) {
	// Reset version vars for test
	originalVersion := version
	originalCommit := commit
	originalBuildDate := buildDate

	defer func() {
		version = originalVersion
		commit = originalCommit
		buildDate = originalBuildDate
	}()

	version = "v1.2.3"
	commit = "abc1234"
	buildDate = "2025-01-15T10:30:00Z"

	// Execute version command with --json
	output := executeVersionCmd(t, true)

	// Parse JSON output
	var info VersionInfo

	err := json.Unmarshal([]byte(output), &info)
	if err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if info.Version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %s", info.Version)
	}

	if info.Commit != "abc1234" {
		t.Errorf("expected commit abc1234, got %s", info.Commit)
	}

	if info.BuildDate != "2025-01-15T10:30:00Z" {
		t.Errorf("expected build_date 2025-01-15T10:30:00Z, got %s", info.BuildDate)
	}

	if info.GoVersion == "" {
		t.Error("expected go_version to be non-empty")
	}
}

func TestVersionFlag(t *testing.T) {
	// Reset version vars for test
	originalVersion := version

	defer func() {
		version = originalVersion
	}()

	version = "v2.0.0"

	// Test that printVersion outputs the expected short format
	var buf bytes.Buffer
	printVersion(&buf)

	output := buf.String()
	expected := "canopy version v2.0.0\n"

	if output != expected {
		t.Errorf("printVersion() = %q, want %q", output, expected)
	}
}

// executeVersionCmd runs the version command in isolation and returns captured output.
func executeVersionCmd(t *testing.T, jsonOutput bool) string {
	t.Helper()

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a standalone version command for testing
	cmd := &cobra.Command{
		Use: "version",
		RunE: func(c *cobra.Command, _ []string) error {
			return versionCmd.RunE(c, nil)
		},
	}
	cmd.Flags().Bool("json", false, "Output version information as JSON")
	cmd.SetOut(&buf)

	if jsonOutput {
		cmd.SetArgs([]string{"--json"})
	} else {
		cmd.SetArgs([]string{})
	}

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	return buf.String()
}
