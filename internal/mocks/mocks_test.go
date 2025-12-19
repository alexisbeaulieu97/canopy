package mocks_test

import (
	"errors"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/config"
	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

func TestMockHookExecutor(t *testing.T) {
	t.Parallel()

	t.Run("records calls", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockHookExecutor()

		hooks := []config.Hook{{Command: "echo test"}}
		ctx := domain.HookContext{WorkspaceID: "test-ws"}

		_, err := mock.ExecuteHooks(hooks, ctx, ports.HookExecuteOptions{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if mock.CallCount() != 1 {
			t.Errorf("expected 1 call, got %d", mock.CallCount())
		}

		if mock.ExecuteHooksCalls[0].Ctx.WorkspaceID != "test-ws" {
			t.Errorf("expected workspace ID 'test-ws', got %q", mock.ExecuteHooksCalls[0].Ctx.WorkspaceID)
		}
	})

	t.Run("returns configured error", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockHookExecutor()
		mock.ExecuteHooksErr = errors.New("hook failed")

		_, err := mock.ExecuteHooks(nil, domain.HookContext{}, ports.HookExecuteOptions{})
		if err == nil || err.Error() != "hook failed" {
			t.Errorf("expected 'hook failed' error, got %v", err)
		}
	})

	t.Run("calls custom function", func(t *testing.T) {
		t.Parallel()

		called := false
		mock := mocks.NewMockHookExecutor()
		mock.ExecuteHooksFunc = func(_ []config.Hook, _ domain.HookContext, _ ports.HookExecuteOptions) ([]domain.HookCommandPreview, error) {
			called = true
			return nil, nil
		}

		_, _ = mock.ExecuteHooks(nil, domain.HookContext{}, ports.HookExecuteOptions{})

		if !called {
			t.Error("custom function was not called")
		}
	})

	t.Run("reset clears calls", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockHookExecutor()
		_, _ = mock.ExecuteHooks(nil, domain.HookContext{}, ports.HookExecuteOptions{})

		mock.ResetCalls()

		if mock.CallCount() != 0 {
			t.Errorf("expected 0 calls after reset, got %d", mock.CallCount())
		}
	})
}

func TestMockDiskUsage(t *testing.T) {
	t.Parallel()

	t.Run("returns default values", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockDiskUsage()
		mock.DefaultUsage = 1024

		usage, _, err := mock.CachedUsage("/some/path")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if usage != 1024 {
			t.Errorf("expected usage 1024, got %d", usage)
		}

		if len(mock.CachedUsageCalls) != 1 || mock.CachedUsageCalls[0] != "/some/path" {
			t.Error("call not recorded correctly")
		}
	})

	t.Run("calls custom functions", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockDiskUsage()
		expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		mock.CalculateFunc = func(_ string) (int64, time.Time, error) {
			return 2048, expectedTime, nil
		}

		usage, modTime, err := mock.Calculate("/test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if usage != 2048 {
			t.Errorf("expected usage 2048, got %d", usage)
		}

		if !modTime.Equal(expectedTime) {
			t.Errorf("expected time %v, got %v", expectedTime, modTime)
		}
	})

	t.Run("records invalidate calls", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockDiskUsage()
		mock.InvalidateCache("/path1")
		mock.InvalidateCache("/path2")

		if len(mock.InvalidateCacheCalls) != 2 {
			t.Errorf("expected 2 invalidate calls, got %d", len(mock.InvalidateCacheCalls))
		}
	})

	t.Run("records clear cache calls", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockDiskUsage()
		mock.ClearCache()
		mock.ClearCache()

		if mock.ClearCacheCalls != 2 {
			t.Errorf("expected 2 clear cache calls, got %d", mock.ClearCacheCalls)
		}
	})

	t.Run("reset clears all calls", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockDiskUsage()
		_, _, _ = mock.CachedUsage("/test")
		_, _, _ = mock.Calculate("/test")
		mock.InvalidateCache("/test")
		mock.ClearCache()

		mock.ResetCalls()

		if len(mock.CachedUsageCalls) != 0 {
			t.Error("CachedUsageCalls not reset")
		}

		if len(mock.CalculateCalls) != 0 {
			t.Error("CalculateCalls not reset")
		}

		if len(mock.InvalidateCacheCalls) != 0 {
			t.Error("InvalidateCacheCalls not reset")
		}

		if mock.ClearCacheCalls != 0 {
			t.Error("ClearCacheCalls not reset")
		}
	})
}

func TestMockWorkspaceCache(t *testing.T) {
	t.Parallel()

	t.Run("stores and retrieves entries", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()

		ws := &domain.Workspace{ID: "test-ws", BranchName: "feature"}
		mock.Set("test-ws", ws, "test-ws-dir")

		retrieved, dirName, ok := mock.Get("test-ws")
		if !ok {
			t.Error("expected to find cached workspace")
		}

		if retrieved.ID != "test-ws" {
			t.Errorf("expected ID 'test-ws', got %q", retrieved.ID)
		}

		if dirName != "test-ws-dir" {
			t.Errorf("expected dir 'test-ws-dir', got %q", dirName)
		}
	})

	t.Run("returns not found for missing entries", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()

		_, _, ok := mock.Get("missing")
		if ok {
			t.Error("expected not found for missing entry")
		}
	})

	t.Run("invalidate removes entry", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()
		mock.Set("test-ws", &domain.Workspace{ID: "test-ws"}, "dir")

		mock.Invalidate("test-ws")

		_, _, ok := mock.Get("test-ws")
		if ok {
			t.Error("expected entry to be removed after invalidate")
		}
	})

	t.Run("invalidate all clears storage", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()
		mock.Set("ws1", &domain.Workspace{ID: "ws1"}, "dir1")
		mock.Set("ws2", &domain.Workspace{ID: "ws2"}, "dir2")

		mock.InvalidateAll()

		if mock.Size() != 0 {
			t.Errorf("expected empty cache, got size %d", mock.Size())
		}
	})

	t.Run("size returns entry count", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()
		mock.Set("ws1", &domain.Workspace{}, "dir1")
		mock.Set("ws2", &domain.Workspace{}, "dir2")

		if mock.Size() != 2 {
			t.Errorf("expected size 2, got %d", mock.Size())
		}
	})

	t.Run("records all calls", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()

		mock.Get("id1")
		mock.Set("id2", &domain.Workspace{}, "dir")
		mock.Invalidate("id3")
		mock.InvalidateAll()
		mock.Size()

		if len(mock.GetCalls) != 1 || mock.GetCalls[0] != "id1" {
			t.Error("Get calls not recorded correctly")
		}

		if len(mock.SetCalls) != 1 || mock.SetCalls[0].ID != "id2" {
			t.Error("Set calls not recorded correctly")
		}

		if len(mock.InvalidateCalls) != 1 || mock.InvalidateCalls[0] != "id3" {
			t.Error("Invalidate calls not recorded correctly")
		}

		if mock.InvalidateAllCalls != 1 {
			t.Error("InvalidateAll calls not recorded correctly")
		}

		if mock.SizeCalls != 1 {
			t.Error("Size calls not recorded correctly")
		}
	})

	t.Run("reset clears calls but not storage", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()
		mock.Set("ws", &domain.Workspace{}, "dir")
		mock.Get("ws")

		mock.ResetCalls()

		if len(mock.GetCalls) != 0 || len(mock.SetCalls) != 0 {
			t.Error("calls not reset")
		}

		// Storage should still have the entry
		_, _, ok := mock.Get("ws")
		if !ok {
			t.Error("storage was incorrectly cleared")
		}
	})

	t.Run("reset storage clears entries", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockWorkspaceCache()
		mock.Set("ws", &domain.Workspace{}, "dir")

		mock.ResetStorage()

		_, _, ok := mock.Get("ws")
		if ok {
			t.Error("storage not cleared")
		}
	})
}
