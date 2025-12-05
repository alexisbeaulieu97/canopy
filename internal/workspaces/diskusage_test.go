package workspaces

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiskUsageCalculator_Calculate(t *testing.T) {
	t.Parallel()

	t.Run("empty directory returns zero", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		calc := DefaultDiskUsageCalculator()

		usage, _, err := calc.Calculate(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if usage != 0 {
			t.Errorf("expected 0 bytes, got %d", usage)
		}
	})

	t.Run("sums file sizes", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		// Create some test files
		file1 := filepath.Join(dir, "file1.txt")
		file2 := filepath.Join(dir, "file2.txt")

		if err := os.WriteFile(file1, []byte("hello"), 0o600); err != nil {
			t.Fatalf("failed to write file1: %v", err)
		}

		if err := os.WriteFile(file2, []byte("world!"), 0o600); err != nil {
			t.Fatalf("failed to write file2: %v", err)
		}

		calc := DefaultDiskUsageCalculator()

		usage, _, err := calc.Calculate(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// "hello" = 5 bytes, "world!" = 6 bytes
		expected := int64(11)
		if usage != expected {
			t.Errorf("expected %d bytes, got %d", expected, usage)
		}
	})

	t.Run("skips .git directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		// Create a .git directory with a file
		gitDir := filepath.Join(dir, ".git")
		if err := os.MkdirAll(gitDir, 0o750); err != nil {
			t.Fatalf("failed to create .git dir: %v", err)
		}

		gitFile := filepath.Join(gitDir, "config")
		if err := os.WriteFile(gitFile, []byte("large git file content here"), 0o600); err != nil {
			t.Fatalf("failed to write git file: %v", err)
		}

		// Create a regular file
		regularFile := filepath.Join(dir, "main.go")
		if err := os.WriteFile(regularFile, []byte("package main"), 0o600); err != nil {
			t.Fatalf("failed to write regular file: %v", err)
		}

		calc := DefaultDiskUsageCalculator()

		usage, _, err := calc.Calculate(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should only count "package main" = 12 bytes
		expected := int64(12)
		if usage != expected {
			t.Errorf("expected %d bytes (excluding .git), got %d", expected, usage)
		}
	})

	t.Run("returns latest modification time", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		// Create a file
		file := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(file, []byte("test"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		calc := DefaultDiskUsageCalculator()

		_, latest, err := calc.Calculate(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if latest.IsZero() {
			t.Error("expected non-zero latest time")
		}

		// Should be recent (within last minute)
		if time.Since(latest) > time.Minute {
			t.Errorf("expected recent modification time, got %v ago", time.Since(latest))
		}
	})
}

func TestDiskUsageCalculator_CachedUsage(t *testing.T) {
	t.Parallel()

	t.Run("caches results", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		file := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(file, []byte("hello"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		calc := NewDiskUsageCalculator(time.Hour) // Long TTL for testing

		// First call
		usage1, _, err := calc.CachedUsage(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Add more content
		if err := os.WriteFile(file, []byte("hello world"), 0o600); err != nil {
			t.Fatalf("failed to update file: %v", err)
		}

		// Second call should return cached value
		usage2, _, err := calc.CachedUsage(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if usage1 != usage2 {
			t.Errorf("expected cached value %d, got %d", usage1, usage2)
		}
	})

	t.Run("invalidate clears cache", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		file := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(file, []byte("hello"), 0o600); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		calc := NewDiskUsageCalculator(time.Hour)

		// First call
		usage1, _, err := calc.CachedUsage(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Add more content
		if err := os.WriteFile(file, []byte("hello world"), 0o600); err != nil {
			t.Fatalf("failed to update file: %v", err)
		}

		// Invalidate cache
		calc.InvalidateCache(dir)

		// Should get new value
		usage2, _, err := calc.CachedUsage(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if usage1 == usage2 {
			t.Errorf("expected different values after invalidation")
		}

		if usage2 != 11 { // "hello world" = 11 bytes
			t.Errorf("expected 11 bytes, got %d", usage2)
		}
	})
}

func TestDiskUsageCalculator_ClearCache(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	calc := NewDiskUsageCalculator(time.Hour)

	// Populate cache
	if _, _, err := calc.CachedUsage(dir); err != nil {
		t.Fatalf("failed to populate cache: %v", err)
	}

	// Clear cache
	calc.ClearCache()

	// Update file
	if err := os.WriteFile(file, []byte("updated content"), 0o600); err != nil {
		t.Fatalf("failed to update file: %v", err)
	}

	// Should get fresh calculation
	usage, _, err := calc.CachedUsage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usage != 15 { // "updated content" = 15 bytes
		t.Errorf("expected 15 bytes after clear, got %d", usage)
	}
}
