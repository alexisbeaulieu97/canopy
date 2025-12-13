package tui

import (
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestWorkspaceModel_SetAndGetItems(t *testing.T) {
	wm := NewWorkspaceModel(30)

	items := []workspaceItem{
		{Workspace: domain.Workspace{ID: "ws1"}},
		{Workspace: domain.Workspace{ID: "ws2"}},
	}

	wm.SetItems(items, 1024)

	if got := len(wm.Items()); got != 2 {
		t.Errorf("Items() = %d items, want 2", got)
	}

	if got := wm.TotalDiskUsage(); got != 1024 {
		t.Errorf("TotalDiskUsage() = %d, want 1024", got)
	}
}

func TestWorkspaceModel_StatusCache(t *testing.T) {
	wm := NewWorkspaceModel(30)

	status := &domain.WorkspaceStatus{
		Repos: []domain.RepoStatus{{Name: "repo1"}},
	}

	wm.CacheStatus("ws1", status)

	got, ok := wm.GetCachedStatus("ws1")
	if !ok {
		t.Fatal("GetCachedStatus() returned false, want true")
	}

	if got != status {
		t.Error("GetCachedStatus() returned wrong status")
	}

	_, ok = wm.GetCachedStatus("nonexistent")
	if ok {
		t.Error("GetCachedStatus() returned true for nonexistent key, want false")
	}
}

func TestWorkspaceModel_StaleFilter(t *testing.T) {
	wm := NewWorkspaceModel(30)

	if wm.IsStaleFilterActive() {
		t.Error("IsStaleFilterActive() = true, want false initially")
	}

	wm.ToggleStaleFilter()

	if !wm.IsStaleFilterActive() {
		t.Error("IsStaleFilterActive() = false after toggle, want true")
	}

	wm.ToggleStaleFilter()

	if wm.IsStaleFilterActive() {
		t.Error("IsStaleFilterActive() = true after second toggle, want false")
	}
}

func TestWorkspaceModel_StaleThreshold(t *testing.T) {
	wm := NewWorkspaceModel(45)

	if got := wm.StaleThresholdDays(); got != 45 {
		t.Errorf("StaleThresholdDays() = %d, want 45", got)
	}
}

func TestWorkspaceModel_FindItemByID(t *testing.T) {
	wm := NewWorkspaceModel(30)

	items := []workspaceItem{
		{Workspace: domain.Workspace{ID: "ws1"}},
		{Workspace: domain.Workspace{ID: "ws2"}},
	}

	wm.SetItems(items, 0)

	item, ok := wm.FindItemByID("ws1")
	if !ok {
		t.Fatal("FindItemByID() returned false for existing item")
	}

	if item.Workspace.ID != "ws1" {
		t.Errorf("FindItemByID() returned wrong item, got ID %s", item.Workspace.ID)
	}

	_, ok = wm.FindItemByID("nonexistent")
	if ok {
		t.Error("FindItemByID() returned true for nonexistent item, want false")
	}
}

func TestWorkspaceModel_UpdateItemSummary(t *testing.T) {
	wm := NewWorkspaceModel(30)

	items := []workspaceItem{
		{Workspace: domain.Workspace{ID: "ws1"}},
		{Workspace: domain.Workspace{ID: "ws2"}},
	}

	wm.SetItems(items, 0)

	status := &domain.WorkspaceStatus{
		Repos: []domain.RepoStatus{
			{Name: "repo1", IsDirty: true},
		},
	}

	wm.UpdateItemSummary("ws1", status, nil)

	item, _ := wm.FindItemByID("ws1")
	if !item.Loaded {
		t.Error("UpdateItemSummary() did not set Loaded = true")
	}
}

func TestWorkspaceModel_ApplyFilters_Search(t *testing.T) {
	wm := NewWorkspaceModel(30)

	items := []workspaceItem{
		{Workspace: domain.Workspace{ID: "project-alpha"}},
		{Workspace: domain.Workspace{ID: "project-beta"}},
		{Workspace: domain.Workspace{ID: "something-else"}},
	}

	wm.SetItems(items, 0)

	// No filter
	result := wm.ApplyFilters("")
	if len(result) != 3 {
		t.Errorf("ApplyFilters('') = %d items, want 3", len(result))
	}

	// Search filter
	result = wm.ApplyFilters("alpha")
	if len(result) != 1 {
		t.Errorf("ApplyFilters('alpha') = %d items, want 1", len(result))
	}

	// Case insensitive search
	result = wm.ApplyFilters("PROJECT")
	if len(result) != 2 {
		t.Errorf("ApplyFilters('PROJECT') = %d items, want 2", len(result))
	}
}

func TestWorkspaceModel_ApplyFilters_Stale(t *testing.T) {
	wm := NewWorkspaceModel(30)

	recentTime := time.Now()
	staleTime := time.Now().AddDate(0, 0, -60) // 60 days ago

	items := []workspaceItem{
		{Workspace: domain.Workspace{ID: "recent", LastModified: recentTime}},
		{Workspace: domain.Workspace{ID: "stale", LastModified: staleTime}},
	}

	wm.SetItems(items, 0)

	// Without stale filter
	result := wm.ApplyFilters("")
	if len(result) != 2 {
		t.Errorf("ApplyFilters('') without stale filter = %d items, want 2", len(result))
	}

	// With stale filter
	wm.ToggleStaleFilter()
	result = wm.ApplyFilters("")
	if len(result) != 1 {
		t.Errorf("ApplyFilters('') with stale filter = %d items, want 1", len(result))
	}

	// Verify it's the stale workspace
	if result[0].(workspaceItem).Workspace.ID != "stale" {
		t.Error("ApplyFilters() with stale filter returned wrong item")
	}
}
