package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
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

func TestRenderWorkspaceViewHandlesNilStatus(t *testing.T) {
	t.Setenv("CANOPY_COLOR", "0")

	var buf bytes.Buffer
	renderWorkspaceView(&buf, &workspaceViewData{
		Workspace: domain.Workspace{
			ID:             "WS-1",
			DiskUsageBytes: 2048,
			LastModified:   time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
		},
		Path: "/tmp/workspaces/WS-1",
	})

	output := buf.String()
	for _, want := range []string{
		"Branch",
		"unknown",
		"Status unavailable for this workspace.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

type stubWorkspaceViewService struct {
	workspaces []domain.Workspace
	statusErr  error
	lockErr    error
	orphanErr  error
}

func (s stubWorkspaceViewService) GetStatus(context.Context, string) (*domain.WorkspaceStatus, error) {
	if s.statusErr != nil {
		return nil, s.statusErr
	}

	return &domain.WorkspaceStatus{ID: "WS-1", BranchName: "main"}, nil
}

func (s stubWorkspaceViewService) ListWorkspaces(context.Context) ([]domain.Workspace, error) {
	return s.workspaces, nil
}

func (s stubWorkspaceViewService) WorkspacePath(context.Context, string) (string, error) {
	return "/tmp/workspaces/WS-1", nil
}

func (s stubWorkspaceViewService) WorkspaceLocked(string) (bool, error) {
	return false, s.lockErr
}

func (s stubWorkspaceViewService) DetectOrphansForWorkspace(context.Context, string) ([]domain.OrphanedWorktree, error) {
	return nil, s.orphanErr
}

func TestLoadWorkspaceViewDataChecksExistenceBeforeStatus(t *testing.T) {
	t.Parallel()

	service := stubWorkspaceViewService{
		workspaces: nil,
		statusErr:  errors.New("lower-level status error"),
	}

	_, err := loadWorkspaceViewData(context.Background(), service, "MISSING")
	if err == nil {
		t.Fatal("expected missing workspace error")
	}

	var canopyErr *cerrors.CanopyError
	if !errors.As(err, &canopyErr) || canopyErr.Code != cerrors.ErrWorkspaceNotFound {
		t.Fatalf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestLoadWorkspaceViewDataSurfacesWarnings(t *testing.T) {
	t.Parallel()

	viewData, err := loadWorkspaceViewData(context.Background(), stubWorkspaceViewService{
		workspaces: []domain.Workspace{{ID: "WS-1"}},
		lockErr:    errors.New("lock failed"),
		orphanErr:  errors.New("orphans failed"),
	}, "WS-1")
	if err != nil {
		t.Fatalf("loadWorkspaceViewData failed: %v", err)
	}

	if len(viewData.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %+v", viewData.Warnings)
	}
}
