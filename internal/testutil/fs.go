package testutil

import (
	"os"
	"testing"
)

// MustMkdir creates a directory at the specified path, including any necessary
// parent directories. It fails the test immediately if the directory cannot be created.
func MustMkdir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// MustWriteFile writes data to a file at the specified path.
// It fails the test immediately if the file cannot be written.
func MustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil { //nolint:gosec // test helper
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// MustReadFile reads the entire contents of a file.
// It fails the test immediately if the file cannot be read.
func MustReadFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path) //nolint:gosec // test helper
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	return string(content)
}

// MustRemoveAll removes a path and any children it contains.
// It fails the test immediately if the removal fails.
func MustRemoveAll(t *testing.T, path string) {
	t.Helper()

	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("failed to remove %s: %v", path, err)
	}
}

// MustTempDir creates a temporary directory and returns its path.
// The directory is automatically cleaned up when the test finishes.
func MustTempDir(t *testing.T, pattern string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir) //nolint:errcheck,gosec // cleanup best effort
	})

	return dir
}
