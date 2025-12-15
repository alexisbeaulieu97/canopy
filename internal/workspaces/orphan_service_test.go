package workspaces

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestWorkspaceOrphanService_DetectOrphans(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		workspaces    map[string]domain.Workspace
		listErr       error
		canonicalList []string
		canonicalErr  error
		wantOrphans   int
		wantErr       bool
	}{
		{
			name:          "no workspaces returns empty",
			workspaces:    map[string]domain.Workspace{},
			canonicalList: []string{},
			wantOrphans:   0,
			wantErr:       false,
		},
		{
			name:    "list error propagates",
			listErr: errors.New("list failed"),
			wantErr: true,
		},
		{
			name: "canonical list error propagates",
			workspaces: map[string]domain.Workspace{
				"ws1": {ID: "ws1"},
			},
			canonicalErr: errors.New("canonical list failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := mocks.NewMockGitOperations()
			mockGit.ListFunc = func() ([]string, error) {
				return tt.canonicalList, tt.canonicalErr
			}

			mockStorage := &mocks.MockWorkspaceStorage{
				ListFunc: func() (map[string]domain.Workspace, error) {
					return tt.workspaces, tt.listErr
				},
			}

			mockConfig := &mocks.MockConfigProvider{
				WorkspacesRoot: "/workspaces",
			}

			svc := NewOrphanService(mockConfig, mockGit, mockStorage, nil, nil)

			orphans, err := svc.DetectOrphans()

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectOrphans() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(orphans) != tt.wantOrphans {
				t.Errorf("DetectOrphans() got %d orphans, want %d", len(orphans), tt.wantOrphans)
			}
		})
	}
}

func TestWorkspaceOrphanService_DetectOrphansForWorkspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		workspace     *domain.Workspace
		dirName       string
		finderErr     error
		canonicalList []string
		canonicalErr  error
		wantOrphans   int
		wantErr       bool
	}{
		{
			name: "workspace with matching canonical but missing directory",
			workspace: &domain.Workspace{
				ID:    "ws1",
				Repos: []domain.Repo{{Name: "repo1"}},
			},
			dirName:       "ws1",
			canonicalList: []string{"repo1"},
			wantOrphans:   1, // Orphan because worktree directory doesn't exist
			wantErr:       false,
		},
		{
			name:      "workspace not found",
			finderErr: errors.New("not found"),
			wantErr:   true,
		},
		{
			name: "canonical list error",
			workspace: &domain.Workspace{
				ID: "ws1",
			},
			dirName:      "ws1",
			canonicalErr: errors.New("list failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := mocks.NewMockGitOperations()
			mockGit.ListFunc = func() ([]string, error) {
				return tt.canonicalList, tt.canonicalErr
			}

			mockConfig := &mocks.MockConfigProvider{
				WorkspacesRoot: "/workspaces",
			}

			finder := &mockWorkspaceFinder{
				workspace: tt.workspace,
				dirName:   tt.dirName,
				err:       tt.finderErr,
			}

			svc := NewOrphanService(mockConfig, mockGit, nil, nil, finder)

			orphans, err := svc.DetectOrphansForWorkspace("ws1")

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectOrphansForWorkspace() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(orphans) != tt.wantOrphans {
				t.Errorf("DetectOrphansForWorkspace() got %d orphans, want %d", len(orphans), tt.wantOrphans)
			}
		})
	}
}

func TestWorkspaceOrphanService_CheckRepoForOrphan(t *testing.T) {
	t.Parallel()

	// Create a temp directory structure for testing
	tmpDir := t.TempDir()
	workspacesRoot := filepath.Join(tmpDir, "workspaces")
	wsDir := filepath.Join(workspacesRoot, "ws1")
	repoDir := filepath.Join(wsDir, "repo1")
	gitDir := filepath.Join(repoDir, ".git")

	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	mockConfig := &mocks.MockConfigProvider{
		WorkspacesRoot: workspacesRoot,
	}

	svc := NewOrphanService(mockConfig, nil, nil, nil, nil)

	tests := []struct {
		name         string
		workspaceID  string
		repoName     string
		worktreePath string
		canonicalSet map[string]bool
		wantOrphan   bool
		wantReason   domain.OrphanReason
	}{
		{
			name:         "canonical missing",
			workspaceID:  "ws1",
			repoName:     "repo1",
			worktreePath: repoDir,
			canonicalSet: map[string]bool{},
			wantOrphan:   true,
			wantReason:   domain.OrphanReasonCanonicalMissing,
		},
		{
			name:         "worktree directory missing",
			workspaceID:  "ws1",
			repoName:     "repo1",
			worktreePath: filepath.Join(workspacesRoot, "ws1", "missing"),
			canonicalSet: map[string]bool{"repo1": true},
			wantOrphan:   true,
			wantReason:   domain.OrphanReasonDirectoryMissing,
		},
		{
			name:         "git dir missing",
			workspaceID:  "ws1",
			repoName:     "repo1",
			worktreePath: wsDir, // exists but no .git inside
			canonicalSet: map[string]bool{"repo1": true},
			wantOrphan:   true,
			wantReason:   domain.OrphanReasonInvalidGitDir,
		},
		{
			name:         "valid repo",
			workspaceID:  "ws1",
			repoName:     "repo1",
			worktreePath: repoDir,
			canonicalSet: map[string]bool{"repo1": true},
			wantOrphan:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orphan := svc.checkRepoForOrphan(tt.workspaceID, tt.repoName, tt.worktreePath, tt.canonicalSet)

			if tt.wantOrphan && orphan == nil {
				t.Errorf("checkRepoForOrphan() expected orphan, got nil")
			}

			if !tt.wantOrphan && orphan != nil {
				t.Errorf("checkRepoForOrphan() expected no orphan, got %v", orphan)
			}

			if tt.wantOrphan && orphan != nil && orphan.Reason != tt.wantReason {
				t.Errorf("checkRepoForOrphan() reason = %v, want %v", orphan.Reason, tt.wantReason)
			}
		})
	}
}
