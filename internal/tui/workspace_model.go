package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
	"github.com/alexisbeaulieu97/canopy/internal/tui/components"
)

// workspaceModel manages workspace data and caches.
type workspaceModel struct {
	// allItems contains all loaded workspace items (unfiltered).
	allItems []components.WorkspaceItem
	// statusCache maps workspace ID to its status.
	statusCache map[string]*domain.WorkspaceStatus
	// totalDiskUsage is the sum of disk usage across all workspaces.
	totalDiskUsage int64
	// filterStale indicates whether to filter to only stale workspaces.
	filterStale bool
	// staleThresholdDays is the number of days before a workspace is considered stale.
	staleThresholdDays int
}

// newWorkspaceModel creates a new workspaceModel with the given stale threshold.
func newWorkspaceModel(staleThresholdDays int) *workspaceModel {
	return &workspaceModel{
		statusCache:        make(map[string]*domain.WorkspaceStatus),
		staleThresholdDays: staleThresholdDays,
	}
}

// SetItems sets all workspace items and total disk usage.
func (wm *workspaceModel) SetItems(items []components.WorkspaceItem, totalUsage int64) {
	wm.allItems = items
	wm.totalDiskUsage = totalUsage
}

// Items returns all workspace items.
func (wm *workspaceModel) Items() []components.WorkspaceItem {
	return wm.allItems
}

// TotalDiskUsage returns the total disk usage across all workspaces.
func (wm *workspaceModel) TotalDiskUsage() int64 {
	return wm.totalDiskUsage
}

// CacheStatus stores a workspace status in the cache.
func (wm *workspaceModel) CacheStatus(id string, status *domain.WorkspaceStatus) {
	wm.statusCache[id] = status
}

// GetCachedStatus retrieves a cached workspace status.
func (wm *workspaceModel) GetCachedStatus(id string) (*domain.WorkspaceStatus, bool) {
	status, ok := wm.statusCache[id]
	return status, ok
}

// ToggleStaleFilter toggles the stale filter on/off.
func (wm *workspaceModel) ToggleStaleFilter() {
	wm.filterStale = !wm.filterStale
}

// IsStaleFilterActive returns whether the stale filter is active.
func (wm *workspaceModel) IsStaleFilterActive() bool {
	return wm.filterStale
}

// StaleThresholdDays returns the stale threshold in days.
func (wm *workspaceModel) StaleThresholdDays() int {
	return wm.staleThresholdDays
}

// UpdateItemSummary updates the summary for a workspace item.
// If err is non-nil, the item is marked as having an error.
// If status is non-nil (and err is nil), the item is marked as loaded with its summary.
func (wm *workspaceModel) UpdateItemSummary(id string, status *domain.WorkspaceStatus, err error) {
	for idx, it := range wm.allItems {
		if it.Workspace.ID != id {
			continue
		}

		// Error and status handling are mutually exclusive.
		if err != nil {
			it.Err = err
			// Don't mark as loaded on error - keep previous state.
		} else if status != nil {
			it.Loaded = true
			it.Err = nil
			it.Summary = components.SummarizeStatus(status)
		}

		wm.allItems[idx] = it

		return // Only one item can match; exit early.
	}
}

// FindItemByID finds a workspace item by its ID.
func (wm *workspaceModel) FindItemByID(id string) (components.WorkspaceItem, bool) {
	for _, it := range wm.allItems {
		if it.Workspace.ID == id {
			return it, true
		}
	}

	return components.WorkspaceItem{}, false
}

// ApplyFilters returns filtered list items based on current filters and search value.
func (wm *workspaceModel) ApplyFilters(searchValue string) []list.Item {
	var items []list.Item

	search := strings.ToLower(strings.TrimSpace(searchValue))

	for _, it := range wm.allItems {
		if wm.filterStale && !it.Workspace.IsStale(wm.staleThresholdDays) {
			continue
		}

		if search != "" && !workspaceMatchesSearch(it, wm.statusCache[it.Workspace.ID], search, wm.staleThresholdDays) {
			continue
		}

		items = append(items, it)
	}

	return items
}

func workspaceMatchesSearch(
	item components.WorkspaceItem,
	status *domain.WorkspaceStatus,
	search string,
	staleThresholdDays int,
) bool {
	matchers := workspaceSearchTerms(item, status, staleThresholdDays)

	return containsSearchTerm(matchers, search)
}

func workspaceSearchTerms(
	item components.WorkspaceItem,
	status *domain.WorkspaceStatus,
	staleThresholdDays int,
) []string {
	matchers := []string{
		normalizeSearchTerm(item.Workspace.ID),
		normalizeSearchTerm(item.Workspace.BranchName),
	}

	matchers = append(matchers, workspaceStateSearchTerms(item, staleThresholdDays)...)
	matchers = append(matchers, workspaceRepoNames(item)...)
	matchers = append(matchers, cachedStatusSearchTerms(status)...)

	return matchers
}

func workspaceStateSearchTerms(item components.WorkspaceItem, staleThresholdDays int) []string {
	var terms []string

	terms = append(terms, workspaceFlagTerms(item, staleThresholdDays)...)
	terms = append(terms, workspaceSummaryTerms(item)...)

	return terms
}

func workspaceFlagTerms(item components.WorkspaceItem, staleThresholdDays int) []string {
	var terms []string

	if item.Workspace.Locked {
		terms = append(terms, normalizeSearchTerm("locked"))
	}

	if item.Workspace.IsStale(staleThresholdDays) {
		terms = append(terms, normalizeSearchTerm("stale"))
	}

	if item.OrphanCount > 0 || item.OrphanCheckFailed {
		terms = append(terms, normalizeSearchTerm("orphan"))
	}

	if item.Err != nil {
		terms = append(terms, normalizeSearchTerm("error"), normalizeSearchTerm(item.Err.Error()))
	}

	if item.LockCheckFailed {
		terms = append(terms, normalizeSearchTerm("lock"), normalizeSearchTerm("error"))
	}

	return terms
}

func workspaceSummaryTerms(item components.WorkspaceItem) []string {
	var terms []string

	if item.Summary.DirtyRepos > 0 {
		terms = append(terms, normalizeSearchTerm("dirty"))
	}

	if item.Summary.UnpushedRepos > 0 {
		terms = append(terms, normalizeSearchTerm("unpushed"))
	}

	if item.Summary.BehindRepos > 0 {
		terms = append(terms, normalizeSearchTerm("behind"))
	}

	if item.Summary.ErrorRepos > 0 || item.OrphanCheckFailed || item.LockCheckFailed {
		terms = append(terms, normalizeSearchTerm("error"))
	}

	return terms
}

func workspaceRepoNames(item components.WorkspaceItem) []string {
	terms := make([]string, 0, len(item.Workspace.Repos))
	for _, repo := range item.Workspace.Repos {
		terms = append(terms, normalizeSearchTerm(repo.Name))
	}

	return terms
}

func cachedStatusSearchTerms(status *domain.WorkspaceStatus) []string {
	if status == nil {
		return nil
	}

	terms := []string{normalizeSearchTerm(status.BranchName)}
	for _, repo := range status.Repos {
		terms = append(terms, normalizeSearchTerm(repo.Name), normalizeSearchTerm(repo.Branch))
		terms = append(terms, repoStatusKeywords(repo)...)
	}

	return terms
}

func repoStatusKeywords(repo domain.RepoStatus) []string {
	var terms []string

	if repo.IsDirty {
		terms = append(terms, normalizeSearchTerm("dirty"))
	}

	if repo.UnpushedCommits > 0 {
		terms = append(terms, normalizeSearchTerm("unpushed"))
	}

	if repo.BehindRemote > 0 {
		terms = append(terms, normalizeSearchTerm("behind"))
	}

	if repo.Error != "" {
		terms = append(terms, normalizeSearchTerm("error"), normalizeSearchTerm(string(repo.Error)))
	}

	return terms
}

func containsSearchTerm(matchers []string, search string) bool {
	search = normalizeSearchTerm(search)

	for _, matcher := range matchers {
		if strings.Contains(matcher, search) {
			return true
		}
	}

	return false
}

func normalizeSearchTerm(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
