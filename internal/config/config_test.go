package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "yard-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `
projects_root: /tmp/projects
workspaces_root: /tmp/workspaces
workspace_naming: "{{.ID}}"
defaults:
  workspace_patterns:
    - pattern: "^TEST-"
      repos: ["test-repo"]
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set environment variable to point to temp config
	os.Setenv("HOME", tmpDir)
	// Note: config.Load() looks in ~/.config/yardmaster/config.yaml or ./config.yaml
	// We can mock the home directory or just put it in current directory?
	// The Load() function checks current directory first.
	// Let's try to write to ./config.yaml but we need to be careful not to overwrite existing one.
	// Better to modify Load() to accept path? Or just rely on precedence.
	// Since we are running tests, we can change working directory?

	// Let's try to create the directory structure in tmpDir
	configDir := filepath.Join(tmpDir, ".config", "yardmaster")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// We can't easily mock HOME in Go tests for os.UserHomeDir without external libs or modifying code.
	// But config.Load() checks "." first.
	// So let's run test in a temp dir.
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(tmpDir)

	// Write config.yaml to tmpDir (current dir)
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write local config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.ProjectsRoot != "/tmp/projects" {
		t.Errorf("expected ProjectsRoot /tmp/projects, got %s", cfg.ProjectsRoot)
	}
	if cfg.WorkspacesRoot != "/tmp/workspaces" {
		t.Errorf("expected WorkspacesRoot /tmp/workspaces, got %s", cfg.WorkspacesRoot)
	}
}

func TestGetReposForWorkspace(t *testing.T) {
	cfg := &Config{
		Defaults: Defaults{
			WorkspacePatterns: []WorkspacePattern{
				{Pattern: "^TEST-", Repos: []string{"repo-a", "repo-b"}},
				{Pattern: "^PROJ-", Repos: []string{"repo-c"}},
			},
		},
	}

	tests := []struct {
		id       string
		expected []string
	}{
		{"TEST-123", []string{"repo-a", "repo-b"}},
		{"PROJ-456", []string{"repo-c"}},
		{"OTHER-789", nil},
	}

	for _, tt := range tests {
		repos := cfg.GetReposForWorkspace(tt.id)
		if len(repos) != len(tt.expected) {
			t.Errorf("GetReposForWorkspace(%s) returned %d repos, expected %d", tt.id, len(repos), len(tt.expected))
		}
		// Check content if needed, but length check is good first step
	}
}
