package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	apppkg "github.com/alexisbeaulieu97/canopy/internal/app"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

func TestWorkspaceNewRequiresConfiguredRepos(t *testing.T) {
	appInstance := newFrictionTestApp(t, map[string]domain.Workspace{}, nil, nil, nil)

	_, err := executeWorkspaceNewCommand(t, appInstance, "TASK-123")
	if err == nil {
		t.Fatal("expected workspace new to fail without repos")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		t.Fatalf("expected CanopyError, got %T", err)
	}

	if canopyErr.Code != cerrors.ErrNoReposConfigured {
		t.Fatalf("expected ErrNoReposConfigured, got %s", canopyErr.Code)
	}

	if !strings.Contains(canopyErr.Message, "--repos") || !strings.Contains(canopyErr.Message, "--template") {
		t.Fatalf("expected actionable no-repo guidance, got %q", canopyErr.Message)
	}
}

func TestWorkspaceNewSucceedsWithExplicitRepoURL(t *testing.T) {
	appInstance := newFrictionTestApp(t, map[string]domain.Workspace{}, nil, nil, nil)

	stdout, err := executeWorkspaceNewCommand(t, appInstance, "TASK-123", "--repos", "https://github.com/example/backend.git")
	if err != nil {
		t.Fatalf("workspace new failed: %v", err)
	}

	if !strings.Contains(stdout, "Created workspace TASK-123") {
		t.Fatalf("expected success output, got %q", stdout)
	}
}

func TestWorkspaceListStatusShowsMultiSignalSummaryAndNoRepos(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	git := mocks.NewMockGitOperations()
	git.StatusFunc = func(_ context.Context, path string) (bool, int, int, string, error) {
		switch filepath.Base(path) {
		case "repo-a":
			return true, 2, 0, "feature/auth", nil
		case "repo-b":
			return false, 0, 1, "main", nil
		default:
			return false, 0, 0, "main", nil
		}
	}

	disk := mocks.NewMockDiskUsage()
	disk.DefaultUsage = 512
	disk.DefaultModTime = time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)

	workspacesData := map[string]domain.Workspace{
		"WS-A": {
			ID:         "WS-A",
			BranchName: "feature/auth",
			Repos: []domain.Repo{
				{Name: "repo-a", URL: "https://github.com/example/repo-a.git"},
				{Name: "repo-b", URL: "https://github.com/example/repo-b.git"},
			},
		},
		"WS-EMPTY": {
			ID:         "WS-EMPTY",
			BranchName: "main",
		},
	}

	appInstance := newFrictionTestApp(t, workspacesData, git, disk, nil)

	stdout, stderr, err := executeWorkspaceListCommand(t, appInstance, "--status")
	if err != nil {
		t.Fatalf("workspace list failed: %v", err)
	}

	if stderr != "" {
		t.Fatalf("expected no warning spam on stderr, got %q", stderr)
	}

	if !strings.Contains(stdout, "1 dirty") || !strings.Contains(stdout, "2 unpushed") || !strings.Contains(stdout, "1 behind") {
		t.Fatalf("expected multi-signal summary, got:\n%s", stdout)
	}

	if !strings.Contains(stdout, "no repos") {
		t.Fatalf("expected explicit no-repos state, got:\n%s", stdout)
	}
}

func TestWorkspaceListStatusDoesNotEmitPerWorkspaceWarnings(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	git := mocks.NewMockGitOperations()
	git.StatusFunc = func(context.Context, string) (bool, int, int, string, error) {
		return false, 0, 0, "", context.DeadlineExceeded
	}

	appInstance := newFrictionTestApp(t, map[string]domain.Workspace{
		"WS-ERR": {
			ID:    "WS-ERR",
			Repos: []domain.Repo{{Name: "repo-a", URL: "https://github.com/example/repo-a.git"}},
		},
	}, git, nil, nil)

	_, stderr, err := executeWorkspaceListCommand(t, appInstance, "--status")
	if err != nil {
		t.Fatalf("workspace list failed: %v", err)
	}

	if stderr != "" {
		t.Fatalf("expected stderr to stay empty, got %q", stderr)
	}
}

func TestWorkspaceViewShowsExpandedMetadataAndEmptyState(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	disk := mocks.NewMockDiskUsage()
	disk.DefaultUsage = 2048
	disk.DefaultModTime = time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)

	appInstance := newFrictionTestApp(t, map[string]domain.Workspace{
		"WS-EMPTY": {
			ID:         "WS-EMPTY",
			BranchName: "main",
		},
	}, nil, disk, nil)

	stdout, err := executeWorkspaceViewCommand(t, appInstance, "WS-EMPTY")
	if err != nil {
		t.Fatalf("workspace view failed: %v", err)
	}

	for _, want := range []string{
		"Path",
		"Disk",
		"Modified",
		"No repositories in this workspace.",
		"Add one with: canopy workspace repo add WS-EMPTY <repo>",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestRuntimeStatusCommandSuppressesUsageAfterParsing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := mocks.NewMockConfigProvider()
	cfg.WorkspacesRoot = filepath.Join(tmpDir, "workspaces")

	appInstance := &apppkg.App{
		Config: cfg,
		Logger: logging.New(false),
	}

	root := &cobra.Command{
		Use: "canopy",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			prepareRuntimeErrorHandling(cmd)
			ctx := context.WithValue(cmd.Context(), appContextKey, appInstance)
			cmd.SetContext(ctx)
			cmd.Root().SetContext(ctx)

			return nil
		},
	}

	status := &cobra.Command{
		Use:   statusCmd.Use,
		Short: statusCmd.Short,
		RunE:  statusCmd.RunE,
	}
	status.Flags().Bool("json", false, "Output in JSON format")
	root.AddCommand(status)
	root.SetArgs([]string{"status"})

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	root.SetOut(&stdout)
	root.SetErr(&stderr)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	defer func() { _ = os.Chdir(originalWD) }()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected status command to fail outside a workspace")
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no Cobra usage output, got %q", stderr.String())
	}
}

func newFrictionTestApp(
	t *testing.T,
	workspacesData map[string]domain.Workspace,
	git *mocks.MockGitOperations,
	disk *mocks.MockDiskUsage,
	repoNames []string,
) *apppkg.App {
	t.Helper()

	if git == nil {
		git = mocks.NewMockGitOperations()
	}

	if disk == nil {
		disk = mocks.NewMockDiskUsage()
	}

	storage := mocks.NewMockWorkspaceStorage()
	storage.Workspaces = workspacesData

	cfg := mocks.NewMockConfigProvider()
	cfg.WorkspacesRoot = t.TempDir()
	cfg.ClosedRoot = t.TempDir()
	cfg.RepoNames = repoNames

	logger := logging.New(false)

	return &apppkg.App{
		Config:  cfg,
		Logger:  logger,
		Service: workspaces.NewService(cfg, git, storage, logger, workspaces.WithDiskUsage(disk), workspaces.WithHookExecutor(mocks.NewMockHookExecutor())),
	}
}

func executeWorkspaceNewCommand(t *testing.T, appInstance *apppkg.App, args ...string) (string, error) {
	t.Helper()

	cmd := &cobra.Command{
		Use:           workspaceNewCmd.Use,
		Args:          workspaceNewCmd.Args,
		RunE:          workspaceNewCmd.RunE,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().StringSlice("repos", []string{}, "List of repositories to include")
	cmd.Flags().String("branch", "", "Custom branch name (optional)")
	cmd.Flags().Bool("print-path", false, "Print the created workspace path to stdout")
	cmd.Flags().Bool("no-hooks", false, "Skip post_create hooks")
	cmd.Flags().Bool("hooks-only", false, "Run post_create hooks without creating the workspace")
	cmd.Flags().Bool("dry-run-hooks", false, "Preview post_create hooks without executing them")
	cmd.Flags().Bool("json", false, "Output in JSON format (use with --dry-run-hooks)")
	cmd.Flags().String("template", "", "Workspace template to apply")
	cmd.SetArgs(args)
	cmd.SetContext(context.WithValue(context.Background(), appContextKey, appInstance))

	return captureStdout(t, func() error { return cmd.Execute() })
}

func executeWorkspaceListCommand(t *testing.T, appInstance *apppkg.App, args ...string) (string, string, error) {
	t.Helper()

	cmd := &cobra.Command{
		Use:           workspaceListCmd.Use,
		RunE:          workspaceListCmd.RunE,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("closed", false, "List closed workspaces")
	cmd.Flags().Bool("status", false, "Show git status for each repository")
	cmd.Flags().Bool("parallel-status", true, "Fetch workspace status in parallel (default)")
	cmd.Flags().Bool("sequential-status", false, "Fetch workspace status sequentially")
	cmd.Flags().String("timeout", "5s", "Timeout for status check per workspace (e.g. 5s, 10s)")
	cmd.Flags().Bool("show-locks", false, "Show workspace lock status")
	cmd.SetArgs(args)
	cmd.SetContext(context.WithValue(context.Background(), appContextKey, appInstance))

	return captureStdoutAndStderr(t, func() error { return cmd.Execute() })
}

func executeWorkspaceViewCommand(t *testing.T, appInstance *apppkg.App, id string) (string, error) {
	t.Helper()

	cmd := &cobra.Command{
		Use:           workspaceViewCmd.Use,
		Args:          workspaceViewCmd.Args,
		RunE:          workspaceViewCmd.RunE,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.SetArgs([]string{id})
	cmd.SetContext(context.WithValue(context.Background(), appContextKey, appInstance))

	return captureStdout(t, func() error { return cmd.Execute() })
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	defer func() {
		_ = r.Close()
		os.Stdout = oldStdout
	}()

	os.Stdout = w
	runErr := fn()
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	return buf.String(), runErr
}

func captureStdoutAndStderr(t *testing.T, fn func() error) (string, string, error) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	defer func() {
		_ = stdoutR.Close()
		_ = stderrR.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	runErr := fn()
	_ = stdoutW.Close()
	_ = stderrW.Close()

	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)

	if _, err := io.Copy(&stdoutBuf, stdoutR); err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	if _, err := io.Copy(&stderrBuf, stderrR); err != nil {
		t.Fatalf("failed to capture stderr: %v", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), runErr
}
