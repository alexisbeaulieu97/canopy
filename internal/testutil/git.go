package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// CreateRepoWithCommit initializes a git repository at the given path with an
// initial commit. It configures the user email and name, and creates a README.md
// file as the initial content.
//
// This is useful for tests that need a valid git repository with history.
func CreateRepoWithCommit(t *testing.T, path string) {
	t.Helper()

	MustMkdir(t, path)
	RunGit(t, path, "init")
	RunGit(t, path, "config", "user.email", "test@example.com")
	RunGit(t, path, "config", "user.name", "Test User")
	RunGit(t, path, "config", "credential.helper", "")

	filePath := filepath.Join(path, "README.md")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	RunGit(t, path, "add", ".")
	RunGit(t, path, "commit", "-m", "init")
}

// RunGit executes a git command in the specified directory.
// It fails the test immediately if the command fails.
//
// The function sets GIT_CONFIG_GLOBAL and GIT_CONFIG_SYSTEM to /dev/null
// to ensure consistent behavior regardless of user git configuration.
func RunGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %s (%v)", args, strings.TrimSpace(string(output)), err)
	}
}

// RunGitOutput executes a git command and returns its output.
// It fails the test immediately if the command fails.
//
// The returned string is trimmed of leading and trailing whitespace.
func RunGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s (%v)", args, strings.TrimSpace(string(output)), err)
	}

	return strings.TrimSpace(string(output))
}

// CloneToBare clones a git repository as a bare repository.
// This is useful for creating canonical repositories in tests.
func CloneToBare(t *testing.T, sourceRepo, destPath string) {
	t.Helper()

	RunGit(t, filepath.Dir(destPath), "clone", "--bare", sourceRepo, destPath)
}

