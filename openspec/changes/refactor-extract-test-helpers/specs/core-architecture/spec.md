## ADDED Requirements

### Requirement: Shared Test Utilities Package
The codebase SHALL provide a shared test utilities package to avoid duplication of test helper functions.

#### Scenario: Git test helpers available
- **GIVEN** a test needs to create a git repository
- **WHEN** the test imports `internal/testutil`
- **THEN** `testutil.CreateRepoWithCommit(t, path)` SHALL be available
- **AND** the helper SHALL create a valid git repo with initial commit

#### Scenario: Filesystem test helpers available
- **GIVEN** a test needs to create temporary files
- **WHEN** the test imports `internal/testutil`
- **THEN** `testutil.MustMkdir(t, path)` SHALL be available
- **AND** `testutil.MustWriteFile(t, path, content)` SHALL be available

#### Scenario: Service test setup available
- **GIVEN** a test needs a fully configured test service
- **WHEN** the test calls `testutil.NewTestService(t)`
- **THEN** a struct with initialized dependencies SHALL be returned
- **AND** temporary directories SHALL be created
- **AND** cleanup SHALL be registered with t.Cleanup()

### Requirement: Test Helper Consistency
All test helper functions SHALL follow consistent patterns for error handling and cleanup.

#### Scenario: Helper fails test on error
- **GIVEN** a test helper function with `t *testing.T` parameter
- **WHEN** an error occurs during helper execution
- **THEN** the helper SHALL call `t.Fatalf()` with descriptive message
- **AND** the test SHALL stop execution

#### Scenario: Helper registers cleanup
- **GIVEN** a test helper that creates resources
- **WHEN** the helper completes successfully
- **THEN** cleanup functions SHALL be registered via `t.Cleanup()`
- **AND** resources SHALL be cleaned up after test completion

