package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestNewInitializesDependencies(t *testing.T) {
	t.Helper()
	t.Cleanup(viper.Reset)
	viper.Reset()

	tempHome := t.TempDir()
	projectsRoot := filepath.Join(tempHome, "projects")
	workspacesRoot := filepath.Join(tempHome, "workspaces")

	configDir := filepath.Join(tempHome, ".canopy")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	closedRoot := filepath.Join(tempHome, "closed")
	configContent := []byte("projects_root: \"" + projectsRoot + "\"\nworkspaces_root: \"" + workspacesRoot + "\"\nclosed_root: \"" + closedRoot + "\"\nworkspace_close_default: \"delete\"\n")

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

	if app.Config.GetProjectsRoot() != projectsRoot {
		t.Fatalf("unexpected projects root, got %s", app.Config.GetProjectsRoot())
	}

	if app.Config.GetWorkspacesRoot() != workspacesRoot {
		t.Fatalf("unexpected workspaces root, got %s", app.Config.GetWorkspacesRoot())
	}

	if app.Config.GetClosedRoot() != closedRoot {
		t.Fatalf("unexpected closed root, got %s", app.Config.GetClosedRoot())
	}

	if app.Config.GetCloseDefault() != "delete" {
		t.Fatalf("unexpected close default, got %s", app.Config.GetCloseDefault())
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

func TestNewWithMockedDependencies(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.ProjectsRoot = "/mock/projects"
	mockConfig.WorkspacesRoot = "/mock/workspaces"
	mockConfig.ClosedRoot = "/mock/closed"
	mockConfig.CloseDefault = "archive"

	mockGit := mocks.NewMockGitOperations()
	mockStorage := mocks.NewMockWorkspaceStorage()
	mockLogger := logging.New(false)

	app, err := New(
		false,
		WithConfigProvider(mockConfig),
		WithGitOperations(mockGit),
		WithWorkspaceStorage(mockStorage),
		WithLogger(mockLogger),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if app.Config.GetProjectsRoot() != "/mock/projects" {
		t.Fatalf("expected mock projects root, got %s", app.Config.GetProjectsRoot())
	}

	if app.Config.GetWorkspacesRoot() != "/mock/workspaces" {
		t.Fatalf("expected mock workspaces root, got %s", app.Config.GetWorkspacesRoot())
	}

	if app.Config.GetClosedRoot() != "/mock/closed" {
		t.Fatalf("expected mock closed root, got %s", app.Config.GetClosedRoot())
	}

	if app.Config.GetCloseDefault() != "archive" {
		t.Fatalf("expected mock close default, got %s", app.Config.GetCloseDefault())
	}

	if app.Logger != mockLogger {
		t.Fatal("expected mock logger to be used")
	}

	if app.Service == nil {
		t.Fatal("expected service to be initialized")
	}
}

func TestNewWithPartialMockedDependencies(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.ProjectsRoot = "/partial/projects"
	mockConfig.WorkspacesRoot = "/partial/workspaces"
	mockConfig.ClosedRoot = "/partial/closed"

	// Only provide config, let other dependencies use defaults
	app, err := New(
		false,
		WithConfigProvider(mockConfig),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if app.Config.GetProjectsRoot() != "/partial/projects" {
		t.Fatalf("expected mock projects root, got %s", app.Config.GetProjectsRoot())
	}

	// Logger should be created automatically
	if app.Logger == nil {
		t.Fatal("expected logger to be initialized")
	}

	// Service should be created with default git and storage engines
	if app.Service == nil {
		t.Fatal("expected service to be initialized")
	}
}

func TestNewWithCustomLogger(t *testing.T) {
	mockConfig := mocks.NewMockConfigProvider()
	customLogger := logging.New(true) // debug mode

	app, err := New(
		false, // debug flag ignored when custom logger provided
		WithConfigProvider(mockConfig),
		WithLogger(customLogger),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if app.Logger != customLogger {
		t.Fatal("expected custom logger to be used")
	}
}
