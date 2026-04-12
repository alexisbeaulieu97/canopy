package workspaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestWorkspaceHealthService_CheckWorkspace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		workspace  *domain.Workspace
		dirName    string
		finderErr  error
		wantErr    bool
		wantStatus domain.HealthStatus
	}{
		{
			name:      "workspace not found",
			finderErr: errors.New("not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := mocks.NewMockGitOperations()
			mockConfig := &mocks.MockConfigProvider{
				WorkspacesRoot: "/workspaces",
			}

			finder := &mockWorkspaceFinder{
				workspace: tt.workspace,
				dirName:   tt.dirName,
				err:       tt.finderErr,
			}

			svc := NewHealthService(mockConfig, mockGit, nil, nil, finder)

			report, err := svc.CheckWorkspace(context.Background(), "ws1", false)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckWorkspace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && report.OverallStatus != tt.wantStatus {
				t.Errorf("CheckWorkspace() status = %v, want %v", report.OverallStatus, tt.wantStatus)
			}
		})
	}
}

func TestWorkspaceHealthService_CheckAllWorkspaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		workspaces  map[string]domain.Workspace
		listErr     error
		wantReports int
		wantErr     bool
	}{
		{
			name:        "no workspaces returns empty",
			workspaces:  map[string]domain.Workspace{},
			wantReports: 0,
			wantErr:     false,
		},
		{
			name:    "list error propagates",
			listErr: errors.New("list failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGit := mocks.NewMockGitOperations()
			mockStorage := &mocks.MockWorkspaceStorage{
				ListFunc: func(_ context.Context) ([]domain.Workspace, error) {
					if tt.listErr != nil {
						return nil, tt.listErr
					}

					workspaces := make([]domain.Workspace, 0, len(tt.workspaces))
					for _, ws := range tt.workspaces {
						workspaces = append(workspaces, ws)
					}

					return workspaces, nil
				},
			}

			mockConfig := &mocks.MockConfigProvider{
				WorkspacesRoot: "/workspaces",
			}

			svc := NewHealthService(mockConfig, mockGit, mockStorage, nil, nil)

			reports, err := svc.CheckAllWorkspaces(context.Background(), false)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAllWorkspaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(reports) != tt.wantReports {
				t.Errorf("CheckAllWorkspaces() got %d reports, want %d", len(reports), tt.wantReports)
			}
		})
	}
}

func TestWorkspaceHealthService_CheckMetadataConsistency(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "ws1")

	// Create workspace directory with one repo
	repoPath := filepath.Join(workspacePath, "repo1")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	mockConfig := &mocks.MockConfigProvider{
		WorkspacesRoot: tmpDir,
	}

	svc := NewHealthService(mockConfig, nil, nil, nil, nil)

	tests := []struct {
		name          string
		workspace     *domain.Workspace
		workspacePath string
		wantChecks    int
		wantStatus    domain.HealthStatus
	}{
		{
			name: "missing workspace directory",
			workspace: &domain.Workspace{
				ID:    "ws1",
				Repos: []domain.Repo{{Name: "repo1"}},
			},
			workspacePath: filepath.Join(tmpDir, "nonexistent"),
			wantChecks:    1,
			wantStatus:    domain.HealthStatusCritical,
		},
		{
			name: "repo in metadata not on disk",
			workspace: &domain.Workspace{
				ID:    "ws1",
				Repos: []domain.Repo{{Name: "missing_repo"}},
			},
			workspacePath: workspacePath,
			wantChecks:    2, // One for missing repo, one for untracked "repo1" on disk
			wantStatus:    domain.HealthStatusCritical,
		},
		{
			name: "consistent metadata",
			workspace: &domain.Workspace{
				ID:    "ws1",
				Repos: []domain.Repo{{Name: "repo1"}},
			},
			workspacePath: workspacePath,
			wantChecks:    1, // Just the success check
			wantStatus:    domain.HealthStatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checks := svc.checkMetadataConsistency(tt.workspace, tt.workspacePath)

			if len(checks) != tt.wantChecks {
				t.Errorf("checkMetadataConsistency() got %d checks, want %d", len(checks), tt.wantChecks)
			}

			if len(checks) > 0 && checks[0].Status != tt.wantStatus {
				t.Errorf("checkMetadataConsistency() status = %v, want %v", checks[0].Status, tt.wantStatus)
			}
		})
	}
}

func TestWorkspaceHealthService_CheckMetadataConsistency_WorkspacePermissionError(t *testing.T) {
	requirePOSIXPermissions(t)

	tmpDir := t.TempDir()
	blockedParent := filepath.Join(tmpDir, "blocked")
	workspacePath := filepath.Join(blockedParent, "ws1")

	if err := os.MkdirAll(blockedParent, 0o755); err != nil {
		t.Fatalf("failed to create blocked parent: %v", err)
	}

	if err := os.Chmod(blockedParent, 0o000); err != nil {
		t.Fatalf("failed to remove permissions: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(blockedParent, 0o755)
	})

	svc := NewHealthService(&mocks.MockConfigProvider{WorkspacesRoot: tmpDir}, nil, nil, nil, nil)
	checks := svc.checkMetadataConsistency(&domain.Workspace{ID: "ws1"}, workspacePath)

	if len(checks) != 1 {
		t.Fatalf("checkMetadataConsistency() got %d checks, want 1", len(checks))
	}

	if checks[0].Status != domain.HealthStatusCritical {
		t.Fatalf("checkMetadataConsistency() status = %v, want %v", checks[0].Status, domain.HealthStatusCritical)
	}

	if checks[0].Description != "Cannot access workspace directory" {
		t.Fatalf("checkMetadataConsistency() description = %q, want %q", checks[0].Description, "Cannot access workspace directory")
	}
}

func TestWorkspaceHealthService_CheckMetadataConsistency_RepoAndReadDirPermissionErrors(t *testing.T) {
	requirePOSIXPermissions(t)

	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "ws1")

	if err := os.MkdirAll(workspacePath, 0o755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	if err := os.Chmod(workspacePath, 0o000); err != nil {
		t.Fatalf("failed to remove permissions: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(workspacePath, 0o755)
	})

	svc := NewHealthService(&mocks.MockConfigProvider{WorkspacesRoot: tmpDir}, nil, nil, nil, nil)
	checks := svc.checkMetadataConsistency(&domain.Workspace{
		ID:    "ws1",
		Repos: []domain.Repo{{Name: "repo1"}},
	}, workspacePath)

	if !hasHealthCheck(checks, "repo_exists:repo1", "Cannot access repository path from metadata") {
		t.Fatalf("checkMetadataConsistency() missing repository permission error check: %+v", checks)
	}

	if !hasHealthCheck(checks, "workspace_directory_scan", "Cannot read workspace directory") {
		t.Fatalf("checkMetadataConsistency() missing workspace read error check: %+v", checks)
	}
}

func TestWorkspaceHealthService_CheckWorktreeIntegrity(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a valid worktree structure
	worktreePath := filepath.Join(tmpDir, "repo1")
	gitFilePath := filepath.Join(worktreePath, ".git")

	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	// Create a .git file (worktree style)
	gitdirPath := filepath.Join(tmpDir, "canonical", ".git", "worktrees", "repo1")
	if err := os.MkdirAll(gitdirPath, 0o755); err != nil {
		t.Fatalf("failed to create gitdir: %v", err)
	}

	// Write .git file pointing to gitdir
	if err := os.WriteFile(gitFilePath, []byte("gitdir: "+gitdirPath), 0o644); err != nil {
		t.Fatalf("failed to write .git file: %v", err)
	}

	// Write gitdir file pointing back
	if err := os.WriteFile(filepath.Join(gitdirPath, "gitdir"), []byte(gitFilePath), 0o644); err != nil {
		t.Fatalf("failed to write gitdir file: %v", err)
	}

	mockConfig := &mocks.MockConfigProvider{
		WorkspacesRoot: tmpDir,
	}

	svc := NewHealthService(mockConfig, nil, nil, nil, nil)

	tests := []struct {
		name         string
		repoName     string
		worktreePath string
		wantStatus   domain.HealthStatus
	}{
		{
			name:         "missing .git file",
			repoName:     "missing",
			worktreePath: filepath.Join(tmpDir, "missing"),
			wantStatus:   domain.HealthStatusCritical,
		},
		{
			name:         "valid worktree",
			repoName:     "repo1",
			worktreePath: worktreePath,
			wantStatus:   domain.HealthStatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			report := &domain.WorkspaceHealthReport{}
			checks := svc.checkWorktreeIntegrity(context.Background(), tt.repoName, tt.worktreePath, "main", false, report)

			if len(checks) == 0 {
				t.Fatalf("checkWorktreeIntegrity() returned no checks")
			}

			if checks[0].Status != tt.wantStatus {
				t.Errorf("checkWorktreeIntegrity() status = %v, want %v", checks[0].Status, tt.wantStatus)
			}
		})
	}
}

func TestWorkspaceHealthService_CheckWorktreeIntegrity_GitdirPermissionError(t *testing.T) {
	requirePOSIXPermissions(t)

	tmpDir := t.TempDir()
	worktreePath := filepath.Join(tmpDir, "repo1")
	gitFilePath := filepath.Join(worktreePath, ".git")
	blockedParent := filepath.Join(tmpDir, "blocked")
	gitdirPath := filepath.Join(blockedParent, "repo1")

	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	if err := os.MkdirAll(blockedParent, 0o755); err != nil {
		t.Fatalf("failed to create blocked parent: %v", err)
	}

	if err := os.WriteFile(gitFilePath, []byte("gitdir: "+gitdirPath), 0o644); err != nil {
		t.Fatalf("failed to write .git file: %v", err)
	}

	if err := os.Chmod(blockedParent, 0o000); err != nil {
		t.Fatalf("failed to remove permissions: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(blockedParent, 0o755)
	})

	svc := NewHealthService(&mocks.MockConfigProvider{WorkspacesRoot: tmpDir}, nil, nil, nil, nil)
	report := &domain.WorkspaceHealthReport{}
	checks := svc.checkWorktreeIntegrity(context.Background(), "repo1", worktreePath, "main", false, report)

	if len(checks) != 1 {
		t.Fatalf("checkWorktreeIntegrity() got %d checks, want 1", len(checks))
	}

	if checks[0].Description != "Cannot access referenced git directory" {
		t.Fatalf("checkWorktreeIntegrity() description = %q, want %q", checks[0].Description, "Cannot access referenced git directory")
	}

	if checks[0].Fixable {
		t.Fatalf("checkWorktreeIntegrity() fixable = true, want false")
	}
}

func TestWorkspaceHealthService_CheckGitConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a worktree with .git directory (regular git repo style for simplicity)
	repoPath := filepath.Join(tmpDir, "repo1")
	gitDir := filepath.Join(repoPath, ".git")

	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create git directory: %v", err)
	}

	configContent := `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/example/repo.git
`
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	svc := NewHealthService(nil, nil, nil, nil, nil)

	tests := []struct {
		name         string
		repoName     string
		worktreePath string
		wantStatus   domain.HealthStatus
	}{
		{
			name:         "valid git config",
			repoName:     "repo1",
			worktreePath: repoPath,
			wantStatus:   domain.HealthStatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checks := svc.checkGitConfig(tt.repoName, tt.worktreePath)

			if len(checks) == 0 {
				t.Fatalf("checkGitConfig() returned no checks")
			}

			if checks[0].Status != tt.wantStatus {
				t.Errorf("checkGitConfig() status = %v, want %v", checks[0].Status, tt.wantStatus)
			}
		})
	}
}

func TestWorkspaceHealthService_CheckGitConfig_LinkedWorktreeUsesCanonicalConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	worktreePath := filepath.Join(tmpDir, "repo1")
	gitFilePath := filepath.Join(worktreePath, ".git")
	gitdirPath := filepath.Join(tmpDir, "canonical", ".git", "worktrees", "repo1")
	canonicalConfigPath := filepath.Join(tmpDir, "canonical", ".git", "config")

	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	if err := os.MkdirAll(gitdirPath, 0o755); err != nil {
		t.Fatalf("failed to create worktree gitdir: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(canonicalConfigPath), 0o755); err != nil {
		t.Fatalf("failed to create canonical git directory: %v", err)
	}

	if err := os.WriteFile(gitFilePath, []byte("gitdir: "+gitdirPath), 0o644); err != nil {
		t.Fatalf("failed to write .git file: %v", err)
	}

	configContent := `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/example/repo.git
`
	if err := os.WriteFile(canonicalConfigPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write canonical config file: %v", err)
	}

	svc := NewHealthService(nil, nil, nil, nil, nil)
	checks := svc.checkGitConfig("repo1", worktreePath)

	if len(checks) == 0 {
		t.Fatalf("checkGitConfig() returned no checks")
	}

	if checks[0].Status != domain.HealthStatusHealthy {
		t.Fatalf("checkGitConfig() status = %v, want %v", checks[0].Status, domain.HealthStatusHealthy)
	}
}

func TestWorkspaceHealthService_CheckRemoteURLValidity(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a repo with git config containing remote URL
	repoPath := filepath.Join(tmpDir, "repo1")
	gitDir := filepath.Join(repoPath, ".git")

	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create git directory: %v", err)
	}

	configContent := `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/example/repo.git
`
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	svc := NewHealthService(nil, nil, nil, nil, nil)

	tests := []struct {
		name         string
		repoName     string
		worktreePath string
		wantStatus   domain.HealthStatus
	}{
		{
			name:         "valid https URL",
			repoName:     "repo1",
			worktreePath: repoPath,
			wantStatus:   domain.HealthStatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checks := svc.checkRemoteURLValidity(tt.repoName, tt.worktreePath)

			if len(checks) == 0 {
				t.Fatalf("checkRemoteURLValidity() returned no checks")
			}

			if checks[0].Status != tt.wantStatus {
				t.Errorf("checkRemoteURLValidity() status = %v, want %v", checks[0].Status, tt.wantStatus)
			}
		})
	}
}

func TestIsValidRemoteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url   string
		valid bool
	}{
		{"https://github.com/user/repo.git", true},
		{"http://github.com/user/repo.git", true},
		{"git@github.com:user/repo.git", true},
		{"git@example.com:team/repo.git", true},
		{"github.com:user/repo.git", true},
		{"ssh://git@github.com/user/repo.git", true},
		{"git://github.com/user/repo.git", true},
		{"file:///path/to/repo", true},
		{"/path/to/repo", true},
		{"~/src/repo", true},
		{"./repo", true},
		{"../repo", true},
		{"not-a-url", false},
		{"ftp://invalid-scheme.com", false},
		{"host:", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()

			if got := isValidRemoteURL(tt.url); got != tt.valid {
				t.Errorf("isValidRemoteURL(%q) = %v, want %v", tt.url, got, tt.valid)
			}
		})
	}
}

func TestCalculateOverallStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		checks []domain.HealthCheck
		want   domain.HealthStatus
	}{
		{
			name:   "empty checks returns healthy",
			checks: []domain.HealthCheck{},
			want:   domain.HealthStatusHealthy,
		},
		{
			name: "all healthy",
			checks: []domain.HealthCheck{
				{Status: domain.HealthStatusHealthy},
				{Status: domain.HealthStatusHealthy},
			},
			want: domain.HealthStatusHealthy,
		},
		{
			name: "one warning",
			checks: []domain.HealthCheck{
				{Status: domain.HealthStatusHealthy},
				{Status: domain.HealthStatusWarning},
			},
			want: domain.HealthStatusWarning,
		},
		{
			name: "one critical overrides warning",
			checks: []domain.HealthCheck{
				{Status: domain.HealthStatusWarning},
				{Status: domain.HealthStatusCritical},
			},
			want: domain.HealthStatusCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := calculateOverallStatus(tt.checks); got != tt.want {
				t.Errorf("calculateOverallStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRemoteURL(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "standard spacing",
			content: `[remote "origin"]
	url = https://github.com/example/repo.git
`,
			want: "https://github.com/example/repo.git",
		},
		{
			name: "no spaces around equals",
			content: `[remote "origin"]
	url=git@example.com:team/repo.git
`,
			want: "git@example.com:team/repo.git",
		},
		{
			name: "quoted value",
			content: `[remote "origin"]
	url = "https://github.com/example/repo.git"
`,
			want: "https://github.com/example/repo.git",
		},
		{
			name: "surrounding whitespace",
			content: `[remote "origin"]
	  url =   '../repo.git'   
`,
			want: "../repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configPath := filepath.Join(tmpDir, strings.ReplaceAll(tt.name, " ", "_")+".config")
			if err := os.WriteFile(configPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			got, err := parseRemoteURL(configPath)
			if err != nil {
				t.Fatalf("parseRemoteURL() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("parseRemoteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkspaceHealthService_FixMissingWorktree_CanonicalPathPermissionError(t *testing.T) {
	requirePOSIXPermissions(t)

	tmpDir := t.TempDir()
	blockedProjectsRoot := filepath.Join(tmpDir, "projects")
	worktreePath := filepath.Join(tmpDir, "worktree")

	if err := os.MkdirAll(blockedProjectsRoot, 0o755); err != nil {
		t.Fatalf("failed to create projects root: %v", err)
	}

	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("failed to create worktree path: %v", err)
	}

	if err := os.Chmod(blockedProjectsRoot, 0o000); err != nil {
		t.Fatalf("failed to remove permissions: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(blockedProjectsRoot, 0o755)
	})

	svc := NewHealthService(&mocks.MockConfigProvider{ProjectsRoot: blockedProjectsRoot}, nil, nil, nil, nil)

	if err := svc.fixMissingWorktree(context.Background(), "repo1", worktreePath, "main"); err == nil {
		t.Fatal("fixMissingWorktree() error = nil, want permission error")
	}

	if _, statErr := os.Stat(worktreePath); statErr != nil {
		t.Fatalf("worktree path should remain untouched, stat error = %v", statErr)
	}
}

func TestResolveGitdirPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		gitdirPath   string
		worktreePath string
		want         string
	}{
		{
			name:         "absolute path unchanged",
			gitdirPath:   "/absolute/path/to/gitdir",
			worktreePath: "/worktree/path",
			want:         "/absolute/path/to/gitdir",
		},
		{
			name:         "relative path resolved from worktree",
			gitdirPath:   "../.git/worktrees/repo1",
			worktreePath: "/home/user/workspaces/ws1/repo1",
			want:         "/home/user/workspaces/ws1/.git/worktrees/repo1",
		},
		{
			name:         "simple relative path",
			gitdirPath:   ".git/worktrees/test",
			worktreePath: "/worktree",
			want:         "/worktree/.git/worktrees/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveGitdirPath(tt.gitdirPath, tt.worktreePath)
			if got != tt.want {
				t.Errorf("resolveGitdirPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func requirePOSIXPermissions(t *testing.T) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("permission-based filesystem tests are not reliable on windows")
	}

	if os.Geteuid() == 0 {
		t.Skip("permission-based filesystem tests are not reliable when running as root")
	}
}

func hasHealthCheck(checks []domain.HealthCheck, name, description string) bool {
	for _, check := range checks {
		if check.Name == name && check.Description == description {
			return true
		}
	}

	return false
}

func TestWorkspaceHealthService_ContextCancellation(t *testing.T) {
	t.Parallel()

	mockGit := mocks.NewMockGitOperations()
	mockStorage := &mocks.MockWorkspaceStorage{
		ListFunc: func(ctx context.Context) ([]domain.Workspace, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			return []domain.Workspace{{ID: "ws1"}}, nil
		},
	}

	mockConfig := &mocks.MockConfigProvider{
		WorkspacesRoot: "/workspaces",
	}

	svc := NewHealthService(mockConfig, mockGit, mockStorage, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.CheckAllWorkspaces(ctx, false)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}
