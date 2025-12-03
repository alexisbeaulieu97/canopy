## ADDED Requirements

### Requirement: Interface-Based Dependencies
Core services SHALL depend on interfaces rather than concrete implementations.

#### Scenario: Git operations via interface
- **GIVEN** the Service depends on GitOperations interface
- **WHEN** tests provide a mock implementation
- **THEN** tests SHALL run without real git operations

#### Scenario: Workspace storage via interface
- **GIVEN** the Service depends on WorkspaceStorage interface
- **WHEN** tests provide a mock implementation
- **THEN** tests SHALL run without filesystem access

### Requirement: Hexagonal Architecture
The codebase SHALL follow hexagonal architecture patterns.

#### Scenario: Port definitions
- **GIVEN** interfaces are defined in `internal/ports/`
- **WHEN** adapters implement these interfaces
- **THEN** the domain layer SHALL remain decoupled from infrastructure
