package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/mocks"
)

func TestLoadWorkspaceViewData(t *testing.T) {
	t.Parallel()

	modifiedAt := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	git := mocks.NewMockGitOperations()
	disk := mocks.NewMockDiskUsage()
	disk.DefaultUsage = 1024
	disk.DefaultModTime = modifiedAt

	appInstance := newFrictionTestApp(t, map[string]domain.Workspace{
		"WS-1": {
			ID:         "WS-1",
			BranchName: "main",
			Repos:      []domain.Repo{{Name: "repo-a", URL: "https://github.com/example/repo-a.git"}},
		},
	}, git, disk, nil)

	viewData, err := loadWorkspaceViewData(context.Background(), appInstance.Service, "WS-1")
	if err != nil {
		t.Fatalf("loadWorkspaceViewData failed: %v", err)
	}

	if viewData.Path == "" {
		t.Fatal("expected workspace path to be populated")
	}

	if viewData.Workspace.DiskUsageBytes != 1024 {
		t.Fatalf("expected disk usage to be loaded, got %d", viewData.Workspace.DiskUsageBytes)
	}

	if viewData.Workspace.LastModified.IsZero() {
		t.Fatal("expected last modified to be populated")
	}
}

func TestRenderWorkspaceViewIncludesMetadataAndEmptyState(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	var buf bytes.Buffer
	renderWorkspaceView(&buf, &workspaceViewData{
		Workspace: domain.Workspace{
			ID:             "WS-1",
			DiskUsageBytes: 2048,
			LastModified:   time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
		},
		Status: &domain.WorkspaceStatus{
			ID:         "WS-1",
			BranchName: "main",
		},
		Path: "/tmp/workspaces/WS-1",
	})

	output := buf.String()
	for _, want := range []string{
		"Path",
		"Disk",
		"Modified",
		"No repositories in this workspace.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}
