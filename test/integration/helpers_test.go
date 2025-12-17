//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestContext holds common test fixtures for integration tests.
type TestContext struct {
	t            *testing.T
	testRoot     string
	projectsRoot string
	wsRoot       string
	configDir    string
}

// newTestContext creates a new TestContext with isolated directories.
// Cleanup is handled by TestMain which removes the shared testRoot directory.
func newTestContext(t *testing.T) *TestContext {
	t.Helper()

	tc := &TestContext{
		t:            t,
		testRoot:     testRoot, // from TestMain
		projectsRoot: filepath.Join(testRoot, "projects"),
		wsRoot:       filepath.Join(testRoot, "workspaces"),
		configDir:    filepath.Join(testRoot, ".canopy"),
	}

	t.Setenv("HOME", testRoot)

	return tc
}

// createWorkspace creates a workspace using the CLI and returns the workspace directory path.
func (tc *TestContext) createWorkspace(id string, repos ...string) string {
	tc.t.Helper()

	args := []string{"workspace", "new", id}
	if len(repos) > 0 {
		args = append(args, "--repos", strings.Join(repos, ","))
	}

	out, err := runCanopy(args...)
	if err != nil {
		tc.t.Fatalf("Failed to create workspace %s: %v\nOutput: %s", id, err, out)
	}

	return filepath.Join(tc.wsRoot, id)
}

// createWorkspaceWithBranch creates a workspace with an explicit branch name.
func (tc *TestContext) createWorkspaceWithBranch(id, branch string, repos ...string) string {
	tc.t.Helper()

	args := []string{"workspace", "new", id, "--branch", branch}
	if len(repos) > 0 {
		args = append(args, "--repos", strings.Join(repos, ","))
	}

	out, err := runCanopy(args...)
	if err != nil {
		tc.t.Fatalf("Failed to create workspace %s: %v\nOutput: %s", id, err, out)
	}

	return filepath.Join(tc.wsRoot, id)
}

// closeWorkspace closes a workspace and returns the output.
func (tc *TestContext) closeWorkspace(id string) string {
	tc.t.Helper()

	out, err := runCanopy("workspace", "close", id)
	if err != nil {
		tc.t.Fatalf("Failed to close workspace %s: %v\nOutput: %s", id, err, out)
	}

	return out
}

// closeWorkspaceWithFlags closes a workspace with additional flags.
func (tc *TestContext) closeWorkspaceWithFlags(id string, flags ...string) (string, error) {
	tc.t.Helper()

	args := []string{"workspace", "close", id}
	args = append(args, flags...)

	return runCanopy(args...)
}

// reopenWorkspace reopens a closed workspace.
func (tc *TestContext) reopenWorkspace(id string, force bool) (string, error) {
	tc.t.Helper()

	args := []string{"workspace", "reopen", id}
	if force {
		args = append(args, "--force")
	}

	return runCanopy(args...)
}

// renameWorkspace renames a workspace.
func (tc *TestContext) renameWorkspace(oldID, newID string, renameBranch bool) (string, error) {
	tc.t.Helper()

	args := []string{"workspace", "rename", oldID, newID}
	if renameBranch {
		args = append(args, "--rename-branch")
	}

	return runCanopy(args...)
}

// addRepoToWorkspace adds a repo to an existing workspace.
func (tc *TestContext) addRepoToWorkspace(workspaceID, repoName string) (string, error) {
	tc.t.Helper()

	return runCanopy("workspace", "repo", "add", workspaceID, repoName)
}

// removeRepoFromWorkspace removes a repo from a workspace.
func (tc *TestContext) removeRepoFromWorkspace(workspaceID, repoName string) (string, error) {
	tc.t.Helper()

	return runCanopy("workspace", "repo", "remove", workspaceID, repoName)
}

// getWorkspaceStatus gets the status of a workspace.
func (tc *TestContext) getWorkspaceStatus(id string) (string, error) {
	tc.t.Helper()

	return runCanopy("workspace", "view", id)
}

// switchBranch switches all repos in a workspace to a branch.
func (tc *TestContext) switchBranch(workspaceID, branch string, create bool) (string, error) {
	tc.t.Helper()

	args := []string{"workspace", "branch", workspaceID, branch}
	if create {
		args = append(args, "--create")
	}

	return runCanopy(args...)
}

// workspaceExists checks if a workspace directory exists.
func (tc *TestContext) workspaceExists(id string) bool {
	tc.t.Helper()

	wsDir := filepath.Join(tc.wsRoot, id)
	_, err := os.Stat(wsDir)

	return err == nil
}

// makeDirty creates a dirty state in a repo by modifying a file without committing.
func (tc *TestContext) makeDirty(repoPath string) {
	tc.t.Helper()

	testFile := filepath.Join(repoPath, "dirty.txt")
	if err := os.WriteFile(testFile, []byte("dirty content\n"), 0o600); err != nil {
		tc.t.Fatalf("failed to make repo dirty: %v", err)
	}
}

// gitInDir runs a git command in a directory.
func (tc *TestContext) gitInDir(dir string, args ...string) string {
	tc.t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		tc.t.Fatalf("git %v failed in %s: %v\n%s", args, dir, err, out)
	}

	return string(out)
}

// getCurrentBranch returns the current branch name for a repo.
func (tc *TestContext) getCurrentBranch(repoPath string) string {
	tc.t.Helper()

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath

	out, err := cmd.CombinedOutput()
	if err != nil {
		tc.t.Fatalf("failed to get current branch: %v\n%s", err, out)
	}

	// Trim newline
	branch := string(out)
	if len(branch) > 0 && branch[len(branch)-1] == '\n' {
		branch = branch[:len(branch)-1]
	}

	return branch
}

// branchExists checks if a branch exists in a repo.
func (tc *TestContext) branchExists(repoPath, branch string) bool {
	tc.t.Helper()

	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = repoPath

	err := cmd.Run()

	return err == nil
}

// setupBasicConfig creates a basic config with the given repos registered.
func (tc *TestContext) setupBasicConfig(repos map[string]string) {
	tc.t.Helper()

	if err := os.MkdirAll(tc.configDir, 0o750); err != nil {
		tc.t.Fatalf("Failed to create config dir: %v", err)
	}

	if err := os.MkdirAll(tc.projectsRoot, 0o750); err != nil {
		tc.t.Fatalf("Failed to create projects root: %v", err)
	}

	if err := os.MkdirAll(tc.wsRoot, 0o750); err != nil {
		tc.t.Fatalf("Failed to create workspaces root: %v", err)
	}

	configContent := `
projects_root: "` + tc.projectsRoot + `"
workspaces_root: "` + tc.wsRoot + `"
`

	configFile := filepath.Join(tc.configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		tc.t.Fatalf("Failed to write config file: %v", err)
	}

	if len(repos) > 0 {
		registryContent := "repos:\n"
		for alias, url := range repos {
			registryContent += "  " + alias + ":\n    url: \"" + url + "\"\n"
		}

		registryFile := filepath.Join(tc.configDir, "repos.yaml")
		if err := os.WriteFile(registryFile, []byte(registryContent), 0o600); err != nil {
			tc.t.Fatalf("Failed to write registry file: %v", err)
		}
	}
}
