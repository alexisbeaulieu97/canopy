package hooks

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
)

func TestExecuteHooks_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	hooks := []config.Hook{
		{Command: "echo hello"},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos:         []domain.Repo{},
	}

	err := executor.ExecuteHooks(hooks, ctx, false)
	if err != nil {
		t.Fatalf("ExecuteHooks failed: %v", err)
	}
}

func TestExecuteHooks_CommandFailed(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	hooks := []config.Hook{
		{Command: "exit 1"},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos:         []domain.Repo{},
	}

	err := executor.ExecuteHooks(hooks, ctx, false)
	if err == nil {
		t.Fatal("Expected error for failed hook")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		t.Fatalf("Expected CanopyError, got %T", err)
	}

	if canopyErr.Code != cerrors.ErrHookFailed {
		t.Errorf("Expected ErrHookFailed, got %s", canopyErr.Code)
	}
}

func TestExecuteHooks_ContinueOnError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	// Create a marker file to verify second hook ran
	markerFile := filepath.Join(tmpDir, "marker")

	hooks := []config.Hook{
		{Command: "exit 1"},
		{Command: "touch " + markerFile},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos:         []domain.Repo{},
	}

	// With continueOnError=true, should continue despite first hook failing
	err := executor.ExecuteHooks(hooks, ctx, true)
	if err != nil {
		t.Fatalf("ExecuteHooks should succeed with continueOnError=true: %v", err)
	}

	// Verify second hook ran
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("Second hook should have run and created marker file")
	}
}

func TestExecuteHooks_HookContinueOnError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	// Create a marker file to verify second hook ran
	markerFile := filepath.Join(tmpDir, "marker")

	hooks := []config.Hook{
		{Command: "exit 1", ContinueOnError: true},
		{Command: "touch " + markerFile},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos:         []domain.Repo{},
	}

	err := executor.ExecuteHooks(hooks, ctx, false)
	if err != nil {
		t.Fatalf("ExecuteHooks should succeed when hook has ContinueOnError=true: %v", err)
	}

	// Verify second hook ran
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("Second hook should have run and created marker file")
	}
}

func TestExecuteHooks_Timeout(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	hooks := []config.Hook{
		{Command: "sleep 10", Timeout: 1}, // 1 second timeout
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos:         []domain.Repo{},
	}

	start := time.Now()
	err := executor.ExecuteHooks(hooks, ctx, false)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("Expected timeout error")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		t.Fatalf("Expected CanopyError, got %T", err)
	}

	if canopyErr.Code != cerrors.ErrHookTimeout {
		t.Errorf("Expected ErrHookTimeout, got %s", canopyErr.Code)
	}

	// Verify it didn't wait the full 10 seconds
	if duration > 3*time.Second {
		t.Errorf("Timeout didn't work, took %v", duration)
	}
}

func TestExecuteHooks_EnvironmentVariables(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	outputFile := filepath.Join(tmpDir, "env_output")

	hooks := []config.Hook{
		{Command: "echo $CANOPY_WORKSPACE_ID,$CANOPY_BRANCH > " + outputFile},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws-123",
		WorkspacePath: tmpDir,
		BranchName:    "feature/test",
		Repos:         []domain.Repo{},
	}

	err := executor.ExecuteHooks(hooks, ctx, false)
	if err != nil {
		t.Fatalf("ExecuteHooks failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expected := "test-ws-123,feature/test\n"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestExecuteHooks_RepoFilter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	// Create repo directories
	frontendDir := filepath.Join(tmpDir, "frontend")
	backendDir := filepath.Join(tmpDir, "backend")

	if err := os.MkdirAll(frontendDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatal(err)
	}

	frontendMarker := filepath.Join(frontendDir, "marker")
	backendMarker := filepath.Join(backendDir, "marker")

	hooks := []config.Hook{
		{
			Command: "touch marker",
			Repos:   []string{"frontend"}, // Only run in frontend
		},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos: []domain.Repo{
			{Name: "frontend", URL: "https://example.com/frontend.git"},
			{Name: "backend", URL: "https://example.com/backend.git"},
		},
	}

	err := executor.ExecuteHooks(hooks, ctx, false)
	if err != nil {
		t.Fatalf("ExecuteHooks failed: %v", err)
	}

	// Frontend marker should exist
	if _, err := os.Stat(frontendMarker); os.IsNotExist(err) {
		t.Error("Hook should have run in frontend repo")
	}

	// Backend marker should NOT exist
	if _, err := os.Stat(backendMarker); !os.IsNotExist(err) {
		t.Error("Hook should NOT have run in backend repo")
	}
}

func TestExecuteHooks_RepoEnvironmentVariables(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	// Create repo directory
	repoDir := filepath.Join(tmpDir, "myrepo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(repoDir, "env_output")

	hooks := []config.Hook{
		{
			Command: "echo $CANOPY_REPO_NAME > " + outputFile,
			Repos:   []string{"myrepo"},
		},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos: []domain.Repo{
			{Name: "myrepo", URL: "https://example.com/myrepo.git"},
		},
	}

	err := executor.ExecuteHooks(hooks, ctx, false)
	if err != nil {
		t.Fatalf("ExecuteHooks failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expected := "myrepo\n"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestExecuteHooks_WorkingDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	outputFile := filepath.Join(tmpDir, "pwd_output")

	hooks := []config.Hook{
		{Command: "pwd > " + outputFile},
	}

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos:         []domain.Repo{},
	}

	err := executor.ExecuteHooks(hooks, ctx, false)
	if err != nil {
		t.Fatalf("ExecuteHooks failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Working directory should be the workspace path
	got := strings.TrimSpace(string(content))

	// Handle symlinks (macOS /tmp -> /private/tmp)
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks for tmpDir: %v", err)
	}

	realGot, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks for pwd output: %v", err)
	}

	if realGot != realTmpDir {
		t.Errorf("Working directory mismatch: expected %s, got %s", realTmpDir, realGot)
	}
}

func TestFilterRepos(t *testing.T) {
	t.Parallel()

	repos := []domain.Repo{
		{Name: "frontend", URL: "https://example.com/frontend.git"},
		{Name: "backend", URL: "https://example.com/backend.git"},
		{Name: "api", URL: "https://example.com/api.git"},
	}

	filtered := filterRepos(repos, []string{"frontend", "api"})

	if len(filtered) != 2 {
		t.Fatalf("Expected 2 repos, got %d", len(filtered))
	}

	names := make(map[string]bool)
	for _, r := range filtered {
		names[r.Name] = true
	}

	if !names["frontend"] || !names["api"] {
		t.Errorf("Expected frontend and api, got %v", filtered)
	}

	if names["backend"] {
		t.Error("backend should not be in filtered list")
	}
}

func TestExecuteHooks_EmptyHooks(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := logging.New(false)
	executor := NewExecutor(logger)

	ctx := domain.HookContext{
		WorkspaceID:   "test-ws",
		WorkspacePath: tmpDir,
		BranchName:    "main",
		Repos:         []domain.Repo{},
	}

	// Empty hooks should succeed
	err := executor.ExecuteHooks([]config.Hook{}, ctx, false)
	if err != nil {
		t.Fatalf("Empty hooks should succeed: %v", err)
	}

	// Nil hooks should also succeed
	err = executor.ExecuteHooks(nil, ctx, false)
	if err != nil {
		t.Fatalf("Nil hooks should succeed: %v", err)
	}
}
