package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// mockLogger captures logged errors for testing.
type mockLogger struct {
	errors []string
}

func (m *mockLogger) Errorf(format string, _ ...interface{}) {
	m.errors = append(m.errors, format)
}

func TestSaveRegistryWithRollback_SuccessfulSave(t *testing.T) {
	// Create a temp directory and registry
	tmpDir := t.TempDir()
	registryPath := tmpDir + "/repos.yaml"

	registry, err := config.LoadRepoRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	// Pre-populate with an entry
	if err := registry.Register("test-alias", ports.RegistryEntry{URL: "https://github.com/test/repo"}, false); err != nil {
		t.Fatalf("failed to register test entry: %v", err)
	}

	rollbackCalled := false
	rollbackFn := func() error {
		rollbackCalled = true
		return nil
	}

	logger := &mockLogger{}

	err = saveRegistryWithRollback(registry, rollbackFn, "test operation", logger)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if rollbackCalled {
		t.Error("rollback should not be called on successful save")
	}

	if len(logger.errors) > 0 {
		t.Errorf("expected no logged errors, got: %v", logger.errors)
	}
}

func TestSaveRegistryWithRollback_SaveFailureWithSuccessfulRollback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows - directory permissions behave differently")
	}

	// Create a temp directory and registry
	tmpDir := t.TempDir()
	registryPath := filepath.Join(tmpDir, "repos.yaml")

	registry, err := config.LoadRepoRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	// Add an entry so registry has content
	if err := registry.Register("test-alias", ports.RegistryEntry{URL: "https://github.com/test/repo"}, false); err != nil {
		t.Fatalf("failed to register test entry: %v", err)
	}

	// First save to create the file
	if err := registry.Save(); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Make the file read-only to cause Save() to fail
	if err := os.Chmod(registryPath, 0o444); err != nil {
		t.Fatalf("failed to make file read-only: %v", err)
	}
	// Restore permissions at the end so cleanup works
	t.Cleanup(func() {
		_ = os.Chmod(registryPath, 0o644)
	})

	rollbackCalled := false
	rollbackFn := func() error {
		rollbackCalled = true
		return nil
	}

	logger := &mockLogger{}

	err = saveRegistryWithRollback(registry, rollbackFn, "test operation", logger)
	if err == nil {
		t.Error("expected error when save fails, got nil")
	}

	if !rollbackCalled {
		t.Error("rollback should be called when save fails")
	}

	// Rollback save will also fail due to read-only file, so we should see a log
	if len(logger.errors) == 0 {
		t.Error("expected logged error for failed rollback save")
	}
}

func TestSaveRegistryWithRollback_RollbackFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows - directory permissions behave differently")
	}

	// Create a temp directory and registry
	tmpDir := t.TempDir()
	registryPath := filepath.Join(tmpDir, "repos.yaml")

	registry, err := config.LoadRepoRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	// Add an entry so registry has content
	if err := registry.Register("test-alias", ports.RegistryEntry{URL: "https://github.com/test/repo"}, false); err != nil {
		t.Fatalf("failed to register test entry: %v", err)
	}

	// First save to create the file
	if err := registry.Save(); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Make the file read-only to cause Save() to fail
	if err := os.Chmod(registryPath, 0o444); err != nil {
		t.Fatalf("failed to make file read-only: %v", err)
	}
	// Restore permissions at the end so cleanup works
	t.Cleanup(func() {
		_ = os.Chmod(registryPath, 0o644)
	})

	// Rollback function that returns an error
	rollbackFn := func() error {
		return os.ErrPermission
	}

	logger := &mockLogger{}

	err = saveRegistryWithRollback(registry, rollbackFn, "test operation", logger)
	if err == nil {
		t.Error("expected error when save fails, got nil")
	}

	// Should log the rollback failure
	if len(logger.errors) == 0 {
		t.Error("expected logged error for failed rollback")
	}
}

func TestSaveRegistryWithRollback_NilLogger(t *testing.T) {
	// Test that nil logger doesn't cause panic
	tmpDir := t.TempDir()
	registryPath := tmpDir + "/repos.yaml"

	registry, err := config.LoadRepoRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	// Add an entry so registry has content
	if err := registry.Register("test-alias", ports.RegistryEntry{URL: "https://github.com/test/repo"}, false); err != nil {
		t.Fatalf("failed to register test entry: %v", err)
	}

	rollbackFn := func() error {
		return nil
	}

	// Should not panic with nil logger
	err = saveRegistryWithRollback(registry, rollbackFn, "test operation", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRegisterAlias_Success(t *testing.T) {
	tmpDir := t.TempDir()
	registryPath := tmpDir + "/repos.yaml"

	registry, err := config.LoadRepoRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	entry := ports.RegistryEntry{URL: "https://github.com/test/repo"}
	logger := &mockLogger{}

	alias, err := registerAlias(registry, "test-alias", entry, logger)
	if err != nil {
		t.Fatalf("registerAlias failed: %v", err)
	}

	if alias != "test-alias" {
		t.Errorf("expected alias 'test-alias', got '%s'", alias)
	}

	// Verify the entry was persisted
	resolved, exists := registry.Resolve("test-alias")
	if !exists {
		t.Error("expected alias to exist after registration")
	}

	if resolved.URL != "https://github.com/test/repo" {
		t.Errorf("expected URL 'https://github.com/test/repo', got '%s'", resolved.URL)
	}
}

func TestRegisterAlias_DuplicateError(t *testing.T) {
	tmpDir := t.TempDir()
	registryPath := tmpDir + "/repos.yaml"

	registry, err := config.LoadRepoRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	entry := ports.RegistryEntry{URL: "https://github.com/test/repo"}
	logger := &mockLogger{}

	// First registration should succeed
	_, err = registerAlias(registry, "test-alias", entry, logger)
	if err != nil {
		t.Fatalf("first registerAlias failed: %v", err)
	}

	// Second registration with same alias should fail
	_, err = registerAlias(registry, "test-alias", ports.RegistryEntry{URL: "https://github.com/other/repo"}, logger)
	if err == nil {
		t.Error("expected error for duplicate alias, got nil")
	}
}

func TestRegisterRegistryEntry_RollbackRestoresPreviousEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows - file permissions behave differently")
	}

	tmpDir := t.TempDir()
	registryPath := filepath.Join(tmpDir, "repos.yaml")

	registry, err := config.LoadRepoRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}

	original := ports.RegistryEntry{URL: "https://github.com/test/original"}
	if err := registry.Register("test-alias", original, false); err != nil {
		t.Fatalf("failed to register original entry: %v", err)
	}

	if err := registry.Save(); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	if err := os.Chmod(registryPath, 0o444); err != nil {
		t.Fatalf("failed to make registry read-only: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(registryPath, 0o644)
	})

	logger := &mockLogger{}

	_, err = registerRegistryEntry(
		registry,
		" test-alias ",
		ports.RegistryEntry{URL: "https://github.com/test/updated"},
		true,
		logger,
	)
	if err == nil {
		t.Fatal("expected registration save failure, got nil")
	}

	resolved, exists := registry.Resolve("test-alias")
	if !exists {
		t.Fatal("expected original alias to be restored after rollback")
	}

	if resolved.URL != original.URL {
		t.Fatalf("expected rollback to restore %q, got %q", original.URL, resolved.URL)
	}
}

func TestCanonicalPathStatus_StatFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows - file permissions behave differently")
	}

	tmpDir := t.TempDir()
	parent := filepath.Join(tmpDir, "projects")
	repoPath := filepath.Join(parent, "repo")

	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("failed to create repo path: %v", err)
	}

	if err := os.Chmod(parent, 0o000); err != nil {
		t.Fatalf("failed to remove parent permissions: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(parent, 0o755)
	})

	_, err := canonicalPathStatus(repoPath)
	if err == nil {
		t.Fatal("expected stat failure, got nil")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		t.Fatalf("expected canopy error, got %T", err)
	}

	if canopyErr.Code != cerrors.ErrIOFailed {
		t.Fatalf("expected IO_FAILED, got %s", canopyErr.Code)
	}
}
