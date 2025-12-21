package workspaces

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestGetWorkspaceStatusBatch(t *testing.T) {
	t.Parallel()

	mockStorage := mocks.NewMockWorkspaceStorage()
	mockStorage.Workspaces["WS-A"] = domain.Workspace{
		ID:         "WS-A",
		BranchName: "main",
		Repos: []domain.Repo{
			{Name: "repo-a"},
			{Name: "repo-b"},
		},
	}
	mockStorage.Workspaces["WS-B"] = domain.Workspace{
		ID:         "WS-B",
		BranchName: "main",
		Repos: []domain.Repo{
			{Name: "repo-c"},
		},
	}

	mockGit := mocks.NewMockGitOperations()
	mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
		return false, 1, 0, "main", nil
	}

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.ParallelWorkers = 2

	svc := NewService(mockConfig, mockGit, mockStorage, nil)

	workspaceIDs := []string{"WS-B", "WS-A"}

	results, err := svc.GetWorkspaceStatusBatch(context.Background(), workspaceIDs, 0)
	if err != nil {
		t.Fatalf("GetWorkspaceStatusBatch failed: %v", err)
	}

	if len(results) != len(workspaceIDs) {
		t.Fatalf("expected %d results, got %d", len(workspaceIDs), len(results))
	}

	if results[0].WorkspaceID != "WS-B" || results[1].WorkspaceID != "WS-A" {
		t.Fatalf("results out of order: %+v", results)
	}

	if results[0].Status == nil || len(results[0].Status.Repos) != 1 {
		t.Fatalf("expected status for WS-B with 1 repo, got %+v", results[0].Status)
	}

	if results[1].Status == nil || len(results[1].Status.Repos) != 2 {
		t.Fatalf("expected status for WS-A with 2 repos, got %+v", results[1].Status)
	}
}

func TestGetWorkspaceStatusBatch_OrderPreserved(t *testing.T) {
	t.Parallel()

	mockStorage := mocks.NewMockWorkspaceStorage()
	mockStorage.Workspaces["WS-A"] = domain.Workspace{
		ID:         "WS-A",
		BranchName: "main",
		Repos:      []domain.Repo{{Name: "repo-a"}},
	}
	mockStorage.Workspaces["WS-B"] = domain.Workspace{
		ID:         "WS-B",
		BranchName: "main",
		Repos:      []domain.Repo{{Name: "repo-b"}},
	}

	mockGit := mocks.NewMockGitOperations()
	mockGit.StatusFunc = func(_ context.Context, path string) (bool, int, int, string, error) {
		if strings.Contains(path, "WS-A") {
			time.Sleep(25 * time.Millisecond)
		}

		return false, 0, 0, "main", nil
	}

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.ParallelWorkers = 2

	svc := NewService(mockConfig, mockGit, mockStorage, nil)

	workspaceIDs := []string{"WS-A", "WS-B"}

	results, err := svc.GetWorkspaceStatusBatch(context.Background(), workspaceIDs, 0)
	if err != nil {
		t.Fatalf("GetWorkspaceStatusBatch failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].WorkspaceID != "WS-A" || results[1].WorkspaceID != "WS-B" {
		t.Fatalf("results out of order: %+v", results)
	}
}

func BenchmarkGetWorkspaceStatusBatch(b *testing.B) {
	benchmarkStatusBatch(b, 1)
	benchmarkStatusBatch(b, 4)
}

func benchmarkStatusBatch(b *testing.B, workers int) {
	b.Helper()

	mockStorage := mocks.NewMockWorkspaceStorage()
	workspaceIDs := make([]string, 0, 20)

	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("WS-%02d", i)
		workspaceIDs = append(workspaceIDs, id)
		mockStorage.Workspaces[id] = domain.Workspace{
			ID:         id,
			BranchName: "main",
			Repos: []domain.Repo{
				{Name: "repo-a"},
				{Name: "repo-b"},
				{Name: "repo-c"},
			},
		}
	}

	mockGit := mocks.NewMockGitOperations()
	mockGit.StatusFunc = func(_ context.Context, _ string) (bool, int, int, string, error) {
		for i := 0; i < 1000; i++ {
			_ = i * i
		}

		return false, 0, 0, "main", nil
	}

	mockConfig := mocks.NewMockConfigProvider()
	mockConfig.ParallelWorkers = workers

	svc := NewService(mockConfig, mockGit, mockStorage, nil)

	b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = svc.GetWorkspaceStatusBatch(context.Background(), workspaceIDs, 0)
		}
	})
}
