# Tasks: Split Workspaces Service

## Implementation Checklist

### 1. Extract RepoResolver
- [x] 1.1 Create `internal/workspaces/resolver.go`
- [x] 1.2 Move `resolveRepoIdentifier()` and related functions
- [x] 1.3 Move `isLikelyURL()` and `repoNameFromURL()`
- [x] 1.4 Define `RepoResolver` struct with registry dependency
- [x] 1.5 Update `Service` to use `RepoResolver`
- [x] 1.6 Add unit tests for `RepoResolver`

### 2. Extract DiskUsageCalculator
- [x] 2.1 Create `internal/workspaces/diskusage.go`
- [x] 2.2 Move `cachedWorkspaceUsage()` and `CalculateDiskUsage()`
- [x] 2.3 Move `usageEntry` type and `usageCache` map
- [x] 2.4 Define `DiskUsageCalculator` struct
- [x] 2.5 Update `Service` to use `DiskUsageCalculator`
- [x] 2.6 Add unit tests for `DiskUsageCalculator`

### 3. Extract CanonicalRepoService
- [x] 3.1 Create `internal/workspaces/canonical.go`
- [x] 3.2 Move `ListCanonicalRepos()`, `AddCanonicalRepo()`, `RemoveCanonicalRepo()`, `SyncCanonicalRepo()`
- [x] 3.3 Define `CanonicalRepoService` struct
- [x] 3.4 Update `Service` to delegate to `CanonicalRepoService`
- [x] 3.5 Add unit tests for `CanonicalRepoService`

### 4. Cleanup and Integration
- [x] 4.1 Ensure `Service` constructor wires all sub-services
- [x] 4.2 Update `internal/app/app.go` if needed
- [x] 4.3 Run full test suite
- [x] 4.4 Verify all CLI commands work correctly
