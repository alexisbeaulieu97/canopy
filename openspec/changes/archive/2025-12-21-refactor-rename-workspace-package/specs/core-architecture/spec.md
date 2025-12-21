## MODIFIED Requirements

### Requirement: Hexagonal Architecture
The codebase SHALL follow hexagonal architecture patterns.

#### Scenario: Port definitions
- **GIVEN** interfaces are defined in `internal/ports/`
- **WHEN** adapters implement these interfaces
- **THEN** the domain layer SHALL remain decoupled from infrastructure

#### Scenario: Adapter package naming
- **GIVEN** the hexagonal architecture separates ports from adapters
- **WHEN** a developer looks for the WorkspaceStorage implementation
- **THEN** they find it in `internal/storage/` package
- **AND** the package name reflects its role as a storage adapter
