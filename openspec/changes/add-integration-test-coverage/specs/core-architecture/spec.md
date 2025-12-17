## ADDED Requirements

### Requirement: Integration Test Coverage

The project SHALL maintain integration tests covering all major user workflows.

Integration tests SHALL exercise the complete stack from CLI through service layer to filesystem operations, using real git repositories in isolated temporary directories.

#### Scenario: Workspace lifecycle coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL cover workspace create, list, view, close, restore, and rename operations

#### Scenario: Repository management coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL cover adding repos to workspaces, removing repos, and status reporting

#### Scenario: Branch operation coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL cover branch switching and creation across workspace repositories

#### Scenario: Error handling coverage

- **WHEN** running the integration test suite
- **THEN** tests SHALL verify appropriate error messages for common failure scenarios including dirty repos, missing workspaces, and invalid configuration
