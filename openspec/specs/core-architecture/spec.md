# core-architecture Specification

## Purpose
Defines the core architecture patterns for Canopy including service initialization, dependency injection via App context, command registration, hexagonal architecture, and project branding conventions.
## Requirements
### Requirement: Hexagonal Architecture with Port Interfaces
The system SHALL use hexagonal architecture with interface-defined ports to decouple the service layer from infrastructure implementations.

#### Scenario: Service depends on interfaces
- **WHEN** the Service struct is initialized
- **THEN** it accepts interface types (GitOperations, WorkspaceStorage, ConfigProvider)
- **AND** concrete implementations are injected at runtime

#### Scenario: Interface definitions in ports package
- **WHEN** a developer looks for interface contracts
- **THEN** they find all port interfaces in `internal/ports/`
- **AND** each interface is documented with its contract

#### Scenario: Mock implementations for testing
- **WHEN** a test needs to isolate the Service
- **THEN** mock implementations from `internal/mocks/` can be injected
- **AND** error scenarios can be tested without filesystem or git access

#### Scenario: Compile-time interface checks
- **WHEN** an implementation is updated
- **THEN** compile-time assertions verify interface compliance
- **EXAMPLES**: `var _ ports.GitOperations = (*gitx.GitEngine)(nil)`

### Requirement: Centralized Service Initialization
The system SHALL initialize all services through a centralized App struct that manages dependencies and lifecycle.

#### Scenario: App creation succeeds
- **WHEN** `app.New(debug)` is called with valid config
- **THEN** an App struct is returned with initialized config, service, and logger
- **AND** all services are ready for use

#### Scenario: App creation fails with missing config
- **WHEN** `app.New(debug)` is called and config file does not exist
- **THEN** an error is returned describing the missing config
- **AND** no App instance is created

### Requirement: Command Registration Uses App Context
Commands SHALL be registered through builder functions that retrieve dependencies from the App stored in command context.

#### Scenario: Workspace commands registered
- **WHEN** the root command is initialized
- **THEN** workspace command builder functions are called
- **AND** workspace subcommands are attached to the root command
- **AND** each command can access the App via context

#### Scenario: Command execution with dependencies
- **WHEN** a user executes `canopy workspace new PROJ-123`
- **THEN** the command handler retrieves the App from context
- **AND** uses the App service to create the workspace
- **AND** no duplicate service initialization occurs

### Requirement: Testable Command Handlers
Command handlers SHALL support swapping dependencies for tests through the App struct.

#### Scenario: Unit test with mock service
- **WHEN** a test creates an App with mocked services
- **THEN** a command can execute using the mock
- **AND** the test can verify service method calls

#### Scenario: Integration test with real services
- **WHEN** a test creates an App with temporary directories
- **THEN** commands execute against the real filesystem and config
- **AND** the test can verify end-to-end behavior

### Requirement: No Global Service Variables
The system SHALL avoid global service or config variables, requiring commands to obtain dependencies from the App context.

#### Scenario: Command reads config without globals
- **WHEN** a command needs configuration or logger access
- **THEN** it retrieves the App from context
- **AND** uses App.Config and App.Logger instead of any global variables

### Requirement: Project Naming and Branding
The system SHALL be named "Canopy" with the binary named `canopy`, using forest/tree metaphors in all user-facing communication.

#### Scenario: Binary installation and invocation
- **WHEN** a user installs the tool via `go install`
- **THEN** the binary is named `canopy` (not `yard` or `yardmaster`)
- **AND** all commands are invoked as `canopy <command>`

#### Scenario: Configuration directory naming
- **WHEN** the system initializes or loads configuration
- **THEN** configuration is stored in `~/.canopy/` directory
- **AND** config file is `~/.canopy/config.yaml`

#### Scenario: Environment variables
- **WHEN** configuration is loaded from environment
- **THEN** environment variables use `CANOPY_` prefix
- **EXAMPLES**: `CANOPY_PROJECTS_ROOT`, `CANOPY_WORKSPACES_ROOT`

#### Scenario: Documentation uses consistent branding
- **WHEN** users read help text, README, or error messages
- **THEN** the project is referred to as "Canopy"
- **AND** metaphors reference canopy, forest, trees, and branches (not railroad/yard terminology)
- **AND** the metaphor explanation appears in the README introduction

### Requirement: Canopy Metaphor Documentation
The README SHALL include an explanation of the canopy metaphor in the introduction section.

#### Scenario: README metaphor explanation
- **WHEN** a user reads the README introduction
- **THEN** they see an explanation that canopy represents a bird's-eye view above the forest
- **AND** the explanation connects the metaphor to managing git workspaces and branches
- **AND** it clarifies that the TUI provides a literal canopy-level view of all workspaces

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

