# Change: Split Service Layer into Focused Sub-Services

## Why
The `internal/workspaces/service.go` file is 1211 lines with 45 methods, violating Single Responsibility Principle. Breaking it into focused sub-services will improve maintainability and testability.

## What Changes
- Extract `WorkspaceGitService` for git command coordination (Push, RunGit, Switch)
- Extract `WorkspaceOrphanService` for orphan detection and remediation
- Extract `WorkspaceExportService` for export/import functionality
- Keep core `Service` as a thin coordinator with workspace lifecycle (create, close, restore, list)
- Already extracted services remain unchanged: `CanonicalRepoService`, `RepoResolver`, `DiskUsageCalculator`, `WorkspaceCache`

**Proposed Structure:**
```
internal/workspaces/
├── service.go           # Coordinator (~300 lines) - lifecycle operations
├── git_service.go       # Git operations (~200 lines)
├── orphan_service.go    # Orphan detection (~150 lines)
├── export_service.go    # Export/import (~150 lines)
├── canonical.go         # (unchanged)
├── resolver.go          # (unchanged)
├── diskusage.go         # (unchanged)
├── cache.go             # (unchanged)
```

## Impact
- Affected specs: core-architecture
- Affected code:
  - `internal/workspaces/service.go` (major reduction)
  - New files: `git_service.go`, `orphan_service.go`, `export_service.go`
  - `internal/app/app.go` (service wiring)
  - CLI commands that call extracted methods
- No breaking changes to public API - main Service maintains all existing methods via delegation

