# core-architecture Specification

## Purpose
TBD - created by archiving change refactor-service-initialization. Update Purpose after archive.
## Requirements
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
- **WHEN** a user executes `yard workspace new PROJ-123`
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

