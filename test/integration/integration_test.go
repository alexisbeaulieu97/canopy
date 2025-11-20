package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	yardBinary string
	testRoot   string
)

func TestMain(m *testing.M) {
	// Setup
	var err error
	testRoot, err = os.MkdirTemp("", "yard-integration-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(testRoot)

	// Build yard binary
	yardBinary = filepath.Join(testRoot, "yard")
	cmd := exec.Command("go", "build", "-o", yardBinary, "../../cmd/yard")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build yard: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Teardown
	os.Exit(code)
}

func runYard(args ...string) (string, error) {
	cmd := exec.Command(yardBinary, args...)
	// Set environment variables to point to test config/dirs
	cmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", testRoot)) // Mock HOME to use local config if needed, or explicit config flag if we had one.
	// Yard looks for config in ~/.yard/config.yaml or ./config.yaml.
	// Let's create a config.yaml in the testRoot and run yard from there?
	// Or better, set YARD_CONFIG env var if we supported it? We don't yet.
	// But we can set HOME to testRoot, so it looks in testRoot/.yard/config.yaml

	return runCommand(cmd)
}

func runCommand(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func setupConfig(t *testing.T) {
	configDir := filepath.Join(testRoot, ".yard")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	projectsRoot := filepath.Join(testRoot, "projects")
	workspacesRoot := filepath.Join(testRoot, "workspaces")
	os.MkdirAll(projectsRoot, 0755)
	os.MkdirAll(workspacesRoot, 0755)

	configContent := fmt.Sprintf(`
projects_root: "%s"
workspaces_root: "%s"
defaults:
  ticket_patterns:
    - pattern: "^TEST-"
      repos: ["repo-a", "repo-b"]
`, projectsRoot, workspacesRoot)

	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
}

func TestWorkspaceLifecycle(t *testing.T) {
	setupConfig(t)

	// 1. Create Workspace
	out, err := runYard("workspace", "new", "TEST-LIFECYCLE")
	if err != nil {
		t.Fatalf("Failed to create workspace: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "Created workspace TEST-LIFECYCLE") {
		t.Errorf("Unexpected output: %s", out)
	}

	// Verify directory exists
	wsDir := filepath.Join(testRoot, "workspaces", "TEST-LIFECYCLE")
	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		t.Errorf("Workspace directory not created at %s", wsDir)
	}

	// 2. List Workspaces
	out, err = runYard("workspace", "list")
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "TEST-LIFECYCLE") {
		t.Errorf("List output missing workspace:\n%s", out)
	}

	// 3. View Workspace
	out, err = runYard("workspace", "view", "TEST-LIFECYCLE")
	if err != nil {
		t.Fatalf("Failed to view workspace: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "Workspace: TEST-LIFECYCLE") {
		t.Errorf("View output incorrect:\n%s", out)
	}

	// 4. Close Workspace
	out, err = runYard("workspace", "close", "TEST-LIFECYCLE")
	if err != nil {
		t.Fatalf("Failed to close workspace: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "Closed workspace TEST-LIFECYCLE") {
		t.Errorf("Unexpected close output: %s", out)
	}

	// Verify directory gone
	if _, err := os.Stat(wsDir); !os.IsNotExist(err) {
		t.Errorf("Workspace directory still exists after close")
	}
}

func TestPathCommands(t *testing.T) {
	setupConfig(t)

	// Create a dummy repo in projects root to test repo path
	repoName := "dummy-repo"
	repoPath := filepath.Join(testRoot, "projects", repoName)
	os.MkdirAll(repoPath, 0755)

	// Create a workspace
	runYard("workspace", "new", "TEST-PATH")

	// Test Workspace Path
	out, err := runYard("workspace", "path", "TEST-PATH")
	if err != nil {
		t.Fatalf("Failed to get workspace path: %v\nOutput: %s", err, out)
	}
	expectedWsPath := filepath.Join(testRoot, "workspaces", "TEST-PATH")
	if strings.TrimSpace(out) != expectedWsPath {
		t.Errorf("Expected workspace path %s, got %s", expectedWsPath, out)
	}

	// Test Repo Path
	out, err = runYard("repo", "path", repoName)
	if err != nil {
		t.Fatalf("Failed to get repo path: %v\nOutput: %s", err, out)
	}
	if strings.TrimSpace(out) != repoPath {
		t.Errorf("Expected repo path %s, got %s", repoPath, out)
	}
}
