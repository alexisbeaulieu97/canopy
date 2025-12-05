# Change: Split Workspaces Service

## Why
`internal/workspaces/service.go` is 801 lines and handles too many responsibilities:
- Workspace CRUD operations
- Git operations orchestration
- Disk usage calculation and caching
- Repository resolution from aliases/URLs
- Canonical repo management

This violates single responsibility principle and makes the code hard to navigate, test, and extend. Splitting into focused services improves maintainability.

## What Changes
- Extract `RepoResolver` for alias/URL resolution logic (currently ~100 lines)
- Extract `DiskUsageCalculator` for filesystem stats and caching (currently ~80 lines)
- Extract `CanonicalRepoService` for canonical repo operations (currently ~60 lines)
- Keep `WorkspaceService` focused on workspace lifecycle only
- Wire services together via constructor injection

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/workspaces/service.go` - Split into multiple files
  - `internal/workspaces/resolver.go` - New file for repo resolution
  - `internal/workspaces/diskusage.go` - New file for disk calculations
  - `internal/workspaces/canonical.go` - New file for canonical repos
  - `internal/app/app.go` - Wire new services
- **Risk**: Medium - Internal refactoring but touches core service
