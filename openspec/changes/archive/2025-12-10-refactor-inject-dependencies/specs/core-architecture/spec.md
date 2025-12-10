# core-architecture Specification Delta

## MODIFIED Requirements

### Requirement: Centralized Service Initialization
The system SHALL initialize all services through a centralized App struct that manages dependencies and lifecycle. The App struct SHALL support functional options for injecting custom implementations.

#### Scenario: App creation with defaults
- **WHEN** `app.New(debug)` is called with valid config and no options
- **THEN** an App struct is returned with default GitEngine, WorkspaceStore, and ConfigProvider
- **AND** all services are ready for use

#### Scenario: App creation with custom dependencies
- **WHEN** `app.New(debug, WithGitOperations(mockGit), WithWorkspaceStorage(mockStore))` is called
- **THEN** an App struct is returned with the provided mock implementations
- **AND** the service uses the injected dependencies

#### Scenario: App creation fails with missing config
- **WHEN** `app.New(debug)` is called and config file does not exist
- **THEN** an error is returned describing the missing config
- **AND** no App instance is created

#### Scenario: Unit test with injected mocks
- **WHEN** a test creates App with `WithGitOperations(mocks.NewGitOps())`
- **THEN** the App uses the mock implementation
- **AND** no real git operations occur
