# Tasks: Split Service Layer into Focused Sub-Services

## Implementation Checklist

### 1. Extract WorkspaceGitService
- [ ] 1.1 Create `internal/workspaces/git_service.go`
- [ ] 1.2 Move `PushWorkspace` method
- [ ] 1.3 Move `RunGitInWorkspace` method
- [ ] 1.4 Move `runGitSequential` method
- [ ] 1.5 Move `runGitParallel` method
- [ ] 1.6 Move `SwitchBranch` method
- [ ] 1.7 Add `GitService` interface to enable mocking
- [ ] 1.8 Create constructor `NewGitService`
- [ ] 1.9 Add delegation methods in main `Service`
- [ ] 1.10 Write unit tests for `WorkspaceGitService`

### 2. Extract WorkspaceOrphanService
- [ ] 2.1 Create `internal/workspaces/orphan_service.go`
- [ ] 2.2 Move `DetectOrphans` method
- [ ] 2.3 Move `DetectOrphansForWorkspace` method
- [ ] 2.4 Move `buildCanonicalRepoSet` method
- [ ] 2.5 Move `checkWorkspaceForOrphans` method
- [ ] 2.6 Move `checkRepoForOrphan` method
- [ ] 2.7 Move `logStatError` helper
- [ ] 2.8 Add `OrphanService` interface
- [ ] 2.9 Create constructor `NewOrphanService`
- [ ] 2.10 Add delegation methods in main `Service`
- [ ] 2.11 Write unit tests for `WorkspaceOrphanService`

### 3. Extract WorkspaceExportService
- [ ] 3.1 Create `internal/workspaces/export_service.go`
- [ ] 3.2 Move `ExportWorkspace` method
- [ ] 3.3 Move `ImportWorkspace` method
- [ ] 3.4 Move `resolveImportOverrides` method
- [ ] 3.5 Move `prepareForImport` method
- [ ] 3.6 Move `resolveExportedRepos` method
- [ ] 3.7 Add `ExportService` interface
- [ ] 3.8 Create constructor `NewExportService`
- [ ] 3.9 Add delegation methods in main `Service`
- [ ] 3.10 Write unit tests for `WorkspaceExportService`

### 4. Refactor Main Service
- [ ] 4.1 Remove extracted methods from `service.go`
- [ ] 4.2 Add sub-service dependencies to `Service` struct
- [ ] 4.3 Update `NewService` constructor to accept/create sub-services
- [ ] 4.4 Implement delegation pattern for extracted methods
- [ ] 4.5 Verify main `Service` is now ~300-400 lines

### 5. Update App Initialization
- [ ] 5.1 Update `internal/app/app.go` to wire sub-services
- [ ] 5.2 Ensure functional options work with new structure

### 6. Testing
- [ ] 6.1 Run existing test suite - all tests should pass
- [ ] 6.2 Verify test coverage for new service files
- [ ] 6.3 Run with race detector

### 7. Documentation
- [ ] 7.1 Update code comments explaining the service structure
- [ ] 7.2 Add package-level documentation if needed

