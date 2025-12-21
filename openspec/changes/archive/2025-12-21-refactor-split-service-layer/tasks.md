# Tasks: Split Service Layer into Focused Sub-Services

## Implementation Checklist

### 1. Extract WorkspaceGitService
- [x] 1.1 Create `internal/workspaces/git_service.go`
- [x] 1.2 Move `PushWorkspace` method
- [x] 1.3 Move `RunGitInWorkspace` method
- [x] 1.4 Move `runGitSequential` method
- [x] 1.5 Move `runGitParallel` method
- [x] 1.6 Move `SwitchBranch` method
- [x] 1.7 Add `GitService` interface to enable mocking
- [x] 1.8 Create constructor `NewGitService`
- [x] 1.9 Add delegation methods in main `Service`
- [x] 1.10 Write unit tests for `WorkspaceGitService`

### 2. Extract WorkspaceOrphanService
- [x] 2.1 Create `internal/workspaces/orphan_service.go`
- [x] 2.2 Move `DetectOrphans` method
- [x] 2.3 Move `DetectOrphansForWorkspace` method
- [x] 2.4 Move `buildCanonicalRepoSet` method
- [x] 2.5 Move `checkWorkspaceForOrphans` method
- [x] 2.6 Move `checkRepoForOrphan` method
- [x] 2.7 Move `logStatError` helper
- [x] 2.8 Add `OrphanService` interface
- [x] 2.9 Create constructor `NewOrphanService`
- [x] 2.10 Add delegation methods in main `Service`
- [x] 2.11 Write unit tests for `WorkspaceOrphanService`

### 3. Extract WorkspaceExportService
- [x] 3.1 Create `internal/workspaces/export_service.go`
- [x] 3.2 Move `ExportWorkspace` method
- [x] 3.3 Move `ImportWorkspace` method
- [x] 3.4 Move `resolveImportOverrides` method
- [x] 3.5 Move `prepareForImport` method
- [x] 3.6 Move `resolveExportedRepos` method
- [x] 3.7 Add `ExportService` interface
- [x] 3.8 Create constructor `NewExportService`
- [x] 3.9 Add delegation methods in main `Service`
- [x] 3.10 Write unit tests for `WorkspaceExportService`

### 4. Refactor Main Service
- [x] 4.1 Remove extracted methods from `service.go`
- [x] 4.2 Add sub-service dependencies to `Service` struct
- [x] 4.3 Update `NewService` constructor to accept/create sub-services
- [x] 4.4 Implement delegation pattern for extracted methods
- [x] 4.5 Verify main `Service` is now ~300-400 lines (reduced from 1506 to 1075 lines)

### 5. Update App Initialization
- [x] 5.1 Update `internal/app/app.go` to wire sub-services (sub-services are created internally by NewService)
- [x] 5.2 Ensure functional options work with new structure

### 6. Testing
- [x] 6.1 Run existing test suite - all tests should pass
- [x] 6.2 Verify test coverage for new service files
- [x] 6.3 Run with race detector

### 7. Documentation
- [x] 7.1 Update code comments explaining the service structure
- [x] 7.2 Add package-level documentation if needed

