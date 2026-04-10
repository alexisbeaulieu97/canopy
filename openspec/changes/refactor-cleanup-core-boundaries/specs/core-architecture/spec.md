## MODIFIED Requirements
### Requirement: Hexagonal Architecture with Port Interfaces
The system SHALL use hexagonal architecture with interface-defined ports to decouple the service layer from infrastructure implementations. Port contracts SHALL be owned by `internal/ports` and SHALL NOT require consumers in core packages to depend on concrete adapter package types.

#### Scenario: Service depends on port-owned configuration contracts
- **WHEN** the Service or other core packages consume configuration-derived values
- **THEN** they depend on interfaces or DTOs defined in `internal/ports`
- **AND** they do not import concrete `internal/config` types to perform core operations

#### Scenario: Adapter packages convert to port-owned DTOs
- **WHEN** configuration is loaded from YAML
- **THEN** `internal/config` remains responsible for parsing and validation
- **AND** outward-facing values exposed through `ConfigProvider` are converted to port-owned DTOs or interfaces

### Requirement: Centralized Service Initialization
The system SHALL initialize all services through a centralized App struct that manages dependencies and lifecycle. The App struct SHALL support functional options for injecting custom implementations.

#### Scenario: App wires concrete hook execution
- **WHEN** `app.New` creates the default application graph
- **THEN** concrete hook execution is constructed in the app layer
- **AND** the workspace service receives that executor through dependency injection
- **AND** core packages do not instantiate hook adapters directly

### Requirement: Hexagonal Architecture
The codebase SHALL follow hexagonal architecture patterns.

#### Scenario: Core package does not own shell-execution policy
- **GIVEN** shell command execution is an infrastructure concern
- **WHEN** workspace template setup commands are executed
- **THEN** execution flows through an injected adapter contract
- **AND** the workspace core package does not directly select shells or spawn commands for that concern
