## MODIFIED Requirements

### Requirement: Single Responsibility Service Components
Service components SHALL follow the Single Responsibility Principle, with each service focused on a cohesive set of operations.

The workspace service layer SHALL be organized as follows:
- `Service` - Coordinator for workspace lifecycle (create, close, restore, list, status)
- `WorkspaceGitService` - Git command execution across workspace repos
- `WorkspaceOrphanService` - Orphan worktree detection and remediation
- `WorkspaceExportService` - Workspace export/import functionality
- `CanonicalRepoService` - Canonical repository management (existing)
- `RepoResolver` - Repository identifier resolution (existing)
- `DiskUsageCalculator` - Disk usage calculation (existing)
- `WorkspaceCache` - Workspace lookup caching (existing)

#### Scenario: Main service coordinates sub-services
- **GIVEN** a workspace operation that spans multiple concerns
- **WHEN** the operation is invoked on the main Service
- **THEN** the Service SHALL delegate to appropriate sub-services
- **AND** the public API SHALL remain unchanged

#### Scenario: Sub-services are independently testable
- **GIVEN** a sub-service like WorkspaceGitService
- **WHEN** unit tests are written
- **THEN** the sub-service SHALL be testable without instantiating the full Service
- **AND** dependencies SHALL be injectable via interfaces

#### Scenario: Service file size is manageable
- **GIVEN** the service layer structure
- **WHEN** any single service file is examined
- **THEN** it SHALL contain fewer than 500 lines of code
- **AND** it SHALL have a clear, single purpose

## ADDED Requirements

### Requirement: Service Delegation Pattern
The main `Service` struct SHALL maintain backward compatibility by delegating to sub-services for extracted functionality.

#### Scenario: Existing method calls work unchanged
- **GIVEN** code that calls `service.PushWorkspace()`
- **WHEN** the method is invoked after refactoring
- **THEN** it SHALL delegate to `WorkspaceGitService.Push()`
- **AND** the behavior SHALL be identical to before refactoring

