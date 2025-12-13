## MODIFIED Requirements

### Requirement: Hexagonal Architecture
The application SHALL follow hexagonal architecture principles:
- Core business logic resides in `internal/workspaces/`, `internal/workspace/`, `internal/domain/`
- External dependencies are abstracted through interfaces in `internal/ports/`
- Adapters implement port interfaces and are injected via dependency injection
- Services depend on interfaces, not concrete implementations

All injectable dependencies used by the service layer SHALL have corresponding interfaces in `internal/ports/`.

#### Scenario: Service depends on interface for hooks
- **GIVEN** the workspaces Service
- **WHEN** hook execution is needed
- **THEN** the Service SHALL use `ports.HookExecutor` interface
- **AND** any implementation satisfying the interface can be injected

#### Scenario: Service depends on interface for disk usage
- **GIVEN** the workspaces Service
- **WHEN** disk usage calculation is needed
- **THEN** the Service SHALL use `ports.DiskUsage` interface
- **AND** any implementation satisfying the interface can be injected

#### Scenario: Service depends on interface for caching
- **GIVEN** the workspaces Service
- **WHEN** workspace caching is needed
- **THEN** the Service SHALL use `ports.WorkspaceCache` interface
- **AND** any implementation satisfying the interface can be injected

#### Scenario: All port interfaces are mockable
- **GIVEN** any interface in `internal/ports/`
- **WHEN** writing unit tests
- **THEN** a mock implementation SHALL exist in `internal/mocks/`
- **AND** the mock SHALL be usable for testing without external dependencies

## ADDED Requirements

### Requirement: Complete Port Interface Coverage
Every injectable service dependency SHALL have a corresponding interface in `internal/ports/`.

#### Scenario: New dependency requires interface
- **GIVEN** a new dependency is added to the Service
- **WHEN** the dependency is injectable
- **THEN** an interface SHALL be created in `internal/ports/`
- **AND** a mock implementation SHALL be created in `internal/mocks/`

