# Tasks: Split Workspaces Service

## Implementation Checklist

### 1. Extract RepoResolver
- [ ] 1.1 Create `internal/workspaces/resolver.go`
- [ ] 1.2 Move `resolveRepoIdentifier()` and related functions
- [ ] 1.3 Move `isLikelyURL()` and `repoNameFromURL()`
- [ ] 1.4 Define `RepoResolver` struct with registry dependency
- [ ] 1.5 Update `Service` to use `RepoResolver`
- [ ] 1.6 Add unit tests for `RepoResolver`

### 2. Extract DiskUsageCalculator
- [ ] 2.1 Create `internal/workspaces/diskusage.go`
- [ ] 2.2 Move `cachedWorkspaceUsage()` and `CalculateDiskUsage()`
- [ ] 2.3 Move `usageEntry` type and `usageCache` map
- [ ] 2.4 Define `DiskUsageCalculator` struct
- [ ] 2.5 Update `Service` to use `DiskUsageCalculator`
- [ ] 2.6 Add unit tests for `DiskUsageCalculator`

### 3. Extract CanonicalRepoService
- [ ] 3.1 Create `internal/workspaces/canonical.go`
- [ ] 3.2 Move `ListCanonicalRepos()`, `AddCanonicalRepo()`, `RemoveCanonicalRepo()`, `SyncCanonicalRepo()`
- [ ] 3.3 Define `CanonicalRepoService` struct
- [ ] 3.4 Update `Service` to delegate to `CanonicalRepoService`
- [ ] 3.5 Add unit tests for `CanonicalRepoService`

### 4. Cleanup and Integration
- [ ] 4.1 Ensure `Service` constructor wires all sub-services
- [ ] 4.2 Update `internal/app/app.go` if needed
- [ ] 4.3 Run full test suite
- [ ] 4.4 Verify all CLI commands work correctly
