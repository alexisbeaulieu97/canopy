package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestNewInitializesDependencies(t *testing.T) {
	t.Helper()
	t.Cleanup(viper.Reset)
	viper.Reset()

	tempHome := t.TempDir()
	projectsRoot := filepath.Join(tempHome, "projects")
	workspacesRoot := filepath.Join(tempHome, "workspaces")

	configDir := filepath.Join(tempHome, ".yard")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := []byte("projects_root: \"" + projectsRoot + "\"\nworkspaces_root: \"" + workspacesRoot + "\"\n")
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, configContent, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("HOME", tempHome)

	app, err := New(false)
	if err != nil {
		t.Fatalf("expected app to initialize, got error: %v", err)
	}

	if app.Config == nil {
		t.Fatalf("expected config to be initialized")
	}
	if app.Config.ProjectsRoot != projectsRoot {
		t.Fatalf("unexpected projects root, got %s", app.Config.ProjectsRoot)
	}
	if app.Config.WorkspacesRoot != workspacesRoot {
		t.Fatalf("unexpected workspaces root, got %s", app.Config.WorkspacesRoot)
	}
	if app.Logger == nil {
		t.Fatalf("expected logger to be initialized")
	}
	if app.Service == nil {
		t.Fatalf("expected service to be initialized")
	}
}

func TestShutdownIsNoop(t *testing.T) {
	app := &App{}
	if err := app.Shutdown(); err != nil {
		t.Fatalf("expected shutdown to be noop, got %v", err)
	}
}
