package workspaces

import (
	"context"
	"errors"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// mockWorkspaceFinder implements WorkspaceFinder for testing.
type mockWorkspaceFinder struct {
	workspace *domain.Workspace
	dirName   string
	err       error
	lastCtx   context.Context
}

func (m *mockWorkspaceFinder) FindWorkspace(ctx context.Context, _ string) (*domain.Workspace, string, error) {
	m.lastCtx = ctx
	if ctx.Err() != nil {
		return nil, "", ctx.Err()
	}

	if m.err != nil {
		return nil, "", m.err
	}

	return m.workspace, m.dirName, nil
}

func TestWorkspaceGitService_PushWorkspace_ContextCancellation(t *testing.T) {
	t.Parallel()

	mockConfig := &mocks.MockConfigProvider{
		WorkspacesRoot: "/workspaces",
	}

	finder := &mockWorkspaceFinder{
		workspace: &domain.Workspace{
			ID:    "test-ws",
			Repos: []domain.Repo{{Name: "repo1", URL: "https://example.com/repo1.git"}},
		},
		dirName: "test-ws",
	}

	svc := NewGitService(mockConfig, mocks.NewMockGitOperations(), nil, nil, nil, finder)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.PushWorkspace(ctx, "test-ws")

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got %v", err)
	}

	if finder.lastCtx == nil {
		t.Fatal("expected workspace finder to receive context")
	}
}

func TestWorkspaceGitService_PushWorkspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		workspace      *domain.Workspace
		dirName        string
		finderErr      error
		pushErr        error
		wantErr        bool
		expectedPushes int
	}{
		{
			name: "push all repos successfully",
			workspace: &domain.Workspace{
				ID:         "test-ws",
				BranchName: "main",
				Repos: []domain.Repo{
					{Name: "repo1", URL: "https://example.com/repo1.git"},
					{Name: "repo2", URL: "https://example.com/repo2.git"},
				},
			},
			dirName:        "test-ws",
			wantErr:        false,
			expectedPushes: 2,
		},
		{
			name:      "workspace not found",
			finderErr: errors.New("workspace not found"),
			wantErr:   true,
		},
		{
			name: "push error propagates",
			workspace: &domain.Workspace{
				ID:         "test-ws",
				BranchName: "main",
				Repos: []domain.Repo{
					{Name: "repo1", URL: "https://example.com/repo1.git"},
				},
			},
			dirName:        "test-ws",
			pushErr:        errors.New("push failed"),
			wantErr:        true,
			expectedPushes: 1,
		},
		{
			name: "empty repos succeeds",
			workspace: &domain.Workspace{
				ID:         "test-ws",
				BranchName: "main",
				Repos:      []domain.Repo{},
			},
			dirName:        "test-ws",
			wantErr:        false,
			expectedPushes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pushCount := 0
			mockGit := mocks.NewMockGitOperations()
			mockGit.PushFunc = func(_ context.Context, _, _ string) error {
				pushCount++
				return tt.pushErr
			}

			mockConfig := &mocks.MockConfigProvider{
				WorkspacesRoot: "/workspaces",
			}

			finder := &mockWorkspaceFinder{
				workspace: tt.workspace,
				dirName:   tt.dirName,
				err:       tt.finderErr,
			}

			svc := NewGitService(mockConfig, mockGit, nil, nil, nil, finder)

			err := svc.PushWorkspace(context.Background(), "test-ws")

			if (err != nil) != tt.wantErr {
				t.Errorf("PushWorkspace() error = %v, wantErr %v", err, tt.wantErr)
			}

			if pushCount != tt.expectedPushes {
				t.Errorf("PushWorkspace() push count = %d, want %d", pushCount, tt.expectedPushes)
			}
		})
	}
}

func TestWorkspaceGitService_RunGitInWorkspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		workspace   *domain.Workspace
		dirName     string
		finderErr   error
		cmdResult   *ports.CommandResult
		cmdErr      error
		opts        GitRunOptions
		wantErr     bool
		wantResults int
	}{
		{
			name: "sequential execution success",
			workspace: &domain.Workspace{
				ID: "test-ws",
				Repos: []domain.Repo{
					{Name: "repo1"},
					{Name: "repo2"},
				},
			},
			dirName: "test-ws",
			cmdResult: &ports.CommandResult{
				Stdout:   "output",
				ExitCode: 0,
			},
			opts:        GitRunOptions{Parallel: false, ContinueOnError: false},
			wantErr:     false,
			wantResults: 2,
		},
		{
			name: "parallel execution success",
			workspace: &domain.Workspace{
				ID: "test-ws",
				Repos: []domain.Repo{
					{Name: "repo1"},
					{Name: "repo2"},
				},
			},
			dirName: "test-ws",
			cmdResult: &ports.CommandResult{
				Stdout:   "output",
				ExitCode: 0,
			},
			opts:        GitRunOptions{Parallel: true, ContinueOnError: false},
			wantErr:     false,
			wantResults: 2,
		},
		{
			name:      "workspace not found",
			finderErr: errors.New("not found"),
			wantErr:   true,
		},
		{
			name: "empty repos returns nil",
			workspace: &domain.Workspace{
				ID:    "test-ws",
				Repos: []domain.Repo{},
			},
			dirName:     "test-ws",
			wantErr:     false,
			wantResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := mocks.NewMockGitOperations()
			mockGit.RunCommandFunc = func(_ context.Context, _ string, _ ...string) (*ports.CommandResult, error) {
				if tt.cmdErr != nil {
					return nil, tt.cmdErr
				}

				return tt.cmdResult, nil
			}

			mockConfig := &mocks.MockConfigProvider{
				WorkspacesRoot: "/workspaces",
			}

			finder := &mockWorkspaceFinder{
				workspace: tt.workspace,
				dirName:   tt.dirName,
				err:       tt.finderErr,
			}

			svc := NewGitService(mockConfig, mockGit, nil, nil, nil, finder)

			results, err := svc.RunGitInWorkspace(context.Background(), "test-ws", []string{"status"}, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunGitInWorkspace() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(results) != tt.wantResults {
				t.Errorf("RunGitInWorkspace() got %d results, want %d", len(results), tt.wantResults)
			}
		})
	}
}

func TestWorkspaceGitService_SwitchBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		workspace   *domain.Workspace
		dirName     string
		finderErr   error
		checkoutErr error
		saveErr     error
		wantErr     bool
	}{
		{
			name: "switch branch successfully",
			workspace: &domain.Workspace{
				ID:         "test-ws",
				BranchName: "main",
				Repos: []domain.Repo{
					{Name: "repo1"},
				},
			},
			dirName: "test-ws",
			wantErr: false,
		},
		{
			name:      "workspace not found",
			finderErr: errors.New("not found"),
			wantErr:   true,
		},
		{
			name: "checkout error",
			workspace: &domain.Workspace{
				ID: "test-ws",
				Repos: []domain.Repo{
					{Name: "repo1"},
				},
			},
			dirName:     "test-ws",
			checkoutErr: errors.New("checkout failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := mocks.NewMockGitOperations()
			mockGit.CheckoutFunc = func(_ context.Context, _, _ string, _ bool) error {
				return tt.checkoutErr
			}

			mockStorage := &mocks.MockWorkspaceStorage{
				SaveFunc: func(_ context.Context, _ domain.Workspace) error {
					return tt.saveErr
				},
			}

			mockConfig := &mocks.MockConfigProvider{
				WorkspacesRoot: "/workspaces",
			}

			mockCache := &mocks.MockWorkspaceCache{}

			finder := &mockWorkspaceFinder{
				workspace: tt.workspace,
				dirName:   tt.dirName,
				err:       tt.finderErr,
			}

			svc := NewGitService(mockConfig, mockGit, mockStorage, nil, mockCache, finder)

			err := svc.SwitchBranch(context.Background(), "test-ws", "feature", false)

			if (err != nil) != tt.wantErr {
				t.Errorf("SwitchBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
