## MODIFIED Requirements

### Requirement: Typed Error System
All errors returned by internal packages SHALL use the typed error system defined in `internal/errors`. Raw `fmt.Errorf` calls MUST NOT be used in production code paths.

The error system provides:
- `CanopyError` struct with `Code`, `Message`, `Cause`, and `Context` fields
- Constructor functions for each error type (e.g., `NewWorkspaceNotFound`, `NewIOFailed`)
- Sentinel errors for use with `errors.Is()` matching
- `Wrap()` for adding context while preserving the underlying error

#### Scenario: All internal packages use typed errors
- **GIVEN** the internal package structure
- **WHEN** any error is returned from internal code
- **THEN** the error SHALL be a `*CanopyError` or wrap one
- **AND** the error SHALL be matchable with `errors.Is(err, cerrors.SomeError)`

#### Scenario: Config validation returns typed errors
- **GIVEN** a config with an invalid value
- **WHEN** `ValidateValues()` is called
- **THEN** the error SHALL be a `ConfigValidation` error
- **AND** the error context SHALL include the field name and reason

#### Scenario: Path validation returns typed errors
- **GIVEN** a configured path that doesn't exist or isn't a directory
- **WHEN** `ValidateEnvironment()` is called
- **THEN** the error SHALL be a `PathInvalid` or `PathNotDirectory` error
- **AND** the error context SHALL include the path

#### Scenario: Workspace storage returns typed errors
- **GIVEN** workspace storage operations
- **WHEN** an I/O error occurs (read, write, delete)
- **THEN** the error SHALL be wrapped with `NewIOFailed` or `NewWorkspaceMetadataError`
- **AND** the underlying cause SHALL be preserved via `Unwrap()`

## ADDED Requirements

### Requirement: Configuration Validation Error Type
Config validation errors SHALL use the `ErrConfigValidation` error code for semantic validation failures.

#### Scenario: Invalid field value
- **WHEN** a config field has an invalid value
- **THEN** `NewConfigValidation(field, detail)` SHALL be returned
- **AND** the error message SHALL be user-friendly

### Requirement: Path Error Types
Path-related errors SHALL use specific error codes for different failure modes.

#### Scenario: Path does not exist
- **WHEN** a required path does not exist
- **THEN** `NewPathInvalid(path, "does not exist")` SHALL be returned

#### Scenario: Path is not a directory
- **WHEN** a path exists but is not a directory
- **THEN** `NewPathNotDirectory(path)` SHALL be returned

