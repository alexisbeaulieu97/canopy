package workspace

import (
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestMigrateWorkspace_V0ToV1(t *testing.T) {
	t.Parallel()

	ws := &domain.Workspace{
		Version:    0,
		ID:         "test-ws",
		BranchName: "main",
		Repos: []domain.Repo{
			{Name: "repo1", URL: "https://example.com/repo1.git"},
		},
	}

	migrated, err := MigrateWorkspace(ws)
	if err != nil {
		t.Fatalf("MigrateWorkspace failed: %v", err)
	}

	if !migrated {
		t.Error("expected migration to occur")
	}

	if ws.Version != domain.CurrentWorkspaceVersion {
		t.Errorf("expected version %d, got %d", domain.CurrentWorkspaceVersion, ws.Version)
	}

	// Verify data is preserved
	if ws.ID != "test-ws" {
		t.Errorf("expected ID 'test-ws', got %q", ws.ID)
	}

	if ws.BranchName != "main" {
		t.Errorf("expected BranchName 'main', got %q", ws.BranchName)
	}

	if len(ws.Repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(ws.Repos))
	}
}

func TestMigrateWorkspace_AlreadyCurrent(t *testing.T) {
	t.Parallel()

	ws := &domain.Workspace{
		Version:    domain.CurrentWorkspaceVersion,
		ID:         "test-ws",
		BranchName: "main",
	}

	migrated, err := MigrateWorkspace(ws)
	if err != nil {
		t.Fatalf("MigrateWorkspace failed: %v", err)
	}

	if migrated {
		t.Error("expected no migration for current version")
	}

	if ws.Version != domain.CurrentWorkspaceVersion {
		t.Errorf("expected version %d, got %d", domain.CurrentWorkspaceVersion, ws.Version)
	}
}

func TestMigrateWorkspace_FutureVersion(t *testing.T) {
	t.Parallel()

	ws := &domain.Workspace{
		Version: domain.CurrentWorkspaceVersion + 1,
		ID:      "test-ws",
	}

	migrated, err := MigrateWorkspace(ws)
	if err != nil {
		t.Fatalf("MigrateWorkspace failed: %v", err)
	}

	if migrated {
		t.Error("expected no migration for future version")
	}

	// Version should remain unchanged
	if ws.Version != domain.CurrentWorkspaceVersion+1 {
		t.Errorf("expected version %d, got %d", domain.CurrentWorkspaceVersion+1, ws.Version)
	}
}

func TestNeedsMigration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version int
		want    bool
	}{
		{"version 0 needs migration", 0, true},
		{"current version no migration", domain.CurrentWorkspaceVersion, false},
		{"future version no migration", domain.CurrentWorkspaceVersion + 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ws := &domain.Workspace{Version: tt.version}
			got := NeedsMigration(ws)

			if got != tt.want {
				t.Errorf("NeedsMigration() = %v, want %v", got, tt.want)
			}
		})
	}
}
