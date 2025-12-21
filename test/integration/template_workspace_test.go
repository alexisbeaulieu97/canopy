//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTemplateConfig(t *testing.T) (string, string, string) {
	t.Helper()

	configDir := filepath.Join(testRoot, ".canopy")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	t.Setenv("HOME", testRoot)

	projectsRoot := filepath.Join(testRoot, "projects")
	workspacesRoot := filepath.Join(testRoot, "workspaces")
	closedRoot := filepath.Join(testRoot, "closed")

	if err := os.MkdirAll(projectsRoot, 0o750); err != nil {
		t.Fatalf("Failed to create projects root: %v", err)
	}

	if err := os.MkdirAll(workspacesRoot, 0o750); err != nil {
		t.Fatalf("Failed to create workspaces root: %v", err)
	}

	if err := os.MkdirAll(closedRoot, 0o750); err != nil {
		t.Fatalf("Failed to create closed root: %v", err)
	}

	repoAURL := createLocalRepo(t, "template-repo-a")
	repoBURL := createLocalRepo(t, "template-repo-b")
	repoCURL := createLocalRepo(t, "template-repo-c")

	configContent := fmt.Sprintf(`projects_root: "%s"
workspaces_root: "%s"
closed_root: "%s"
workspace_naming: "{{.ID}}"

templates:
  backend:
    description: "Backend workspace defaults"
    repos: ["repo-a", "repo-b"]
    default_branch: "main"
  setup:
    repos: ["repo-a"]
    setup_commands:
      - "touch setup-ok.txt"
      - "false"
      - "touch setup-after.txt"
`, projectsRoot, workspacesRoot, closedRoot)

	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	registryContent := fmt.Sprintf(`repos:
  repo-a:
    url: "%s"
  repo-b:
    url: "%s"
  repo-c:
    url: "%s"
`, repoAURL, repoBURL, repoCURL)

	registryFile := filepath.Join(configDir, "repos.yaml")
	if err := os.WriteFile(registryFile, []byte(registryContent), 0o600); err != nil {
		t.Fatalf("Failed to write registry file: %v", err)
	}

	return workspacesRoot, "repo-a", "repo-b"
}

func TestWorkspaceCreateWithTemplate(t *testing.T) {
	workspacesRoot, _, _ := setupTemplateConfig(t)

	out, err := runCanopy("workspace", "new", "TEST-TEMPLATE", "--template", "backend")
	if err != nil {
		t.Fatalf("Failed to create workspace from template: %v\nOutput: %s", err, out)
	}

	wsDir := filepath.Join(workspacesRoot, "TEST-TEMPLATE")
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		t.Fatalf("Workspace directory not created at %s", wsDir)
	}

	for _, repo := range []string{"repo-a", "repo-b"} {
		repoDir := filepath.Join(wsDir, repo)
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			t.Fatalf("Expected repo directory %s to exist", repoDir)
		}
	}

	metaPath := filepath.Join(wsDir, "workspace.yaml")
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read workspace metadata: %v", err)
	}

	if !strings.Contains(string(metaBytes), "branch_name: main") {
		t.Fatalf("Expected default branch from template, got metadata: %s", string(metaBytes))
	}
}

func TestWorkspaceTemplateWithExtraRepos(t *testing.T) {
	workspacesRoot, _, _ := setupTemplateConfig(t)

	out, err := runCanopy(
		"workspace",
		"new",
		"TEST-TEMPLATE-EXTRA",
		"--template",
		"backend",
		"--repos",
		"repo-c",
		"--branch",
		"extra-branch",
	)
	if err != nil {
		t.Fatalf("Failed to create workspace from template with extras: %v\nOutput: %s", err, out)
	}

	wsDir := filepath.Join(workspacesRoot, "TEST-TEMPLATE-EXTRA")
	for _, repo := range []string{"repo-a", "repo-b", "repo-c"} {
		repoDir := filepath.Join(wsDir, repo)
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			t.Fatalf("Expected repo directory %s to exist", repoDir)
		}
	}
}

func TestWorkspaceTemplateSetupCommands(t *testing.T) {
	workspacesRoot, _, _ := setupTemplateConfig(t)

	out, err := runCanopy("workspace", "new", "TEST-TEMPLATE-SETUP", "--template", "setup")
	if err != nil {
		t.Fatalf("Failed to create workspace from template with setup commands: %v\nOutput: %s", err, out)
	}

	wsDir := filepath.Join(workspacesRoot, "TEST-TEMPLATE-SETUP")
	if _, err := os.Stat(filepath.Join(wsDir, "setup-ok.txt")); os.IsNotExist(err) {
		t.Fatalf("Expected setup-ok.txt to exist")
	}

	if _, err := os.Stat(filepath.Join(wsDir, "setup-after.txt")); os.IsNotExist(err) {
		t.Fatalf("Expected setup-after.txt to exist")
	}

	metaPath := filepath.Join(wsDir, "workspace.yaml")
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read workspace metadata: %v", err)
	}

	if !strings.Contains(string(metaBytes), "setup_incomplete: true") {
		t.Fatalf("Expected setup_incomplete in metadata, got: %s", string(metaBytes))
	}
}
