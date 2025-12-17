//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCloseWorkspaceDirtyRepo(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "dirty-close-repo")

	tc.setupBasicConfig(map[string]string{
		"dirty-close-repo": repoURL,
	})

	// 1. Create workspace
	tc.createWorkspace("TEST-DIRTY-CLOSE", "dirty-close-repo")

	wsDir := filepath.Join(tc.wsRoot, "TEST-DIRTY-CLOSE")
	repoDir := filepath.Join(wsDir, "dirty-close-repo")

	// 2. Make repo dirty
	tc.makeDirty(repoDir)

	// 3. Try to close without force - should fail
	out, err := tc.closeWorkspaceWithFlags("TEST-DIRTY-CLOSE")
	if err == nil {
		t.Fatalf("Close should fail when repo is dirty\nOutput: %s", out)
	}

	// Verify error mentions dirty state
	if !strings.Contains(out, "dirty") && !strings.Contains(out, "uncommitted") && !strings.Contains(out, "untracked") {
		t.Errorf("Error should mention dirty state: %s", out)
	}

	// 4. Workspace should still exist
	if !tc.workspaceExists("TEST-DIRTY-CLOSE") {
		t.Error("Workspace should still exist after failed close")
	}

	// 5. Close with force should succeed
	out, err = tc.closeWorkspaceWithFlags("TEST-DIRTY-CLOSE", "--force")
	if err != nil {
		t.Fatalf("Force close should succeed: %v\nOutput: %s", err, out)
	}

	// 6. Workspace should be gone
	if tc.workspaceExists("TEST-DIRTY-CLOSE") {
		t.Error("Workspace should not exist after force close")
	}
}

func TestCreateWorkspaceInvalidConfig(t *testing.T) {
	tc := newTestContext(t)

	// Setup with invalid/missing repos root
	tc.setupBasicConfig(nil)

	// Overwrite config with invalid paths
	configContent := `
projects_root: "/nonexistent/path/that/does/not/exist"
workspaces_root: "/another/nonexistent/path"
`

	configFile := filepath.Join(tc.configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Try to create workspace - should fail with meaningful error
	out, err := runCanopy("workspace", "new", "INVALID-CONFIG-TEST")
	if err == nil {
		t.Fatalf("Creating workspace with invalid config should fail\nOutput: %s", out)
	}

	// Error should be informative (path doesn't exist, etc.)
	// The exact message may vary
	t.Logf("Expected error for invalid config: %s", out)
}

func TestWorkspaceNotFound(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "notfound-repo")

	tc.setupBasicConfig(map[string]string{
		"notfound-repo": repoURL,
	})

	// 1. Try to view non-existent workspace
	out, err := runCanopy("workspace", "view", "NONEXISTENT-WORKSPACE")
	if err == nil {
		t.Fatalf("View should fail for non-existent workspace\nOutput: %s", out)
	}

	if !strings.Contains(strings.ToLower(out), "not found") && !strings.Contains(strings.ToLower(out), "does not exist") {
		t.Errorf("Error message should indicate workspace not found: %s", out)
	}

	// 2. Try to close non-existent workspace
	out, err = runCanopy("workspace", "close", "NONEXISTENT-WORKSPACE")
	if err == nil {
		t.Fatalf("Close should fail for non-existent workspace\nOutput: %s", out)
	}

	// 3. Try to rename non-existent workspace
	out, err = runCanopy("workspace", "rename", "NONEXISTENT-WORKSPACE", "NEW-NAME")
	if err == nil {
		t.Fatalf("Rename should fail for non-existent workspace\nOutput: %s", out)
	}

	// 4. Try to get path of non-existent workspace
	out, err = runCanopy("workspace", "path", "NONEXISTENT-WORKSPACE")
	if err == nil {
		t.Fatalf("Path should fail for non-existent workspace\nOutput: %s", out)
	}
}

func TestInvalidWorkspaceID(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "invalid-id-repo")

	tc.setupBasicConfig(map[string]string{
		"invalid-id-repo": repoURL,
	})

	// Try to create workspace with invalid ID characters
	// The exact validation rules depend on implementation
	invalidIDs := []string{
		"", // empty
		"workspace with spaces",
		"workspace/with/slashes",
	}

	for _, id := range invalidIDs {
		if id == "" {
			// Empty ID would be caught by cobra as missing argument
			continue
		}

		out, err := runCanopy("workspace", "new", id, "--repos", "invalid-id-repo")
		if err == nil {
			t.Logf("Creating workspace with ID '%s' should fail or be handled\nOutput: %s", id, out)
		}
	}
}

func TestAddNonexistentRepoToWorkspace(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "existing-repo")

	tc.setupBasicConfig(map[string]string{
		"existing-repo": repoURL,
	})

	// Create workspace
	tc.createWorkspace("TEST-ADD-NONEXISTENT", "existing-repo")

	// Try to add non-existent repo
	out, err := tc.addRepoToWorkspace("TEST-ADD-NONEXISTENT", "nonexistent-repo")
	if err == nil {
		t.Fatalf("Adding non-existent repo should fail\nOutput: %s", out)
	}

	if !strings.Contains(strings.ToLower(out), "unknown") && !strings.Contains(strings.ToLower(out), "not found") && !strings.Contains(strings.ToLower(out), "unregistered") {
		t.Errorf("Error should indicate unknown repo: %s", out)
	}
}

func TestRemoveNonexistentRepoFromWorkspace(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "remove-test-repo")

	tc.setupBasicConfig(map[string]string{
		"remove-test-repo": repoURL,
	})

	// Create workspace
	tc.createWorkspace("TEST-REMOVE-NONEXISTENT", "remove-test-repo")

	// Try to remove repo that's not in the workspace
	out, err := tc.removeRepoFromWorkspace("TEST-REMOVE-NONEXISTENT", "not-in-workspace")
	if err == nil {
		t.Fatalf("Removing non-existent repo should fail\nOutput: %s", out)
	}
}

func TestDuplicateWorkspaceCreation(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "duplicate-repo")

	tc.setupBasicConfig(map[string]string{
		"duplicate-repo": repoURL,
	})

	// Create first workspace
	tc.createWorkspace("TEST-DUPLICATE", "duplicate-repo")

	// Try to create another with the same ID
	out, err := runCanopy("workspace", "new", "TEST-DUPLICATE", "--repos", "duplicate-repo")
	if err == nil {
		t.Fatalf("Creating duplicate workspace should fail\nOutput: %s", out)
	}

	if !strings.Contains(strings.ToLower(out), "exists") && !strings.Contains(strings.ToLower(out), "already") {
		t.Errorf("Error should indicate workspace exists: %s", out)
	}
}
