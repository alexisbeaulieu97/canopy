//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceRestoreFlow(t *testing.T) {
	tc := newTestContext(t)

	// Create repos
	repoAURL := createLocalRepo(t, "restore-repo-a")
	repoBURL := createLocalRepo(t, "restore-repo-b")

	tc.setupBasicConfig(map[string]string{
		"restore-repo-a": repoAURL,
		"restore-repo-b": repoBURL,
	})

	// 1. Create workspace
	tc.createWorkspace("TEST-RESTORE", "restore-repo-a", "restore-repo-b")

	// Verify workspace exists
	if !tc.workspaceExists("TEST-RESTORE") {
		t.Fatal("Workspace should exist after creation")
	}

	// 2. Close workspace with --keep to preserve metadata
	out, err := tc.closeWorkspaceWithFlags("TEST-RESTORE", "--keep")
	if err != nil {
		t.Fatalf("Failed to close workspace with --keep: %v\nOutput: %s", err, out)
	}

	// Verify workspace directory is gone
	if tc.workspaceExists("TEST-RESTORE") {
		t.Fatal("Workspace directory should be removed after close")
	}

	// 3. Reopen workspace
	out, err = tc.reopenWorkspace("TEST-RESTORE", false)
	if err != nil {
		t.Fatalf("Failed to reopen workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Restored workspace TEST-RESTORE") {
		t.Errorf("Unexpected reopen output: %s", out)
	}

	// 4. Verify workspace is fully restored
	if !tc.workspaceExists("TEST-RESTORE") {
		t.Fatal("Workspace should exist after reopen")
	}

	// Verify repos are present
	wsDir := filepath.Join(tc.wsRoot, "TEST-RESTORE")
	repoADir := filepath.Join(wsDir, "restore-repo-a")
	repoBDir := filepath.Join(wsDir, "restore-repo-b")

	if _, err := os.Stat(repoADir); os.IsNotExist(err) {
		t.Errorf("Repo A should exist after reopen")
	}

	if _, err := os.Stat(repoBDir); os.IsNotExist(err) {
		t.Errorf("Repo B should exist after reopen")
	}

	// Verify workspace can be viewed
	out, err = tc.getWorkspaceStatus("TEST-RESTORE")
	if err != nil {
		t.Fatalf("Failed to view restored workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "TEST-RESTORE") {
		t.Errorf("View output should contain workspace ID: %s", out)
	}
}

func TestWorkspaceRestoreForceOverwrite(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "force-repo")

	tc.setupBasicConfig(map[string]string{
		"force-repo": repoURL,
	})

	// 1. Create and close workspace with metadata preserved
	tc.createWorkspace("TEST-FORCE", "force-repo")

	_, err := tc.closeWorkspaceWithFlags("TEST-FORCE", "--keep")
	if err != nil {
		t.Fatalf("Failed to close workspace: %v", err)
	}

	// 2. Create a new workspace with same ID (should work because old one is closed)
	tc.createWorkspace("TEST-FORCE", "force-repo")

	// 3. Try to reopen without force - should fail
	out, err := tc.reopenWorkspace("TEST-FORCE", false)
	if err == nil {
		t.Fatalf("Reopen without force should fail when workspace exists\nOutput: %s", out)
	}

	// 4. Reopen with force - should succeed
	out, err = tc.reopenWorkspace("TEST-FORCE", true)
	if err != nil {
		t.Fatalf("Reopen with force should succeed: %v\nOutput: %s", err, out)
	}

	// 5. Verify workspace exists and is functional
	if !tc.workspaceExists("TEST-FORCE") {
		t.Fatal("Workspace should exist after force reopen")
	}
}

func TestWorkspaceRename(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "rename-repo")

	tc.setupBasicConfig(map[string]string{
		"rename-repo": repoURL,
	})

	// 1. Create workspace
	tc.createWorkspace("TEST-OLD", "rename-repo")

	// Verify it exists
	if !tc.workspaceExists("TEST-OLD") {
		t.Fatal("Workspace should exist")
	}

	// 2. Rename workspace
	out, err := tc.renameWorkspace("TEST-OLD", "TEST-NEW", false)
	if err != nil {
		t.Fatalf("Failed to rename workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Renamed workspace TEST-OLD to TEST-NEW") {
		t.Errorf("Unexpected rename output: %s", out)
	}

	// 3. Verify old workspace doesn't exist
	if tc.workspaceExists("TEST-OLD") {
		t.Error("Old workspace should not exist after rename")
	}

	// 4. Verify new workspace exists
	if !tc.workspaceExists("TEST-NEW") {
		t.Fatal("New workspace should exist after rename")
	}

	// 5. Verify workspace can be used
	out, err = tc.getWorkspaceStatus("TEST-NEW")
	if err != nil {
		t.Fatalf("Failed to view renamed workspace: %v\nOutput: %s", err, out)
	}
}

func TestWorkspaceRenameWithBranch(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "rename-branch-repo")

	tc.setupBasicConfig(map[string]string{
		"rename-branch-repo": repoURL,
	})

	// 1. Create workspace with branch matching workspace ID
	tc.createWorkspaceWithBranch("TEST-BRANCH-OLD", "TEST-BRANCH-OLD", "rename-branch-repo")

	wsDir := filepath.Join(tc.wsRoot, "TEST-BRANCH-OLD")
	repoDir := filepath.Join(wsDir, "rename-branch-repo")

	// Verify branch name matches workspace ID
	branch := tc.getCurrentBranch(repoDir)
	if branch != "TEST-BRANCH-OLD" {
		t.Errorf("Initial branch should be TEST-BRANCH-OLD, got %s", branch)
	}

	// 2. Rename workspace with --rename-branch
	out, err := tc.renameWorkspace("TEST-BRANCH-OLD", "TEST-BRANCH-NEW", true)
	if err != nil {
		t.Fatalf("Failed to rename workspace with branch: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "branches also renamed") {
		t.Errorf("Output should mention branch rename: %s", out)
	}

	// 3. Verify workspace renamed
	if !tc.workspaceExists("TEST-BRANCH-NEW") {
		t.Fatal("New workspace should exist")
	}

	// 4. Verify branch was renamed
	newRepoDir := filepath.Join(tc.wsRoot, "TEST-BRANCH-NEW", "rename-branch-repo")

	newBranch := tc.getCurrentBranch(newRepoDir)
	if newBranch != "TEST-BRANCH-NEW" {
		t.Errorf("Branch should be renamed to TEST-BRANCH-NEW, got %s", newBranch)
	}
}
func TestWorkspaceListWithStatus(t *testing.T) {
	tc := newTestContext(t)

	repoAURL := createLocalRepo(t, "list-status-repo-a")
	repoBURL := createLocalRepo(t, "list-status-repo-b")

	tc.setupBasicConfig(map[string]string{
		"list-status-repo-a": repoAURL,
		"list-status-repo-b": repoBURL,
	})

	// Create workspace
	tc.createWorkspace("TEST-LIST-STATUS", "list-status-repo-a", "list-status-repo-b")

	wsDir := filepath.Join(tc.wsRoot, "TEST-LIST-STATUS")
	repoADir := filepath.Join(wsDir, "list-status-repo-a")

	// Make repo-a dirty
	tc.makeDirty(repoADir)

	// 1. Test list without --status (should work)
	out, err := runCanopy("workspace", "list")
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "TEST-LIST-STATUS") {
		t.Errorf("List output should contain workspace ID: %s", out)
	}

	// 2. Test list with --status
	out, err = runCanopy("workspace", "list", "--status")
	if err != nil {
		t.Fatalf("Failed to list workspaces with status: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "TEST-LIST-STATUS") {
		t.Errorf("Status output should contain workspace ID: %s", out)
	}

	// Should show dirty indicator for repo-a
	if !strings.Contains(out, "dirty") {
		t.Errorf("Status output should show dirty indicator for repo-a: %s", out)
	}

	// repo-b should be clean
	if !strings.Contains(out, "clean") {
		t.Errorf("Status output should show clean indicator for repo-b: %s", out)
	}

	// 3. Test list with --status --json
	out, err = runCanopy("workspace", "list", "--status", "--json")
	if err != nil {
		t.Fatalf("Failed to list workspaces with status in JSON: %v\nOutput: %s", err, out)
	}

	// JSON output should contain repo_statuses
	if !strings.Contains(out, "repo_statuses") {
		t.Errorf("JSON output should contain repo_statuses: %s", out)
	}

	// JSON should contain IsDirty field
	if !strings.Contains(out, "IsDirty") {
		t.Errorf("JSON output should contain IsDirty field: %s", out)
	}
}