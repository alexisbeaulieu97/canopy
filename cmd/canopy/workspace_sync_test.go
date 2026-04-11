package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/cobra"

	apppkg "github.com/alexisbeaulieu97/canopy/internal/app"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/logging"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
	"github.com/alexisbeaulieu97/canopy/internal/workspaces"
)

var workspaceSyncStdoutMutex sync.Mutex

type syncJSONResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type bulkSyncJSONResult struct {
	WorkspaceID string             `json:"workspace_id"`
	Result      *domain.SyncResult `json:"result"`
	Error       string             `json:"error"`
}

func TestWorkspaceSyncBulkJSONNoMatches(t *testing.T) {
	appInstance := newWorkspaceSyncTestApp(map[string]domain.Workspace{}, nil, nil)

	stdout, err := executeWorkspaceSyncCommand(t, appInstance, "--pattern", "^NOPE", "--json")
	if err != nil {
		t.Fatalf("workspace sync command failed: %v", err)
	}

	response := parseSyncJSONResponse(t, stdout)

	var payload []bulkSyncJSONResult
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload) != 0 {
		t.Fatalf("expected empty payload, got %d items", len(payload))
	}
}

func TestWorkspaceSyncBulkJSONSuccess(t *testing.T) {
	workspaces := map[string]domain.Workspace{
		"WS-1": {
			ID:      "WS-1",
			DirName: "WS-1",
			Repos:   []domain.Repo{{Name: "repo-a"}},
		},
		"WS-2": {
			ID:      "WS-2",
			DirName: "WS-2",
			Repos:   []domain.Repo{{Name: "repo-b"}},
		},
	}

	git := mocks.NewMockGitOperations()
	git.FetchFunc = func(context.Context, string) error { return nil }
	git.StatusFunc = func(_ context.Context, path string) (bool, int, int, string, error) {
		if filepath.Base(path) == "repo-a" {
			return false, 0, 2, "main", nil
		}

		return false, 0, 0, "main", nil
	}
	git.PullFunc = func(context.Context, string) error { return nil }

	appInstance := newWorkspaceSyncTestApp(workspaces, nil, git)

	stdout, err := executeWorkspaceSyncCommand(t, appInstance, "--pattern", "^WS-", "--json")
	if err != nil {
		t.Fatalf("workspace sync command failed: %v", err)
	}

	response := parseSyncJSONResponse(t, stdout)

	var payload []bulkSyncJSONResult
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload) != 2 {
		t.Fatalf("expected 2 payload items, got %d", len(payload))
	}

	if payload[0].Error != "" || payload[1].Error != "" {
		t.Fatalf("expected successful payload entries, got %+v", payload)
	}
}

func TestWorkspaceSyncBulkJSONReturnsNonZeroForWorkspaceErrors(t *testing.T) {
	workspaces := map[string]domain.Workspace{
		"WS-FAIL": {
			ID:      "WS-FAIL",
			DirName: "WS-FAIL",
			Repos:   []domain.Repo{{Name: "repo-fail"}},
		},
		"WS-OK": {
			ID:      "WS-OK",
			DirName: "WS-OK",
			Repos:   []domain.Repo{{Name: "repo-ok"}},
		},
	}

	storage := mocks.NewMockWorkspaceStorage()
	storage.Workspaces = workspaces
	storage.LoadFunc = func(_ context.Context, id string) (*domain.Workspace, error) {
		if id == "WS-FAIL" {
			return nil, cerrors.NewWorkspaceNotFound(id)
		}

		ws := storage.Workspaces[id]

		return &ws, nil
	}

	git := mocks.NewMockGitOperations()
	git.FetchFunc = func(context.Context, string) error { return nil }
	git.StatusFunc = func(context.Context, string) (bool, int, int, string, error) {
		return false, 0, 0, "main", nil
	}

	appInstance := newWorkspaceSyncTestApp(workspaces, storage, git)

	stdout, err := executeWorkspaceSyncCommand(t, appInstance, "--pattern", "^WS-", "--json")
	assertCommandFailed(t, err)

	response := parseSyncJSONResponse(t, stdout)

	var payload []bulkSyncJSONResult
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload) != 2 {
		t.Fatalf("expected 2 payload items, got %d", len(payload))
	}

	if payload[0].WorkspaceID != "WS-FAIL" || payload[0].Error == "" {
		t.Fatalf("expected failing workspace entry first, got %+v", payload[0])
	}
}

func TestWorkspaceSyncBulkJSONReturnsNonZeroForPartialResults(t *testing.T) {
	workspaces := map[string]domain.Workspace{
		"WS-PARTIAL": {
			ID:      "WS-PARTIAL",
			DirName: "WS-PARTIAL",
			Repos: []domain.Repo{
				{Name: "repo-ok"},
				{Name: "repo-fail"},
			},
		},
	}

	git := mocks.NewMockGitOperations()
	git.FetchFunc = func(_ context.Context, repoName string) error {
		if repoName == "repo-fail" {
			return errors.New("fetch failed")
		}

		return nil
	}
	git.StatusFunc = func(context.Context, string) (bool, int, int, string, error) {
		return false, 0, 1, "main", nil
	}
	git.PullFunc = func(context.Context, string) error { return nil }

	appInstance := newWorkspaceSyncTestApp(workspaces, nil, git)

	stdout, err := executeWorkspaceSyncCommand(t, appInstance, "--pattern", "^WS-", "--json")
	assertCommandFailed(t, err)

	response := parseSyncJSONResponse(t, stdout)

	var payload []bulkSyncJSONResult
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload) != 1 {
		t.Fatalf("expected 1 payload item, got %d", len(payload))
	}

	if payload[0].Result == nil || payload[0].Result.TotalErrors != 1 {
		t.Fatalf("expected partial sync result with one error, got %+v", payload[0].Result)
	}
}

func newWorkspaceSyncTestApp(
	workspaces map[string]domain.Workspace,
	storage *mocks.MockWorkspaceStorage,
	git *mocks.MockGitOperations,
) *apppkg.App {
	if storage == nil {
		storage = mocks.NewMockWorkspaceStorage()
	}

	storage.Workspaces = workspaces

	if git == nil {
		git = mocks.NewMockGitOperations()
	}

	cfg := mocks.NewMockConfigProvider()
	cfg.ParallelWorkers = 1
	cfg.WorkspacesRoot = "/workspaces"

	logger := logging.New(false)

	return &apppkg.App{
		Config:  cfg,
		Logger:  logger,
		Service: newWorkspaceSyncService(cfg, git, storage, logger),
	}
}

func newWorkspaceSyncService(
	cfg *mocks.MockConfigProvider,
	git *mocks.MockGitOperations,
	storage *mocks.MockWorkspaceStorage,
	logger *logging.Logger,
) *workspaces.Service {
	return workspaces.NewService(cfg, git, storage, logger, workspaces.WithHookExecutor(mocks.NewMockHookExecutor()))
}

func executeWorkspaceSyncCommand(t *testing.T, appInstance *apppkg.App, args ...string) (string, error) {
	t.Helper()

	workspaceSyncStdoutMutex.Lock()
	defer workspaceSyncStdoutMutex.Unlock()

	cmd := &cobra.Command{
		Use:           "sync [ID]",
		Args:          workspaceSyncCmd.Args,
		RunE:          workspaceSyncCmd.RunE,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.Flags().String("timeout", "60s", "Timeout for each repository sync (e.g. 30s, 2m)")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().String("pattern", "", "Sync workspaces matching a regex pattern")
	cmd.Flags().Bool("all", false, "Sync all workspaces (equivalent to --pattern \".*\")")
	cmd.Flags().Bool("no-progress", false, "Disable progress indicators")
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetContext(context.WithValue(context.Background(), appContextKey, appInstance))

	oldStdout := os.Stdout

	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create stdout pipe: %v", pipeErr)
	}

	defer func() { _ = r.Close() }()
	defer func() { os.Stdout = oldStdout }()

	os.Stdout = w
	err := cmd.Execute()
	_ = w.Close()

	var buf bytes.Buffer
	if _, copyErr := io.Copy(&buf, r); copyErr != nil {
		t.Fatalf("failed to capture stdout: %v", copyErr)
	}

	return buf.String(), err
}

func parseSyncJSONResponse(t *testing.T, stdout string) syncJSONResponse {
	t.Helper()

	var response syncJSONResponse
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, stdout)
	}

	if !response.Success {
		t.Fatalf("expected success response envelope, got %s", stdout)
	}

	return response
}

func assertCommandFailed(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected command to fail, got nil")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) {
		t.Fatalf("expected CanopyError, got %T", err)
	}

	if canopyErr.Code != cerrors.ErrCommandFailed {
		t.Fatalf("expected command failed error, got %s", canopyErr.Code)
	}
}
