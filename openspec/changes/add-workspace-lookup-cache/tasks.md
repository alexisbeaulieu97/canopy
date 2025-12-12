## 1. Add Direct Lookup to Storage Interface
- [ ] 1.1 Add `LoadByID(id string) (*domain.Workspace, string, error)` to `WorkspaceStorage` interface
- [ ] 1.2 Document the method returns (workspace, dirName, error)
- [ ] 1.3 Update mock implementations

## 2. Implement Direct Lookup in Engine
- [ ] 2.1 Implement `LoadByID` in `workspace.Engine`
- [ ] 2.2 Use direct path construction: `workspacesRoot/<id>/workspace.yaml`
- [ ] 2.3 Handle case where ID differs from directory name (scan fallback)
- [ ] 2.4 Add unit tests for direct lookup

## 3. Add Workspace Cache
- [ ] 3.1 Create `WorkspaceCache` struct with TTL support
- [ ] 3.2 Implement `Get(id string) (*domain.Workspace, bool)`
- [ ] 3.3 Implement `Set(id string, ws *domain.Workspace)`
- [ ] 3.4 Implement `Invalidate(id string)` and `InvalidateAll()`
- [ ] 3.5 Add configurable TTL (default 30 seconds)
- [ ] 3.6 Add unit tests for cache operations

## 4. Integrate Cache into Service
- [ ] 4.1 Add cache field to `Service` struct
- [ ] 4.2 Update `findWorkspace()` to check cache first
- [ ] 4.3 Invalidate cache on `CreateWorkspace`, `CloseWorkspace`, `Save`
- [ ] 4.4 Add cache stats for observability (optional)

## 5. Update Callers
- [ ] 5.1 Replace `findWorkspace()` calls with direct lookup where appropriate
- [ ] 5.2 Ensure cache invalidation in all write paths
- [ ] 5.3 Add integration tests for cache behavior

## 6. Configuration
- [ ] 6.1 Add `workspace_cache_ttl` config option (optional)
- [ ] 6.2 Add `disable_workspace_cache` for debugging (optional)
