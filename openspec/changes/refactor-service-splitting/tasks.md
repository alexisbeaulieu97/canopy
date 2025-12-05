# Tasks: Split Workspaces Service

## Implementation Checklist

### Phase 1: Extract RepoResolver
- [ ] Create `internal/workspaces/resolver.go`
- [ ] Move `resolveRepoIdentifier()` and related functions
- [ ] Move `isLikelyURL()` and `repoNameFromURL()`
- [ ] Define `RepoResolver` struct with registry dependency
- [ ] Update `Service` to use `RepoResolver`
- [ ] Add unit tests for `RepoResolver`

### Phase 2: Extract DiskUsageCalculator
- [ ] Create `internal/workspaces/diskusage.go`
- [ ] Move `cachedWorkspaceUsage()` and `CalculateDiskUsage()`
- [ ] Move `usageEntry` type and `usageCache` map
- [ ] Define `DiskUsageCalculator` struct
- [ ] Update `Service` to use `DiskUsageCalculator`
- [ ] Add unit tests for `DiskUsageCalculator`

### Phase 3: Extract CanonicalRepoService
- [ ] Create `internal/workspaces/canonical.go`
- [ ] Move `ListCanonicalRepos()`, `AddCanonicalRepo()`, `RemoveCanonicalRepo()`, `SyncCanonicalRepo()`
- [ ] Define `CanonicalRepoService` struct
- [ ] Update `Service` to delegate to `CanonicalRepoService`
- [ ] Add unit tests

### Phase 4: Cleanup and Integration
- [ ] Ensure `Service` constructor wires all sub-services
- [ ] Update `internal/app/app.go` if needed
- [ ] Run full test suite
- [ ] Verify all CLI commands work correctly
