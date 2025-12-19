//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupConfigWithHooks(t *testing.T, postCommand, preCommand string) {
	t.Helper()

	configDir := filepath.Join(testRoot, ".canopy")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	t.Setenv("HOME", testRoot)

	projectsRoot := filepath.Join(testRoot, "projects")
	workspacesRoot := filepath.Join(testRoot, "workspaces")

	if err := os.MkdirAll(projectsRoot, 0o750); err != nil {
		t.Fatalf("Failed to create projects root: %v", err)
	}

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("Failed to create workspaces root: %v", err)
	}

	repoAURL := createLocalRepo(t, "hook-repo-a")
	repoBURL := createLocalRepo(t, "hook-repo-b")

	configContent := fmt.Sprintf(`
projects_root: %q
workspaces_root: %q
hooks:
  post_create:
    - command: %q
  pre_close:
    - command: %q
defaults:
  workspace_patterns:
    - pattern: "^TEST-"
      repos: ["hook-repo-a", "hook-repo-b"]
`, projectsRoot, workspacesRoot, postCommand, preCommand)

	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	registryContent := fmt.Sprintf(`repos:
  hook-repo-a:
    url: %q
  hook-repo-b:
    url: %q
`, repoAURL, repoBURL)

	registryFile := filepath.Join(configDir, "repos.yaml")
	if err := os.WriteFile(registryFile, []byte(registryContent), 0o600); err != nil {
		t.Fatalf("Failed to write registry file: %v", err)
	}
}

func TestHookDryRunWorkspaceNew(t *testing.T) {
	postOutput := filepath.Join(testRoot, "hook-post.out")
	preOutput := filepath.Join(testRoot, "hook-pre.out")

	postCommand := fmt.Sprintf("echo {{.WorkspaceID}} {{.BranchName}} > %s", postOutput)
	preCommand := fmt.Sprintf("echo preclose > %s", preOutput)

	setupConfigWithHooks(t, postCommand, preCommand)

	out, err := runCanopy("workspace", "new", "TEST-HOOK-DRY", "--dry-run-hooks")
	if err != nil {
		t.Fatalf("Failed to create workspace with dry-run hooks: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Hook dry-run (post_create)") {
		t.Fatalf("Expected dry-run output, got: %s", out)
	}

	if !strings.Contains(out, fmt.Sprintf("echo TEST-HOOK-DRY TEST-HOOK-DRY > %s", postOutput)) {
		t.Fatalf("Expected resolved command in output, got: %s", out)
	}

	if _, err := os.Stat(postOutput); !os.IsNotExist(err) {
		t.Fatalf("Expected hook command not to execute, but file exists at %s", postOutput)
	}

	wsDir := filepath.Join(testRoot, "workspaces", "TEST-HOOK-DRY")
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		t.Fatalf("Workspace directory not created at %s", wsDir)
	}
}

func TestHookDryRunWorkspaceClose(t *testing.T) {
	postOutput := filepath.Join(testRoot, "hook-post-close.out")
	preOutput := filepath.Join(testRoot, "hook-pre-close.out")

	postCommand := fmt.Sprintf("echo post > %s", postOutput)
	preCommand := fmt.Sprintf("echo {{.WorkspaceID}} > %s", preOutput)

	setupConfigWithHooks(t, postCommand, preCommand)

	out, err := runCanopy("workspace", "new", "TEST-HOOK-CLOSE")
	if err != nil {
		t.Fatalf("Failed to create workspace for close test: %v\nOutput: %s", err, out)
	}

	out, err = runCanopy("workspace", "close", "TEST-HOOK-CLOSE", "--delete", "--dry-run-hooks")
	if err != nil {
		t.Fatalf("Failed to close workspace with dry-run hooks: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Hook dry-run (pre_close)") {
		t.Fatalf("Expected dry-run output, got: %s", out)
	}

	if !strings.Contains(out, fmt.Sprintf("echo TEST-HOOK-CLOSE > %s", preOutput)) {
		t.Fatalf("Expected resolved command in output, got: %s", out)
	}

	if _, err := os.Stat(preOutput); !os.IsNotExist(err) {
		t.Fatalf("Expected hook command not to execute, but file exists at %s", preOutput)
	}
}
