package workspaces

import (
	"testing"
	"time"

	"github.com/alexisbeaulieu97/canopy/internal/domain"
)

func TestWorkspaceCache_GetSet(t *testing.T) {
	t.Parallel()

	cache := NewWorkspaceCache(DefaultCacheTTL)

	// Test cache miss
	ws, dirName, ok := cache.Get("test-id")
	if ok {
		t.Error("expected cache miss for non-existent entry")
	}

	if ws != nil || dirName != "" {
		t.Error("expected nil values for cache miss")
	}

	// Add entry to cache
	testWs := &domain.Workspace{
		ID:         "test-id",
		BranchName: "main",
		Repos:      []domain.Repo{{Name: "repo1"}},
	}
	cache.Set("test-id", testWs, "test-dir")

	// Test cache hit
	ws, dirName, ok = cache.Get("test-id")
	if !ok {
		t.Error("expected cache hit")
	}

	if ws == nil {
		t.Fatal("expected non-nil workspace")
	}

	if ws.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %q", ws.ID)
	}

	if dirName != "test-dir" {
		t.Errorf("expected dirName 'test-dir', got %q", dirName)
	}

	if ws.BranchName != "main" {
		t.Errorf("expected BranchName 'main', got %q", ws.BranchName)
	}
}

func TestWorkspaceCache_TTLExpiration(t *testing.T) {
	t.Parallel()

	// Use a short TTL for testing, but long enough to be reliable in CI
	cache := NewWorkspaceCache(50 * time.Millisecond)

	testWs := &domain.Workspace{ID: "test-id"}
	cache.Set("test-id", testWs, "test-dir")

	// Immediately after setting, should be a hit
	_, _, ok := cache.Get("test-id")
	if !ok {
		t.Error("expected cache hit before TTL expiration")
	}

	// Wait for TTL to expire (sleep 3x the TTL to be safe)
	time.Sleep(150 * time.Millisecond)

	// After TTL, should be a miss
	_, _, ok = cache.Get("test-id")
	if ok {
		t.Error("expected cache miss after TTL expiration")
	}
}

func TestWorkspaceCache_Invalidate(t *testing.T) {
	t.Parallel()

	cache := NewWorkspaceCache(DefaultCacheTTL)

	// Add entries
	cache.Set("ws-1", &domain.Workspace{ID: "ws-1"}, "dir-1")
	cache.Set("ws-2", &domain.Workspace{ID: "ws-2"}, "dir-2")

	// Verify both exist
	_, _, ok1 := cache.Get("ws-1")
	_, _, ok2 := cache.Get("ws-2")

	if !ok1 || !ok2 {
		t.Error("expected both entries to exist before invalidation")
	}

	// Invalidate one entry
	cache.Invalidate("ws-1")

	// ws-1 should be gone, ws-2 should remain
	_, _, ok1 = cache.Get("ws-1")
	_, _, ok2 = cache.Get("ws-2")

	if ok1 {
		t.Error("expected ws-1 to be invalidated")
	}

	if !ok2 {
		t.Error("expected ws-2 to still exist")
	}
}

func TestWorkspaceCache_InvalidateAll(t *testing.T) {
	t.Parallel()

	cache := NewWorkspaceCache(DefaultCacheTTL)

	// Add multiple entries
	cache.Set("ws-1", &domain.Workspace{ID: "ws-1"}, "dir-1")
	cache.Set("ws-2", &domain.Workspace{ID: "ws-2"}, "dir-2")
	cache.Set("ws-3", &domain.Workspace{ID: "ws-3"}, "dir-3")

	if cache.Size() != 3 {
		t.Errorf("expected 3 entries, got %d", cache.Size())
	}

	// Invalidate all
	cache.InvalidateAll()

	if cache.Size() != 0 {
		t.Errorf("expected 0 entries after InvalidateAll, got %d", cache.Size())
	}

	// Verify all entries are gone
	for _, id := range []string{"ws-1", "ws-2", "ws-3"} {
		if _, _, ok := cache.Get(id); ok {
			t.Errorf("expected %s to be invalidated", id)
		}
	}
}

func TestWorkspaceCache_DefaultTTL(t *testing.T) {
	t.Parallel()

	// When TTL is 0, should use default
	cache := NewWorkspaceCache(0)

	if cache.ttl != DefaultCacheTTL {
		t.Errorf("expected default TTL %v, got %v", DefaultCacheTTL, cache.ttl)
	}
}

func TestWorkspaceCache_Size(t *testing.T) {
	t.Parallel()

	cache := NewWorkspaceCache(DefaultCacheTTL)

	if cache.Size() != 0 {
		t.Errorf("expected size 0, got %d", cache.Size())
	}

	cache.Set("ws-1", &domain.Workspace{ID: "ws-1"}, "dir-1")

	if cache.Size() != 1 {
		t.Errorf("expected size 1, got %d", cache.Size())
	}

	cache.Set("ws-2", &domain.Workspace{ID: "ws-2"}, "dir-2")

	if cache.Size() != 2 {
		t.Errorf("expected size 2, got %d", cache.Size())
	}

	// Update existing entry shouldn't increase size
	cache.Set("ws-1", &domain.Workspace{ID: "ws-1", BranchName: "updated"}, "dir-1")

	if cache.Size() != 2 {
		t.Errorf("expected size 2 after update, got %d", cache.Size())
	}
}

func TestWorkspaceCache_Concurrency(t *testing.T) {
	t.Parallel()

	cache := NewWorkspaceCache(DefaultCacheTTL)

	// Run concurrent operations to test thread safety
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			wsID := "ws-" + string(rune('a'+id))
			ws := &domain.Workspace{ID: wsID}

			// Perform multiple operations
			for j := 0; j < 100; j++ {
				cache.Set(wsID, ws, "dir-"+wsID)
				cache.Get(wsID)
				cache.Size()
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without a race condition or deadlock, the test passes
}

func TestWorkspaceCache_LazyDeletion(t *testing.T) {
	t.Parallel()

	cache := NewWorkspaceCache(50 * time.Millisecond)

	testWs := &domain.Workspace{ID: "test-id"}
	cache.Set("test-id", testWs, "test-dir")

	// Entry should be in cache
	if cache.Size() != 1 {
		t.Errorf("expected size 1, got %d", cache.Size())
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Get should trigger lazy deletion
	_, _, ok := cache.Get("test-id")
	if ok {
		t.Error("expected cache miss after TTL expiration")
	}

	// Entry should be deleted from cache
	if cache.Size() != 0 {
		t.Errorf("expected size 0 after lazy deletion, got %d", cache.Size())
	}
}

func TestWorkspaceCache_DeepCopy(t *testing.T) {
	t.Parallel()

	cache := NewWorkspaceCache(DefaultCacheTTL)

	// Create workspace with repos
	original := &domain.Workspace{
		ID:         "test-id",
		BranchName: "main",
		Repos: []domain.Repo{
			{Name: "repo1", URL: "https://github.com/org/repo1.git"},
		},
	}
	cache.Set("test-id", original, "test-dir")

	// Get the cached workspace
	ws1, _, ok := cache.Get("test-id")
	if !ok {
		t.Fatal("expected cache hit")
	}

	// Modify the returned workspace
	ws1.BranchName = "modified"
	ws1.Repos[0].Name = "modified-repo"
	ws1.Repos = append(ws1.Repos, domain.Repo{Name: "new-repo"})

	// Get another copy - should be unaffected by previous modifications
	ws2, _, ok := cache.Get("test-id")
	if !ok {
		t.Fatal("expected cache hit")
	}

	if ws2.BranchName != "main" {
		t.Errorf("expected BranchName 'main', got %q (cache was mutated)", ws2.BranchName)
	}

	if len(ws2.Repos) != 1 {
		t.Errorf("expected 1 repo, got %d (cache was mutated)", len(ws2.Repos))
	}

	if ws2.Repos[0].Name != "repo1" {
		t.Errorf("expected repo name 'repo1', got %q (cache was mutated)", ws2.Repos[0].Name)
	}

	// Also verify original wasn't mutated
	if original.BranchName != "main" {
		t.Errorf("original workspace was mutated: BranchName = %q", original.BranchName)
	}
}
