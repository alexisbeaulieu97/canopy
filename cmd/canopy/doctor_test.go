package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckGitInstalled(t *testing.T) {
	result := checkGitInstalled()

	// Git should be installed in the test environment
	if result.Status != "pass" {
		t.Errorf("expected git to be installed, got status=%s, message=%s", result.Status, result.Message)
	}

	if result.Name != "Git Installation" {
		t.Errorf("expected name 'Git Installation', got %s", result.Name)
	}
}

func TestCheckDirectory_Exists(t *testing.T) {
	dir := t.TempDir()

	results := checkDirectory("test_dir", dir, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Status != "pass" {
		t.Errorf("expected status 'pass', got %s: %s", result.Status, result.Message)
	}
}

func TestCheckDirectory_NotExists_NoFix(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent")

	results := checkDirectory("test_dir", dir, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Status != "fail" {
		t.Errorf("expected status 'fail', got %s", result.Status)
	}

	if result.Severity != SeverityError {
		t.Errorf("expected severity 'error', got %s", result.Severity)
	}

	if result.Details != "Run with --fix to create it" {
		t.Errorf("expected details to suggest --fix, got %s", result.Details)
	}
}

func TestCheckDirectory_NotExists_WithFix(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tobecreated")

	results := checkDirectory("test_dir", dir, true)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Status != "fixed" {
		t.Errorf("expected status 'fixed', got %s: %s", result.Status, result.Message)
	}

	// Verify directory was created
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("created path is not a directory")
	}
}

func TestCheckDirectory_NotWritable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping test when running as root")
	}

	dir := t.TempDir()
	// Make directory read-only
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}

	defer func() { _ = os.Chmod(dir, 0o755) }() // Restore for cleanup

	results := checkDirectory("test_dir", dir, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Status != "fail" {
		t.Errorf("expected status 'fail' for non-writable dir, got %s: %s", result.Status, result.Message)
	}
}

func TestCheckConfigFile_Valid(t *testing.T) {
	result := checkConfigFile(nil)

	if result.Status != "pass" {
		t.Errorf("expected status 'pass' for nil error, got %s", result.Status)
	}

	if result.Name != "Configuration" {
		t.Errorf("expected name 'Configuration', got %s", result.Name)
	}
}

func TestCheckConfigFile_Invalid(t *testing.T) {
	testErr := os.ErrNotExist // Use a simple error for testing
	result := checkConfigFile(testErr)

	if result.Status != "fail" {
		t.Errorf("expected status 'fail' for error, got %s", result.Status)
	}

	if result.Severity != SeverityError {
		t.Errorf("expected severity 'error', got %s", result.Severity)
	}
}

func TestStatusSymbol(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"pass", "✓"},
		{"fail", "✗"},
		{"fixed", "⚡"},
		{"unknown", "?"},
	}

	for _, tt := range tests {
		got := statusSymbol(tt.status)
		if got != tt.expected {
			t.Errorf("statusSymbol(%q) = %q, want %q", tt.status, got, tt.expected)
		}
	}
}

func TestSeverityStyleNoColor(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	style := severityStyle(SeverityError)

	got := style("test")
	if got != "test" {
		t.Errorf("severityStyle should return unstyled text when color is disabled, got %q", got)
	}
}

func TestPrintHumanReport(t *testing.T) {
	report := &DoctorReport{
		Checks: []CheckResult{
			{Name: "Test Check", Status: "pass", Severity: SeverityInfo, Message: "all good"},
			{Name: "Failed Check", Status: "fail", Severity: SeverityError, Message: "something wrong", Details: "fix it"},
		},
		Summary:   "1 error(s), 0 warning(s)",
		ExitCode:  2,
		Timestamp: time.Now(),
	}

	var buf bytes.Buffer
	printHumanReport(&buf, report)

	output := buf.String()

	// Check that output contains expected elements
	if !bytes.Contains(buf.Bytes(), []byte("Canopy Doctor")) {
		t.Error("output should contain 'Canopy Doctor'")
	}

	if !bytes.Contains(buf.Bytes(), []byte("Test Check")) {
		t.Error("output should contain 'Test Check'")
	}

	if !bytes.Contains(buf.Bytes(), []byte("Failed Check")) {
		t.Error("output should contain 'Failed Check'")
	}

	if !bytes.Contains(buf.Bytes(), []byte("fix it")) {
		t.Error("output should contain details 'fix it'")
	}

	if !bytes.Contains(buf.Bytes(), []byte("Summary")) {
		t.Error("output should contain 'Summary'")
	}

	_ = output // Avoid unused variable
}

func TestDoctorReportJSON(t *testing.T) {
	report := &DoctorReport{
		Checks: []CheckResult{
			{Name: "Test", Status: "pass", Severity: SeverityInfo, Message: "ok"},
		},
		Summary:   "All checks passed",
		ExitCode:  0,
		Timestamp: time.Now(),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}

	// Verify it can be unmarshaled back
	var decoded DoctorReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal report: %v", err)
	}

	if len(decoded.Checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(decoded.Checks))
	}

	if decoded.Checks[0].Name != "Test" {
		t.Errorf("expected check name 'Test', got %s", decoded.Checks[0].Name)
	}
}

func TestCheckCanonicalRepos_EmptyDir(t *testing.T) {
	// Create a temp config with empty projects root
	tempDir := t.TempDir()

	// checkCanonicalRepos needs a config, so we create a minimal mock
	// For this test, we just verify it doesn't panic on empty dir
	results := checkCanonicalRepos(context.Background(), &mockDoctorConfig{projectsRoot: tempDir, staleThreshold: 30})

	// Should return empty results for empty directory
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty dir, got %d", len(results))
	}
}

func TestCheckCanonicalRepos_StaleRepo(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake repo directory with HEAD file (makes it look like a git repo)
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	// Create HEAD file to make it look like a git repo
	headFile := filepath.Join(repoDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("failed to create HEAD: %v", err)
	}

	// Create FETCH_HEAD with old timestamp
	fetchHead := filepath.Join(repoDir, "FETCH_HEAD")
	if err := os.WriteFile(fetchHead, []byte("abc123\n"), 0o644); err != nil {
		t.Fatalf("failed to create FETCH_HEAD: %v", err)
	}

	// Set modification time to 60 days ago
	oldTime := time.Now().AddDate(0, 0, -60)
	if err := os.Chtimes(fetchHead, oldTime, oldTime); err != nil {
		t.Fatalf("failed to change file time: %v", err)
	}

	results := checkCanonicalRepos(context.Background(), &mockDoctorConfig{projectsRoot: tempDir, staleThreshold: 30})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Status != "fail" {
		t.Errorf("expected status 'fail' for stale repo, got %s", results[0].Status)
	}

	if results[0].Severity != SeverityWarning {
		t.Errorf("expected severity 'warning' for stale repo, got %s", results[0].Severity)
	}
}

func TestCheckCanonicalRepos_NonBareRepo(t *testing.T) {
	tempDir := t.TempDir()

	// Create a non-bare repo directory with .git/HEAD
	repoDir := filepath.Join(tempDir, "non-bare-repo")

	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create .git/HEAD file
	headFile := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("failed to create HEAD: %v", err)
	}

	// Create .git/FETCH_HEAD with recent timestamp
	fetchHead := filepath.Join(gitDir, "FETCH_HEAD")
	if err := os.WriteFile(fetchHead, []byte("abc123\n"), 0o644); err != nil {
		t.Fatalf("failed to create FETCH_HEAD: %v", err)
	}

	results := checkCanonicalRepos(context.Background(), &mockDoctorConfig{projectsRoot: tempDir, staleThreshold: 30})

	if len(results) != 1 {
		t.Fatalf("expected 1 result for non-bare repo, got %d", len(results))
	}

	if results[0].Status != "pass" {
		t.Errorf("expected status 'pass' for healthy non-bare repo, got %s: %s", results[0].Status, results[0].Message)
	}

	if results[0].Name != "Repo: non-bare-repo" {
		t.Errorf("expected name 'Repo: non-bare-repo', got %s", results[0].Name)
	}
}

// mockDoctorConfig implements doctorConfig interface for testing.
type mockDoctorConfig struct {
	projectsRoot   string
	staleThreshold int
}

func (m *mockDoctorConfig) GetProjectsRoot() string    { return m.projectsRoot }
func (m *mockDoctorConfig) GetStaleThresholdDays() int { return m.staleThreshold }
