//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceNamingTemplate(t *testing.T) {
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

	repoURL := createLocalRepo(t, "naming-repo")

	configContent := fmt.Sprintf(`projects_root: "%s"
workspaces_root: "%s"
workspace_naming: "ws-{{.ID}}"
`, projectsRoot, workspacesRoot)

	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	registryContent := fmt.Sprintf(`repos:
  naming-repo:
    url: "%s"
`, repoURL)

	registryFile := filepath.Join(configDir, "repos.yaml")
	if err := os.WriteFile(registryFile, []byte(registryContent), 0o600); err != nil {
		t.Fatalf("Failed to write registry file: %v", err)
	}

	out, err := runCanopy("workspace", "new", "TEST-NAMING", "--repos", "naming-repo")
	if err != nil {
		t.Fatalf("Failed to create workspace: %v\nOutput: %s", err, out)
	}

	expectedDir := filepath.Join(workspacesRoot, "ws-TEST-NAMING")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Fatalf("Workspace directory not created at %s", expectedDir)
	}
}
