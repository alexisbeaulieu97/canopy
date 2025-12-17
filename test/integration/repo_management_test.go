//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddRepoToExistingWorkspace(t *testing.T) {
	tc := newTestContext(t)

	// Create repos
	repoAURL := createLocalRepo(t, "add-repo-a")
	repoBURL := createLocalRepo(t, "add-repo-b")

	tc.setupBasicConfig(map[string]string{
		"add-repo-a": repoAURL,
		"add-repo-b": repoBURL,
	})

	// 1. Create workspace with only one repo
	tc.createWorkspace("TEST-ADD-REPO", "add-repo-a")

	wsDir := filepath.Join(tc.wsRoot, "TEST-ADD-REPO")

	// Verify only repo-a exists
	if _, err := os.Stat(filepath.Join(wsDir, "add-repo-a")); os.IsNotExist(err) {
		t.Fatal("Repo A should exist")
	}

	if _, err := os.Stat(filepath.Join(wsDir, "add-repo-b")); !os.IsNotExist(err) {
		t.Fatal("Repo B should not exist yet")
	}

	// 2. Add repo-b to workspace
	out, err := tc.addRepoToWorkspace("TEST-ADD-REPO", "add-repo-b")
	if err != nil {
		t.Fatalf("Failed to add repo: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Added repository add-repo-b") {
		t.Errorf("Unexpected output: %s", out)
	}

	// 3. Verify repo-b now exists
	if _, err := os.Stat(filepath.Join(wsDir, "add-repo-b")); os.IsNotExist(err) {
		t.Fatal("Repo B should exist after adding")
	}

	// 4. Verify workspace view shows both repos
	out, err = tc.getWorkspaceStatus("TEST-ADD-REPO")
	if err != nil {
		t.Fatalf("Failed to view workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "add-repo-a") || !strings.Contains(out, "add-repo-b") {
		t.Errorf("View should show both repos: %s", out)
	}
}

func TestRemoveRepoFromWorkspace(t *testing.T) {
	tc := newTestContext(t)

	// Create repos
	repoAURL := createLocalRepo(t, "remove-repo-a")
	repoBURL := createLocalRepo(t, "remove-repo-b")

	tc.setupBasicConfig(map[string]string{
		"remove-repo-a": repoAURL,
		"remove-repo-b": repoBURL,
	})

	// 1. Create workspace with both repos
	tc.createWorkspace("TEST-REMOVE-REPO", "remove-repo-a", "remove-repo-b")

	wsDir := filepath.Join(tc.wsRoot, "TEST-REMOVE-REPO")

	// Verify both repos exist
	if _, err := os.Stat(filepath.Join(wsDir, "remove-repo-a")); os.IsNotExist(err) {
		t.Fatal("Repo A should exist")
	}

	if _, err := os.Stat(filepath.Join(wsDir, "remove-repo-b")); os.IsNotExist(err) {
		t.Fatal("Repo B should exist")
	}

	// 2. Remove repo-b from workspace
	out, err := tc.removeRepoFromWorkspace("TEST-REMOVE-REPO", "remove-repo-b")
	if err != nil {
		t.Fatalf("Failed to remove repo: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Removed repository remove-repo-b") {
		t.Errorf("Unexpected output: %s", out)
	}

	// 3. Verify repo-b directory is gone
	if _, err := os.Stat(filepath.Join(wsDir, "remove-repo-b")); !os.IsNotExist(err) {
		t.Fatal("Repo B should be removed from workspace")
	}

	// 4. Verify repo-a still exists
	if _, err := os.Stat(filepath.Join(wsDir, "remove-repo-a")); os.IsNotExist(err) {
		t.Fatal("Repo A should still exist")
	}

	// 5. Verify workspace view shows only repo-a
	out, err = tc.getWorkspaceStatus("TEST-REMOVE-REPO")
	if err != nil {
		t.Fatalf("Failed to view workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "remove-repo-a") {
		t.Errorf("View should show repo A: %s", out)
	}

	if strings.Contains(out, "remove-repo-b") {
		t.Errorf("View should not show repo B: %s", out)
	}
}

func TestRepoStatusInWorkspace(t *testing.T) {
	tc := newTestContext(t)

	repoURL := createLocalRepo(t, "status-repo")

	tc.setupBasicConfig(map[string]string{
		"status-repo": repoURL,
	})

	// 1. Create workspace
	tc.createWorkspace("TEST-STATUS", "status-repo")

	wsDir := filepath.Join(tc.wsRoot, "TEST-STATUS")
	repoDir := filepath.Join(wsDir, "status-repo")

	// 2. View workspace - should show clean status
	out, err := tc.getWorkspaceStatus("TEST-STATUS")
	if err != nil {
		t.Fatalf("Failed to view workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Clean") {
		t.Errorf("Workspace should show clean status: %s", out)
	}

	// 3. Make repo dirty
	tc.makeDirty(repoDir)

	// 4. View workspace again - should show dirty status
	out, err = tc.getWorkspaceStatus("TEST-STATUS")
	if err != nil {
		t.Fatalf("Failed to view dirty workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Dirty") {
		t.Errorf("Workspace should show dirty status: %s", out)
	}
}
