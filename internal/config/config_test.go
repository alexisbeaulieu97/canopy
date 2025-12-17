package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
)

// validGitConfig returns a GitConfig with valid default values for testing.
func validGitConfig() GitConfig {
	return GitConfig{
		Retry: GitRetrySettings{
			MaxAttempts:  3,
			InitialDelay: "1s",
			MaxDelay:     "30s",
			Multiplier:   2.0,
			JitterFactor: 0.25,
		},
	}
}

func TestLoad(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	// Clear CANOPY_CONFIG to ensure test uses fixture from current directory
	t.Setenv("CANOPY_CONFIG", "")

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

	cfg, err := Load("")
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

	if cfg.CloseDefault != CloseDefaultArchive {
		t.Errorf("expected CloseDefault %s, got %s", CloseDefaultArchive, cfg.CloseDefault)
	}
}

func TestLoadWithConfigPath(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	tmpDir, err := os.MkdirTemp("", "canopy-config-path-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	configContent := `
projects_root: /custom/projects
workspaces_root: /custom/workspaces
closed_root: /custom/closed
workspace_close_default: archive
`

	configPath := filepath.Join(tmpDir, "custom-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load(configPath) failed: %v", err)
	}

	if cfg.ProjectsRoot != "/custom/projects" {
		t.Errorf("expected ProjectsRoot /custom/projects, got %s", cfg.ProjectsRoot)
	}

	if cfg.WorkspacesRoot != "/custom/workspaces" {
		t.Errorf("expected WorkspacesRoot /custom/workspaces, got %s", cfg.WorkspacesRoot)
	}

	if cfg.ClosedRoot != "/custom/closed" {
		t.Errorf("expected ClosedRoot /custom/closed, got %s", cfg.ClosedRoot)
	}
}

func TestLoadWithEnvVar(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	tmpDir, err := os.MkdirTemp("", "canopy-config-env-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	configContent := `
projects_root: /env/projects
workspaces_root: /env/workspaces
closed_root: /env/closed
workspace_close_default: delete
`

	configPath := filepath.Join(tmpDir, "env-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set environment variable
	t.Setenv("CANOPY_CONFIG", configPath)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") with CANOPY_CONFIG failed: %v", err)
	}

	if cfg.ProjectsRoot != "/env/projects" {
		t.Errorf("expected ProjectsRoot /env/projects, got %s", cfg.ProjectsRoot)
	}

	if cfg.WorkspacesRoot != "/env/workspaces" {
		t.Errorf("expected WorkspacesRoot /env/workspaces, got %s", cfg.WorkspacesRoot)
	}
}

func TestLoadConfigPriority(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	tmpDir, err := os.MkdirTemp("", "canopy-config-priority-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	// Config for environment variable
	envConfigContent := `
projects_root: /env/projects
workspaces_root: /env/workspaces
closed_root: /env/closed
workspace_close_default: delete
`

	envConfigPath := filepath.Join(tmpDir, "env-config.yaml")
	if err := os.WriteFile(envConfigPath, []byte(envConfigContent), 0o644); err != nil {
		t.Fatalf("failed to write env config file: %v", err)
	}

	// Config for flag override
	flagConfigContent := `
projects_root: /flag/projects
workspaces_root: /flag/workspaces
closed_root: /flag/closed
workspace_close_default: archive
`

	flagConfigPath := filepath.Join(tmpDir, "flag-config.yaml")
	if err := os.WriteFile(flagConfigPath, []byte(flagConfigContent), 0o644); err != nil {
		t.Fatalf("failed to write flag config file: %v", err)
	}

	// Set environment variable
	t.Setenv("CANOPY_CONFIG", envConfigPath)

	// Load with explicit path - should take precedence over env var
	cfg, err := Load(flagConfigPath)
	if err != nil {
		t.Fatalf("Load(flagConfigPath) failed: %v", err)
	}

	if cfg.ProjectsRoot != "/flag/projects" {
		t.Errorf("expected flag path to take precedence, got ProjectsRoot %s", cfg.ProjectsRoot)
	}

	if cfg.CloseDefault != CloseDefaultArchive {
		t.Errorf("expected flag config close_default %s, got %s", CloseDefaultArchive, cfg.CloseDefault)
	}
}

func TestLoadWithMissingExplicitConfigPath(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	// Clear CANOPY_CONFIG to avoid interference
	t.Setenv("CANOPY_CONFIG", "")

	// Try to load with a non-existent explicit config path
	_, err := Load("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Fatal("Load() with non-existent explicit path should return an error")
	}

	// Should fail with an IO error about the file not existing
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("expected error about file not existing, got: %v", err)
	}
}

func TestLoadWithMissingEnvConfig(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	// Set CANOPY_CONFIG to a non-existent path
	t.Setenv("CANOPY_CONFIG", "/nonexistent/env/config.yaml")

	// Try to load - should fail because explicit env path is missing
	_, err := Load("")
	if err == nil {
		t.Fatal("Load() with non-existent CANOPY_CONFIG should return an error")
	}

	// Should fail with an IO error about the file not existing
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("expected error about file not existing, got: %v", err)
	}
}

func TestLoadWithTildeExpansion(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	// Clear CANOPY_CONFIG to avoid interference
	t.Setenv("CANOPY_CONFIG", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Create a temp config directory - use system temp but create a structure
	// that we can reference with tilde by symlinking
	tmpDir, err := os.MkdirTemp("", "canopy-tilde-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	configContent := `
projects_root: /tilde/projects
workspaces_root: /tilde/workspaces
closed_root: /tilde/closed
`

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Test expandPath directly since we can't easily create files in home dir
	// The Load function uses expandPath, so testing it validates the behavior
	expandedPath := expandPath("~/test/config.yaml", home)

	expectedPath := filepath.Join(home, "test/config.yaml")
	if expandedPath != expectedPath {
		t.Errorf("expandPath(\"~/test/config.yaml\", home) = %q, want %q", expandedPath, expectedPath)
	}

	// Test just "~"
	expandedHome := expandPath("~", home)
	if expandedHome != home {
		t.Errorf("expandPath(\"~\", home) = %q, want %q", expandedHome, home)
	}

	// Test non-tilde path stays unchanged
	regularPath := expandPath("/absolute/path", home)
	if regularPath != "/absolute/path" {
		t.Errorf("expandPath(\"/absolute/path\", home) = %q, want \"/absolute/path\"", regularPath)
	}

	// Now test actual Load with the temp config using absolute path
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() with absolute path failed: %v", err)
	}

	if cfg.ProjectsRoot != "/tilde/projects" {
		t.Errorf("expected ProjectsRoot /tilde/projects, got %s", cfg.ProjectsRoot)
	}
}

func TestLoadWithTildeExpansionEnvVar(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Create a temp config directory
	tmpDir, err := os.MkdirTemp("", "canopy-tilde-env-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	configContent := `
projects_root: /tilde-env/projects
workspaces_root: /tilde-env/workspaces
closed_root: /tilde-env/closed
`

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Test that expandPath works correctly for env var paths
	expandedPath := expandPath("~/.canopy/config.yaml", home)

	expectedPath := filepath.Join(home, ".canopy/config.yaml")
	if expandedPath != expectedPath {
		t.Errorf("expandPath(\"~/.canopy/config.yaml\", home) = %q, want %q", expandedPath, expectedPath)
	}

	// Test actual Load with absolute path through CANOPY_CONFIG
	t.Setenv("CANOPY_CONFIG", configPath)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() with CANOPY_CONFIG failed: %v", err)
	}

	if cfg.ProjectsRoot != "/tilde-env/projects" {
		t.Errorf("expected ProjectsRoot /tilde-env/projects, got %s", cfg.ProjectsRoot)
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
		errType   error // optional: use errors.Is() to check error type
	}{
		{
			name: "valid config",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
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
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr:   true,
			errSubstr: "projects_root",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "empty workspaces_root",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr:   true,
			errSubstr: "workspaces_root",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "empty closed_root",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr:   true,
			errSubstr: "closed_root",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "invalid close_default",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "invalid",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr:   true,
			errSubstr: "workspace_close_default",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "empty close_default defaults to delete",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr: false,
		},
		{
			name: "archive close_default is valid",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       CloseDefaultArchive,
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
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
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr:   true,
			errSubstr: "stale_threshold_days",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "zero stale_threshold_days is valid",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 0,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
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
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
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
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
				Defaults: Defaults{
					WorkspacePatterns: []WorkspacePattern{
						{Pattern: "^TEST-.*", Repos: []string{"repo"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid hooks configuration",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
				Hooks: Hooks{
					PostCreate: []Hook{
						{Command: "npm install", Repos: []string{"frontend"}},
					},
					PreClose: []Hook{
						{Command: "git stash"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "post_create hook with empty command",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
				Hooks: Hooks{
					PostCreate: []Hook{
						{Command: ""},
					},
				},
			},
			wantErr:   true,
			errSubstr: "post_create hook[0]",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "pre_close hook with empty command",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
				Hooks: Hooks{
					PreClose: []Hook{
						{Command: "echo first"},
						{Command: ""},
					},
				},
			},
			wantErr:   true,
			errSubstr: "pre_close hook[1]",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "hook command with only whitespace",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
				Hooks: Hooks{
					PostCreate: []Hook{
						{Command: "   \t"},
					},
				},
			},
			wantErr:   true,
			errSubstr: "post_create hook[0]",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "hook command with newline",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
				Hooks: Hooks{
					PreClose: []Hook{
						{Command: "echo first\nrm -rf /"},
					},
				},
			},
			wantErr:   true,
			errSubstr: "pre_close hook[0]",
			errType:   cerrors.ConfigValidation,
		},
		{
			name: "hook with negative timeout",
			cfg: &Config{
				ProjectsRoot:       "/tmp/projects",
				WorkspacesRoot:     "/tmp/workspaces",
				ClosedRoot:         "/tmp/closed",
				CloseDefault:       "delete",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
				Hooks: Hooks{
					PostCreate: []Hook{
						{Command: "npm install", Timeout: -5},
					},
				},
			},
			wantErr:   true,
			errSubstr: "post_create hook[0]",
			errType:   cerrors.ConfigValidation,
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

				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("ValidateValues() error does not match expected sentinel: got %v, want %v", err, tt.errType)
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
			errSubstr: "not a directory",
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
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
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
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr:   true,
			errSubstr: "projects_root",
		},
		{
			name: "value errors take precedence over environment errors",
			cfg: &Config{
				ProjectsRoot:       "relative/path",
				WorkspacesRoot:     workspacesDir,
				ClosedRoot:         closedDir,
				CloseDefault:       "invalid",
				StaleThresholdDays: 14,
				Git:                validGitConfig(),
				ParallelWorkers:    DefaultParallelWorkers,
			},
			wantErr:   true,
			errSubstr: "workspace_close_default",
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

func TestKeybindingsWithDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    Keybindings
		wantQuit []string
	}{
		{
			name:     "empty keybindings gets defaults",
			input:    Keybindings{},
			wantQuit: DefaultQuitKeys,
		},
		{
			name:     "custom keybindings preserved",
			input:    Keybindings{Quit: []string{"x"}},
			wantQuit: []string{"x"},
		},
		{
			name:     "partial custom keybindings preserved with defaults for others",
			input:    Keybindings{Quit: []string{"x"}, Search: []string{}},
			wantQuit: []string{"x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.WithDefaults()
			if len(result.Quit) != len(tt.wantQuit) {
				t.Errorf("WithDefaults().Quit = %v, want %v", result.Quit, tt.wantQuit)
			}

			for i, k := range result.Quit {
				if k != tt.wantQuit[i] {
					t.Errorf("WithDefaults().Quit[%d] = %q, want %q", i, k, tt.wantQuit[i])
				}
			}

			// Verify other defaults are applied when empty
			if len(tt.input.Search) == 0 && len(result.Search) != len(DefaultSearchKeys) {
				t.Errorf("WithDefaults().Search = %v, want %v", result.Search, DefaultSearchKeys)
			}
		})
	}
}

func TestKeybindingsValidation(t *testing.T) {
	tests := []struct {
		name      string
		kb        Keybindings
		wantErr   bool
		errSubstr string
	}{
		{
			name: "no conflicts with defaults",
			kb:   Keybindings{}.WithDefaults(),
		},
		{
			name: "no conflicts with custom keybindings",
			kb: Keybindings{
				Quit:        []string{"x"},
				Search:      []string{"/"},
				Push:        []string{"p"},
				Close:       []string{"c"},
				OpenEditor:  []string{"o"},
				ToggleStale: []string{"s"},
				Details:     []string{"enter"},
				Confirm:     []string{"y"},
				Cancel:      []string{"n"},
			},
		},
		{
			name: "conflict detected between quit and push",
			kb: Keybindings{
				Quit: []string{"p"},
				Push: []string{"p"},
			}.WithDefaults(),
			wantErr:   true,
			errSubstr: "key \"p\" is assigned to multiple actions",
		},
		{
			name: "conflict detected with multiple keys",
			kb: Keybindings{
				Quit:   []string{"q", "x"},
				Search: []string{"x", "/"},
			}.WithDefaults(),
			wantErr:   true,
			errSubstr: "key \"x\" is assigned to multiple actions",
		},
		{
			name: "multiple keys per action without conflict",
			kb: Keybindings{
				OpenEditor: []string{"o", "e"},
				Quit:       []string{"q", "ctrl+c"},
			}.WithDefaults(),
		},
		{
			name: "invalid empty key rejected",
			kb: Keybindings{
				Quit: []string{""},
			}.WithDefaults(),
			wantErr:   true,
			errSubstr: "invalid key \"\" for action \"quit\"",
		},
		{
			name: "invalid key name rejected",
			kb: Keybindings{
				Quit: []string{"foo"},
			}.WithDefaults(),
			wantErr:   true,
			errSubstr: "invalid key \"foo\" for action \"quit\"",
		},
		{
			name: "valid modifier keys accepted",
			kb: Keybindings{
				Quit: []string{"ctrl+c", "alt+q", "shift+x"},
			}.WithDefaults(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.kb.ValidateKeybindings()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateKeybindings() expected error, got nil")
					return
				}

				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateKeybindings() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("ValidateKeybindings() unexpected error: %v", err)
			}
		})
	}
}

func TestConfigValidateKeybindings(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid config with default keybindings",
			cfg: &Config{
				ProjectsRoot:    "/tmp/projects",
				WorkspacesRoot:  "/tmp/workspaces",
				ClosedRoot:      "/tmp/closed",
				CloseDefault:    "delete",
				Git:             validGitConfig(),
				ParallelWorkers: DefaultParallelWorkers,
			},
		},
		{
			name: "valid config with custom keybindings",
			cfg: &Config{
				ProjectsRoot:    "/tmp/projects",
				WorkspacesRoot:  "/tmp/workspaces",
				ClosedRoot:      "/tmp/closed",
				CloseDefault:    "delete",
				Git:             validGitConfig(),
				ParallelWorkers: DefaultParallelWorkers,
				TUI: TUIConfig{
					Keybindings: Keybindings{
						Quit:   []string{"x"},
						Search: []string{"f"},
					},
				},
			},
		},
		{
			name: "invalid config with keybinding conflicts",
			cfg: &Config{
				ProjectsRoot:    "/tmp/projects",
				WorkspacesRoot:  "/tmp/workspaces",
				ClosedRoot:      "/tmp/closed",
				CloseDefault:    "delete",
				Git:             validGitConfig(),
				ParallelWorkers: DefaultParallelWorkers,
				TUI: TUIConfig{
					Keybindings: Keybindings{
						Quit:   []string{"p"},
						Push:   []string{"p"},
						Search: []string{"/"},
					},
				},
			},
			wantErr:   true,
			errSubstr: "tui.keybindings",
		},
		{
			name: "invalid config with bad key name",
			cfg: &Config{
				ProjectsRoot:    "/tmp/projects",
				WorkspacesRoot:  "/tmp/workspaces",
				ClosedRoot:      "/tmp/closed",
				CloseDefault:    "delete",
				Git:             validGitConfig(),
				ParallelWorkers: DefaultParallelWorkers,
				TUI: TUIConfig{
					Keybindings: Keybindings{
						Quit: []string{"invalid-key"},
					},
				},
			},
			wantErr:   true,
			errSubstr: "invalid key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateValues()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateValues() expected error, got nil")
					return
				}

				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateValues() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("ValidateValues() unexpected error: %v", err)
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

func TestValidateGitRetry(t *testing.T) {
	baseConfig := func() *Config {
		return &Config{
			ProjectsRoot:       "/projects",
			WorkspacesRoot:     "/workspaces",
			ClosedRoot:         "/closed",
			CloseDefault:       "delete",
			StaleThresholdDays: 14,
			Git:                validGitConfig(),
			ParallelWorkers:    DefaultParallelWorkers,
		}
	}

	tests := []struct {
		name      string
		modify    func(c *Config)
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid default config",
			modify:  func(_ *Config) {},
			wantErr: false,
		},
		{
			name: "max_attempts zero",
			modify: func(c *Config) {
				c.Git.Retry.MaxAttempts = 0
			},
			wantErr:   true,
			errSubstr: "must be at least 1",
		},
		{
			name: "max_attempts negative",
			modify: func(c *Config) {
				c.Git.Retry.MaxAttempts = -1
			},
			wantErr:   true,
			errSubstr: "must be at least 1",
		},
		{
			name: "max_attempts exceeds limit",
			modify: func(c *Config) {
				c.Git.Retry.MaxAttempts = 11
			},
			wantErr:   true,
			errSubstr: "must not exceed 10",
		},
		{
			name: "initial_delay invalid format",
			modify: func(c *Config) {
				c.Git.Retry.InitialDelay = "not-a-duration"
			},
			wantErr:   true,
			errSubstr: "git.retry.initial_delay",
		},
		{
			name: "initial_delay zero",
			modify: func(c *Config) {
				c.Git.Retry.InitialDelay = "0s"
			},
			wantErr:   true,
			errSubstr: "must be positive",
		},
		{
			name: "initial_delay negative",
			modify: func(c *Config) {
				c.Git.Retry.InitialDelay = "-1s"
			},
			wantErr:   true,
			errSubstr: "must be positive",
		},
		{
			name: "max_delay invalid format",
			modify: func(c *Config) {
				c.Git.Retry.MaxDelay = "invalid"
			},
			wantErr:   true,
			errSubstr: "git.retry.max_delay",
		},
		{
			name: "max_delay zero",
			modify: func(c *Config) {
				c.Git.Retry.MaxDelay = "0s"
			},
			wantErr:   true,
			errSubstr: "must be positive",
		},
		{
			name: "initial_delay exceeds max_delay",
			modify: func(c *Config) {
				c.Git.Retry.InitialDelay = "1m"
				c.Git.Retry.MaxDelay = "30s"
			},
			wantErr:   true,
			errSubstr: "must not exceed max_delay",
		},
		{
			name: "multiplier below 1.0",
			modify: func(c *Config) {
				c.Git.Retry.Multiplier = 0.5
			},
			wantErr:   true,
			errSubstr: "must be at least 1.0",
		},
		{
			name: "jitter_factor negative",
			modify: func(c *Config) {
				c.Git.Retry.JitterFactor = -0.1
			},
			wantErr:   true,
			errSubstr: "must be between 0 and 1",
		},
		{
			name: "jitter_factor exceeds 1",
			modify: func(c *Config) {
				c.Git.Retry.JitterFactor = 1.5
			},
			wantErr:   true,
			errSubstr: "must be between 0 and 1",
		},
		{
			name: "edge case: multiplier exactly 1.0",
			modify: func(c *Config) {
				c.Git.Retry.Multiplier = 1.0
			},
			wantErr: false,
		},
		{
			name: "edge case: jitter_factor exactly 0",
			modify: func(c *Config) {
				c.Git.Retry.JitterFactor = 0.0
			},
			wantErr: false,
		},
		{
			name: "edge case: jitter_factor exactly 1",
			modify: func(c *Config) {
				c.Git.Retry.JitterFactor = 1.0
			},
			wantErr: false,
		},
		{
			name: "edge case: initial_delay equals max_delay",
			modify: func(c *Config) {
				c.Git.Retry.InitialDelay = "30s"
				c.Git.Retry.MaxDelay = "30s"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseConfig()
			tt.modify(cfg)

			err := cfg.ValidateValues()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateValues() expected error containing %q, got nil", tt.errSubstr)

					return
				}

				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateValues() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("ValidateValues() unexpected error: %v", err)
			}
		})
	}
}
