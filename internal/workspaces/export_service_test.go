package workspaces

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

// mockWorkspaceCreator implements WorkspaceCreator for testing.
type mockWorkspaceCreator struct {
	createFunc func(ctx context.Context, id, branchName string, repos []domain.Repo) (string, error)
	closeFunc  func(ctx context.Context, workspaceID string, force bool) error
}

func (m *mockWorkspaceCreator) CreateWorkspace(ctx context.Context, id, branchName string, repos []domain.Repo) (string, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, id, branchName, repos)
	}

	return id, nil
}

func (m *mockWorkspaceCreator) CloseWorkspace(ctx context.Context, workspaceID string, force bool) error {
	if m.closeFunc != nil {
		return m.closeFunc(ctx, workspaceID, force)
	}

	return nil
}

func TestWorkspaceExportService_ExportWorkspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		workspace    *domain.Workspace
		dirName      string
		finderErr    error
		registry     *config.RepoRegistry
		wantErr      bool
		wantReposCnt int
	}{
		{
			name: "export workspace successfully",
			workspace: &domain.Workspace{
				ID:         "test-ws",
				BranchName: "main",
				Repos: []domain.Repo{
					{Name: "repo1", URL: "https://example.com/repo1.git"},
					{Name: "repo2", URL: "https://example.com/repo2.git"},
				},
			},
			dirName:      "test-ws",
			wantErr:      false,
			wantReposCnt: 2,
		},
		{
			name:      "workspace not found",
			finderErr: errors.New("not found"),
			wantErr:   true,
		},
		{
			name: "export with registry alias",
			workspace: &domain.Workspace{
				ID:         "test-ws",
				BranchName: "main",
				Repos: []domain.Repo{
					{Name: "myorg/repo1", URL: "https://github.com/myorg/repo1.git"},
				},
			},
			dirName: "test-ws",
			registry: &config.RepoRegistry{
				Repos: map[string]config.RegistryEntry{
					"myorg/repo1": {Alias: "myorg/repo1", URL: "https://github.com/myorg/repo1.git"},
				},
			},
			wantErr:      false,
			wantReposCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockConfig := &mocks.MockConfigProvider{
				Registry: tt.registry,
			}

			finder := &mockWorkspaceFinder{
				workspace: tt.workspace,
				dirName:   tt.dirName,
				err:       tt.finderErr,
			}

			svc := NewExportService(mockConfig, finder, nil)

			export, err := svc.ExportWorkspace(context.Background(), "test-ws")

			if (err != nil) != tt.wantErr {
				t.Errorf("ExportWorkspace() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if export == nil {
					t.Fatal("ExportWorkspace() returned nil export")
				}

				if export.Version != "1" {
					t.Errorf("ExportWorkspace() version = %v, want 1", export.Version)
				}

				if len(export.Repos) != tt.wantReposCnt {
					t.Errorf("ExportWorkspace() got %d repos, want %d", len(export.Repos), tt.wantReposCnt)
				}

				if export.ID != tt.workspace.ID {
					t.Errorf("ExportWorkspace() ID = %v, want %v", export.ID, tt.workspace.ID)
				}
			}
		})
	}
}

func TestWorkspaceExportService_ImportWorkspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		export         *domain.WorkspaceExport
		idOverride     string
		branchOverride string
		force          bool
		workspace      *domain.Workspace // existing workspace for finder
		dirName        string
		finderErr      error
		createErr      error
		registry       *config.RepoRegistry
		wantErr        bool
	}{
		{
			name: "import new workspace",
			export: &domain.WorkspaceExport{
				Version: "1",
				ID:      "imported-ws",
				Branch:  "main",
				Repos: []domain.RepoExport{
					{Name: "repo1", URL: "https://example.com/repo1.git"},
				},
			},
			finderErr: cerrors.NewWorkspaceNotFound("imported-ws"), // workspace doesn't exist
			wantErr:   false,
		},
		{
			name:    "nil export",
			export:  nil,
			wantErr: true,
		},
		{
			name: "unsupported version",
			export: &domain.WorkspaceExport{
				Version: "2",
			},
			wantErr: true,
		},
		{
			name: "workspace exists without force",
			export: &domain.WorkspaceExport{
				Version: "1",
				ID:      "existing-ws",
				Branch:  "main",
				Repos:   []domain.RepoExport{},
			},
			workspace: &domain.Workspace{ID: "existing-ws"},
			dirName:   "existing-ws",
			force:     false,
			wantErr:   true,
		},
		{
			name: "workspace exists with force",
			export: &domain.WorkspaceExport{
				Version: "1",
				ID:      "existing-ws",
				Branch:  "main",
				Repos: []domain.RepoExport{
					{Name: "repo1", URL: "https://example.com/repo1.git"},
				},
			},
			workspace: &domain.Workspace{ID: "existing-ws"},
			dirName:   "existing-ws",
			force:     true,
			wantErr:   false,
		},
		{
			name: "id override",
			export: &domain.WorkspaceExport{
				Version: "1",
				ID:      "original-id",
				Branch:  "main",
				Repos: []domain.RepoExport{
					{Name: "repo1", URL: "https://example.com/repo1.git"},
				},
			},
			idOverride: "new-id",
			finderErr:  cerrors.NewWorkspaceNotFound("new-id"),
			wantErr:    false,
		},
		{
			name: "resolve via registry alias",
			export: &domain.WorkspaceExport{
				Version: "1",
				ID:      "test-ws",
				Branch:  "main",
				Repos: []domain.RepoExport{
					{Name: "repo1", Alias: "myorg/repo1"},
				},
			},
			finderErr: cerrors.NewWorkspaceNotFound("test-ws"),
			registry: &config.RepoRegistry{
				Repos: map[string]config.RegistryEntry{
					"myorg/repo1": {Alias: "myorg/repo1", URL: "https://github.com/myorg/repo1.git"},
				},
			},
			wantErr: false,
		},
		{
			name: "unresolvable repo",
			export: &domain.WorkspaceExport{
				Version: "1",
				ID:      "test-ws",
				Branch:  "main",
				Repos: []domain.RepoExport{
					{Name: "repo1"}, // no URL, no alias
				},
			},
			finderErr: cerrors.NewWorkspaceNotFound("test-ws"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockConfig := &mocks.MockConfigProvider{
				Registry: tt.registry,
			}

			finder := &mockWorkspaceFinder{
				workspace: tt.workspace,
				dirName:   tt.dirName,
				err:       tt.finderErr,
			}

			creator := &mockWorkspaceCreator{
				createFunc: func(_ context.Context, id, _ string, _ []domain.Repo) (string, error) {
					return id, tt.createErr
				},
				closeFunc: func(_ context.Context, _ string, _ bool) error {
					return nil
				},
			}

			svc := NewExportService(mockConfig, finder, creator)

			_, err := svc.ImportWorkspace(context.Background(), tt.export, tt.idOverride, tt.branchOverride, tt.force)

			if (err != nil) != tt.wantErr {
				t.Errorf("ImportWorkspace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkspaceExportService_ResolveImportOverrides(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		export         *domain.WorkspaceExport
		idOverride     string
		branchOverride string
		wantID         string
		wantBranch     string
	}{
		{
			name: "no overrides",
			export: &domain.WorkspaceExport{
				ID:     "original-id",
				Branch: "original-branch",
			},
			wantID:     "original-id",
			wantBranch: "original-branch",
		},
		{
			name: "id override",
			export: &domain.WorkspaceExport{
				ID:     "original-id",
				Branch: "original-branch",
			},
			idOverride: "new-id",
			wantID:     "new-id",
			wantBranch: "original-branch",
		},
		{
			name: "branch override",
			export: &domain.WorkspaceExport{
				ID:     "original-id",
				Branch: "original-branch",
			},
			branchOverride: "new-branch",
			wantID:         "original-id",
			wantBranch:     "new-branch",
		},
		{
			name: "empty branch defaults to workspace id",
			export: &domain.WorkspaceExport{
				ID:     "ws-id",
				Branch: "",
			},
			wantID:     "ws-id",
			wantBranch: "ws-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &WorkspaceExportService{}

			id, branch := svc.resolveImportOverrides(tt.export, tt.idOverride, tt.branchOverride)

			if id != tt.wantID {
				t.Errorf("resolveImportOverrides() id = %v, want %v", id, tt.wantID)
			}

			if branch != tt.wantBranch {
				t.Errorf("resolveImportOverrides() branch = %v, want %v", branch, tt.wantBranch)
			}
		})
	}
}

func TestWorkspaceExportService_ExportedAt(t *testing.T) {
	t.Parallel()

	mockConfig := &mocks.MockConfigProvider{}

	finder := &mockWorkspaceFinder{
		workspace: &domain.Workspace{
			ID:         "test-ws",
			BranchName: "main",
			Repos:      []domain.Repo{},
		},
		dirName: "test-ws",
	}

	svc := NewExportService(mockConfig, finder, nil)

	before := time.Now().UTC()
	export, err := svc.ExportWorkspace(context.Background(), "test-ws")
	after := time.Now().UTC()

	if err != nil {
		t.Fatalf("ExportWorkspace() error = %v", err)
	}

	if export.ExportedAt.Before(before) || export.ExportedAt.After(after) {
		t.Errorf("ExportWorkspace() ExportedAt = %v, want between %v and %v", export.ExportedAt, before, after)
	}
}
