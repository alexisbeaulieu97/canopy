package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

// WorkspaceModel manages workspace data and caches.
type WorkspaceModel struct {
	// allItems contains all loaded workspace items (unfiltered).
	allItems []workspaceItem
	// statusCache maps workspace ID to its status.
	statusCache map[string]*domain.WorkspaceStatus
	// totalDiskUsage is the sum of disk usage across all workspaces.
	totalDiskUsage int64
	// filterStale indicates whether to filter to only stale workspaces.
	filterStale bool
	// staleThresholdDays is the number of days before a workspace is considered stale.
	staleThresholdDays int
}

// NewWorkspaceModel creates a new WorkspaceModel with the given stale threshold.
func NewWorkspaceModel(staleThresholdDays int) *WorkspaceModel {
	return &WorkspaceModel{
		statusCache:        make(map[string]*domain.WorkspaceStatus),
		staleThresholdDays: staleThresholdDays,
	}
}

// SetItems sets all workspace items and total disk usage.
func (wm *WorkspaceModel) SetItems(items []workspaceItem, totalUsage int64) {
	wm.allItems = items
	wm.totalDiskUsage = totalUsage
}

// Items returns all workspace items.
func (wm *WorkspaceModel) Items() []workspaceItem {
	return wm.allItems
}

// TotalDiskUsage returns the total disk usage across all workspaces.
func (wm *WorkspaceModel) TotalDiskUsage() int64 {
	return wm.totalDiskUsage
}

// CacheStatus stores a workspace status in the cache.
func (wm *WorkspaceModel) CacheStatus(id string, status *domain.WorkspaceStatus) {
	wm.statusCache[id] = status
}

// GetCachedStatus retrieves a cached workspace status.
func (wm *WorkspaceModel) GetCachedStatus(id string) (*domain.WorkspaceStatus, bool) {
	status, ok := wm.statusCache[id]
	return status, ok
}

// ToggleStaleFilter toggles the stale filter on/off.
func (wm *WorkspaceModel) ToggleStaleFilter() {
	wm.filterStale = !wm.filterStale
}

// IsStaleFilterActive returns whether the stale filter is active.
func (wm *WorkspaceModel) IsStaleFilterActive() bool {
	return wm.filterStale
}

// StaleThresholdDays returns the stale threshold in days.
func (wm *WorkspaceModel) StaleThresholdDays() int {
	return wm.staleThresholdDays
}

// UpdateItemSummary updates the summary for a workspace item.
func (wm *WorkspaceModel) UpdateItemSummary(id string, status *domain.WorkspaceStatus, err error) {
	for idx, it := range wm.allItems {
		if it.Workspace.ID != id {
			continue
		}

		if status != nil {
			it.Loaded = true
			it.Err = nil
			it.Summary = summarizeStatus(status)
		}

		if err != nil {
			it.Err = err
		}

		wm.allItems[idx] = it
	}
}

// FindItemByID finds a workspace item by its ID.
func (wm *WorkspaceModel) FindItemByID(id string) (workspaceItem, bool) {
	for _, it := range wm.allItems {
		if it.Workspace.ID == id {
			return it, true
		}
	}

	return workspaceItem{}, false
}

// ApplyFilters returns filtered list items based on current filters and search value.
func (wm *WorkspaceModel) ApplyFilters(searchValue string) []list.Item {
	var items []list.Item

	search := strings.ToLower(strings.TrimSpace(searchValue))

	for _, it := range wm.allItems {
		if wm.filterStale && !it.Workspace.IsStale(wm.staleThresholdDays) {
			continue
		}

		if search != "" && !strings.Contains(strings.ToLower(it.Workspace.ID), search) {
			continue
		}

		items = append(items, it)
	}

	return items
}
