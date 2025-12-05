package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "canopy-config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	configContent := `
projects_root: /tmp/projects
workspaces_root: /tmp/workspaces
closed_root: /tmp/closed
workspace_naming: "{{.ID}}"
workspace_close_default: archive
defaults:
  workspace_patterns:
    - pattern: "^TEST-"
      repos: ["test-repo"]
`

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set environment variable to point to temp config
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	// Note: config.Load() looks in ~/.canopy/config.yaml, ~/.config/canopy/config.yaml, or ./config.yaml
	// We can mock the home directory or just put it in current directory?
	// The Load() function checks current directory first.
	// Let's try to write to ./config.yaml but we need to be careful not to overwrite existing one.
	// Better to modify Load() to accept path? Or just rely on precedence.
	// Since we are running tests, we can change working directory?

	// Let's try to create the directory structure in tmpDir
	configDir := filepath.Join(tmpDir, ".config", "canopy")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// We can't easily mock HOME in Go tests for os.UserHomeDir without external libs or modifying code.
	// But config.Load() checks "." first.
	// So let's run test in a temp dir.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}

	defer func() { _ = os.Chdir(wd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Write config.yaml to tmpDir (current dir)
	if err := os.WriteFile("config.yaml", []byte(configContent), 0o644); err != nil {
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

	if cfg.ClosedRoot != "/tmp/closed" {
		t.Errorf("expected ClosedRoot /tmp/closed, got %s", cfg.ClosedRoot)
	}

	if cfg.CloseDefault != "archive" {
		t.Errorf("expected CloseDefault archive, got %s", cfg.CloseDefault)
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

func TestValidateValues(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid config",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
			},
			wantErr: false,
		},
		{
			name: "empty projects_root",
			cfg: &Config{
				ProjectsRoot:       "",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
			},
			wantErr:   true,
			errSubstr: "projects_root is required",
		},
		{
			name: "empty workspaces_root",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
			},
			wantErr:   true,
			errSubstr: "workspaces_root is required",
		},
		{
			name: "empty closed_root",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
			},
			wantErr:   true,
			errSubstr: "closed_root is required",
		},
		{
			name: "invalid close_default",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "invalid",
				StaleThresholdDays: 14,
			},
			wantErr:   true,
			errSubstr: "workspace_close_default must be either",
		},
		{
			name: "empty close_default defaults to delete",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "",
				StaleThresholdDays: 14,
			},
			wantErr: false,
		},
		{
			name: "archive close_default is valid",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "archive",
				StaleThresholdDays: 14,
			},
			wantErr: false,
		},
		{
			name: "negative stale_threshold_days",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: -1,
			},
			wantErr:   true,
			errSubstr: "stale_threshold_days must be zero or positive",
		},
		{
			name: "zero stale_threshold_days is valid",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid regex pattern",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Defaults: Defaults{
					WorkspacePatterns: []WorkspacePattern{
						{Pattern: "[invalid", Repos: []string{"repo"}},
					},
				},
			},
			wantErr:   true,
			errSubstr: "invalid regex pattern",
		},
		{
			name: "valid regex pattern",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Defaults: Defaults{
					WorkspacePatterns: []WorkspacePattern{
						{Pattern: "^TEST-.*", Repos: []string{"repo"}},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateValues()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateValues() expected error containing %q, got nil", tt.errSubstr)

					return
				}

				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateValues() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("ValidateValues() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "canopy-env-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	// Create subdirectories
	projectsDir := filepath.Join(tmpDir, "projects")
	workspacesDir := filepath.Join(tmpDir, "workspaces")
	closedDir := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("failed to create projects dir: %v", err)
	}

	if err := os.MkdirAll(workspacesDir, 0o755); err != nil {
		t.Fatalf("failed to create workspaces dir: %v", err)
	}

	if err := os.MkdirAll(closedDir, 0o755); err != nil {
		t.Fatalf("failed to create closed dir: %v", err)
	}

	// Create a file (not a directory) for testing
	filePath := filepath.Join(tmpDir, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		cfg       *Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "all paths exist and are directories",
			cfg: &Config{
				ProjectsRoot:   projectsDir,
				WorkspacesRoot: workspacesDir,
				ClosedRoot:     closedDir,
			},
			wantErr: false,
		},
		{
			name: "non-existent absolute path is valid (will be created)",
			cfg: &Config{
				ProjectsRoot:   filepath.Join(tmpDir, "nonexistent"),
				WorkspacesRoot: workspacesDir,
				ClosedRoot:     closedDir,
			},
			wantErr: false,
		},
		{
			name: "relative path that doesn't exist",
			cfg: &Config{
				ProjectsRoot:   "relative/path",
				WorkspacesRoot: workspacesDir,
				ClosedRoot:     closedDir,
			},
			wantErr:   true,
			errSubstr: "must be an absolute path",
		},
		{
			name: "path is a file, not a directory",
			cfg: &Config{
				ProjectsRoot:   filePath,
				WorkspacesRoot: workspacesDir,
				ClosedRoot:     closedDir,
			},
			wantErr:   true,
			errSubstr: "must be a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateEnvironment()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateEnvironment() expected error containing %q, got nil", tt.errSubstr)

					return
				}

				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateEnvironment() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("ValidateEnvironment() unexpected error: %v", err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "canopy-validate-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	// Create subdirectories
	projectsDir := filepath.Join(tmpDir, "projects")
	workspacesDir := filepath.Join(tmpDir, "workspaces")
	closedDir := filepath.Join(tmpDir, "closed")

	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("failed to create projects dir: %v", err)
	}

	if err := os.MkdirAll(workspacesDir, 0o755); err != nil {
		t.Fatalf("failed to create workspaces dir: %v", err)
	}

	if err := os.MkdirAll(closedDir, 0o755); err != nil {
		t.Fatalf("failed to create closed dir: %v", err)
	}

	tests := []struct {
		name      string
		cfg       *Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "fully valid config",
			cfg: &Config{
				ProjectsRoot:       projectsDir,
				WorkspacesRoot:     workspacesDir,
				ClosedRoot:         closedDir,
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
			},
			wantErr: false,
		},
		{
			name: "invalid values fail before environment checks",
			cfg: &Config{
				ProjectsRoot:       "",
				WorkspacesRoot:     workspacesDir,
				ClosedRoot:         closedDir,
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
			},
			wantErr:   true,
			errSubstr: "projects_root is required",
		},
		{
			name: "value errors take precedence over environment errors",
			cfg: &Config{
				ProjectsRoot:       "relative/path",
				WorkspacesRoot:     workspacesDir,
				ClosedRoot:         closedDir,
				CloseDefault:       "invalid",
				StaleThresholdDays: 14,
			},
			wantErr:   true,
			errSubstr: "workspace_close_default must be either",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errSubstr)

					return
				}

				if tt.errSubstr != "" && !contains(err.Error(), tt.errSubstr) {
					t.Errorf("Validate() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

// contains checks if substr is in s
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
