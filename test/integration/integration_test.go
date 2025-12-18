//go:build integration

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
	canopyBinary string
	testRoot     string
)

func TestMain(m *testing.M) {
	// Setup
	var err error

	testRoot, err = os.MkdirTemp("", "canopy-integration-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	// Force Go build cache into the test workspace so sandboxed environments
	// don't attempt to write to the user cache directory.
	goCache := filepath.Join(testRoot, ".gocache")
	if err := os.MkdirAll(goCache, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create go cache dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.Setenv("GOCACHE", goCache); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set GOCACHE: %v\n", err)
		os.Exit(1)
	}

	defer func() { _ = os.RemoveAll(testRoot) }()

	// Build canopy binary
	canopyBinary = filepath.Join(testRoot, "canopy")

	cmd := exec.Command("go", "build", "-o", canopyBinary, "../../cmd/canopy")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build canopy: %v\n%s\n", err, out)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Teardown
	os.Exit(code)
}

func runCanopy(args ...string) (string, error) {
	return runCanopyWithInput("", args...)
}

func runCanopyWithInput(input string, args ...string) (string, error) {
	cmd := exec.Command(canopyBinary, args...)
	// Set environment variables to point to test config/dirs
	cmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", testRoot)) // Mock HOME to use local config if needed, or explicit config flag if we had one.
	// Canopy looks for config in ~/.canopy/config.yaml or ./config.yaml.
	// Let's create a config.yaml in the testRoot and run canopy from there?
	// Or better, set CANOPY_CONFIG env var if we supported it? We don't yet.
	// But we can set HOME to testRoot, so it looks in testRoot/.canopy/config.yaml
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	return runCommand(cmd)
}

func runCommand(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func setupConfig(t *testing.T) {
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

	repoAURL := createLocalRepo(t, "repo-a")
	repoBURL := createLocalRepo(t, "repo-b")

	configContent := fmt.Sprintf(`
projects_root: "%s"
workspaces_root: "%s"
defaults:
  workspace_patterns:
    - pattern: "^TEST-"
      repos: ["repo-a", "repo-b"]
`, projectsRoot, workspacesRoot)

	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	registryContent := fmt.Sprintf(`repos:
  repo-a:
    url: "%s"
  repo-b:
    url: "%s"
`, repoAURL, repoBURL)

	registryFile := filepath.Join(configDir, "repos.yaml")
	if err := os.WriteFile(registryFile, []byte(registryContent), 0o600); err != nil {
		t.Fatalf("Failed to write registry file: %v", err)
	}
}

func TestWorkspaceLifecycle(t *testing.T) {
	setupConfig(t)

	// 1. Create Workspace
	out, err := runCanopy("workspace", "new", "TEST-LIFECYCLE")
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
	out, err = runCanopy("workspace", "list")
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "TEST-LIFECYCLE") {
		t.Errorf("List output missing workspace:\n%s", out)
	}

	// 3. View Workspace
	out, err = runCanopy("workspace", "view", "TEST-LIFECYCLE")
	if err != nil {
		t.Fatalf("Failed to view workspace: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "Workspace: TEST-LIFECYCLE") {
		t.Errorf("View output incorrect:\n%s", out)
	}

	// 4. Close Workspace
	out, err = runCanopy("workspace", "close", "TEST-LIFECYCLE")
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

func TestCloseFlagConflict(t *testing.T) {
	setupConfig(t)

	out, err := runCanopy("workspace", "close", "TEST-FLAGS", "--keep", "--delete")
	if err == nil {
		t.Fatalf("expected error when using conflicting flags")
	}

	if !strings.Contains(out, "cannot use --keep and --delete together") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestPathCommands(t *testing.T) {
	setupConfig(t)

	// Create a dummy repo in projects root to test repo path
	repoName := "dummy-repo"

	repoPath := filepath.Join(testRoot, "projects", repoName)
	if err := os.MkdirAll(repoPath, 0o750); err != nil {
		t.Fatalf("Failed to create repo path: %v", err)
	}

	// Create a workspace
	if _, err := runCanopy("workspace", "new", "TEST-PATH"); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Test Workspace Path
	out, err := runCanopy("workspace", "path", "TEST-PATH")
	if err != nil {
		t.Fatalf("Failed to get workspace path: %v\nOutput: %s", err, out)
	}

	expectedWsPath := filepath.Join(testRoot, "workspaces", "TEST-PATH")
	if strings.TrimSpace(out) != expectedWsPath {
		t.Errorf("Expected workspace path %s, got %s", expectedWsPath, out)
	}

	// Test Repo Path
	out, err = runCanopy("repo", "path", repoName)
	if err != nil {
		t.Fatalf("Failed to get repo path: %v\nOutput: %s", err, out)
	}

	if strings.TrimSpace(out) != repoPath {
		t.Errorf("Expected repo path %s, got %s", repoPath, out)
	}
}

func TestRegistryCommandsAndWorkspace(t *testing.T) {
	setupConfig(t)

	remoteURL := createLocalRepo(t, "backend")

	// Add repo with auto-registration and derived alias
	if out, err := runCanopy("repo", "add", remoteURL, "--alias", "backend"); err != nil {
		t.Fatalf("repo add failed: %v\nOutput: %s", err, out)
	}

	// Adding same repo without alias should trigger prompt; respond with custom alias
	remoteURL2 := createLocalRepo(t, "backend2")
	if out, err := runCanopyWithInput("backend-2\n", "repo", "add", remoteURL2, "--alias", "backend"); err != nil {
		t.Fatalf("repo add with prompt failed: %v\nOutput: %s", err, out)
	}

	out, err := runCanopy("repo", "list-registry")
	if err != nil {
		t.Fatalf("list-registry failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "backend") || !strings.Contains(out, "backend-2") {
		t.Fatalf("registry list missing aliases:\n%s", out)
	}

	// Workspace creation using registry alias should succeed
	if out, err := runCanopy("workspace", "new", "REG-1", "--repos", "backend"); err != nil {
		t.Fatalf("workspace new failed: %v\nOutput: %s", err, out)
	}

	wsDir := filepath.Join(testRoot, "workspaces", "REG-1")
	if _, err := os.Stat(wsDir); err != nil {
		t.Fatalf("workspace directory not created: %v", err)
	}
}

func createLocalRepo(t *testing.T, name string) string {
	t.Helper()

	bare := filepath.Join(testRoot, "remotes", name+".git")
	if _, err := os.Stat(bare); err == nil {
		return "file://" + bare
	}

	src := filepath.Join(testRoot, "sources", name)
	if err := os.MkdirAll(src, 0o750); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	cmd := exec.Command("git", "init")

	cmd.Dir = src
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	for _, cfg := range [][]string{
		{"user.email", "test@example.com"},
		{"user.name", "canopy-tests"},
		{"commit.gpgsign", "false"},
	} {
		cmd = exec.Command("git", "config", cfg[0], cfg[1])

		cmd.Dir = src
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git config %s failed: %v\n%s", cfg[0], err, out)
		}
	}

	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("# "+name+"\n"), 0o600); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}

	cmd = exec.Command("git", "add", ".")

	cmd.Dir = src
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	cmd = exec.Command("git", "commit", "-m", "init")

	cmd.Dir = src
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}

	if err := os.MkdirAll(filepath.Dir(bare), 0o750); err != nil {
		t.Fatalf("failed to create remotes dir: %v", err)
	}

	cmd = exec.Command("git", "clone", "--bare", src, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone --bare failed: %v\n%s", err, out)
	}

	return "file://" + bare
}

func TestRepoStatus(t *testing.T) {
	setupConfig(t)

	// 1. Initially there might be repos from other tests, so we just check if it runs
	out, err := runCanopy("repo", "status")
	if err != nil {
		t.Fatalf("repo status failed: %v\nOutput: %s", err, out)
	}

	// 2. Add a repo
	remoteURL := createLocalRepo(t, "status-test")
	if _, err := runCanopy("repo", "add", remoteURL, "--alias", "status-test"); err != nil {
		t.Fatalf("repo add failed: %v", err)
	}

	// 3. Check status
	out, err = runCanopy("repo", "status")
	if err != nil {
		t.Fatalf("repo status failed: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "status-test") {
		t.Errorf("repo status missing repo:\n%s", out)
	}

	// 4. Check specific status
	out, err = runCanopy("repo", "status", "status-test")
	if err != nil {
		t.Fatalf("repo status <NAME> failed: %v", err)
	}
	if !strings.Contains(out, "Repository:    status-test") {
		t.Errorf("repo status <NAME> output incorrect:\n%s", out)
	}

	// 5. JSON output
	out, err = runCanopy("repo", "status", "--json")
	if err != nil {
		t.Fatalf("repo status --json failed: %v", err)
	}
	if !strings.Contains(out, `"name": "status-test"`) {
		t.Errorf("JSON output incorrect:\n%s", out)
	}
}
