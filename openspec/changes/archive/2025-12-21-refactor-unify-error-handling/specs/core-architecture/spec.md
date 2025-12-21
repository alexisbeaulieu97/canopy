## MODIFIED Requirements

### Requirement: Typed Error System
All errors returned by internal packages SHALL use the typed error system defined in `internal/errors`. Raw `fmt.Errorf` calls MUST NOT be used in production code paths.

The error system provides:
- `CanopyError` struct with `Code`, `Message`, `Cause`, and `Context` fields
- Constructor functions for each error type (e.g., `NewWorkspaceNotFound`, `NewIOFailed`)
- Sentinel errors for use with `errors.Is()` matching
- `Wrap()` for adding context while preserving the underlying error

#### Scenario: All internal packages use typed errors
- **WHEN** any error is returned from internal code, **THEN** it SHALL be a `*CanopyError` or wrap one
- **WHEN** checking error type, **THEN** it SHALL be matchable with `errors.Is(err, cerrors.SomeError)`

#### Scenario: Config validation returns typed errors
- **WHEN** `ValidateValues()` is called on a config with an invalid value, **THEN** the error SHALL be a `ConfigValidation` error
- **WHEN** a config validation error occurs, **THEN** the error context SHALL include the field name and reason

#### Scenario: Path validation returns typed errors
- **WHEN** `ValidateEnvironment()` is called on a path that doesn't exist or isn't a directory, **THEN** the error SHALL be a `PathInvalid` or `PathNotDirectory` error
- **WHEN** a path validation error occurs, **THEN** the error context SHALL include the path

#### Scenario: Workspace storage returns typed errors
- **WHEN** an I/O error occurs during workspace storage operations (read, write, delete), **THEN** the error SHALL be wrapped with `NewIOFailed` or `NewWorkspaceMetadataError`
- **WHEN** wrapping storage errors, **THEN** the underlying cause SHALL be preserved via `Unwrap()`

## ADDED Requirements

### Requirement: Configuration Validation Error Type
Config validation errors SHALL use the `ErrConfigValidation` error code for semantic validation failures.

#### Scenario: Invalid field value
- **WHEN** a config field has an invalid value, **THEN** `NewConfigValidation(field, detail)` SHALL be returned
- **WHEN** returning config validation errors, **THEN** the error message SHALL be user-friendly

### Requirement: Path Error Types
Path-related errors SHALL use specific error codes for different failure modes.

#### Scenario: Path does not exist
- **WHEN** a required path does not exist, **THEN** `NewPathInvalid(path, "does not exist")` SHALL be returned

#### Scenario: Path is not a directory
- **WHEN** a path exists but is not a directory, **THEN** `NewPathNotDirectory(path)` SHALL be returned

