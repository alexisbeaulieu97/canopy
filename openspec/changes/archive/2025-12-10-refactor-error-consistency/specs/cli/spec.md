# cli Specification Delta

## ADDED Requirements

### Requirement: Typed Error Returns
CLI commands SHALL return typed errors from `internal/errors` rather than `fmt.Errorf` strings.

#### Scenario: Command returns typed error
- **WHEN** a CLI command encounters an error condition
- **THEN** it returns a typed error (e.g., `ErrWorkspaceNotFound`, `ErrInvalidArgument`)
- **AND** the error can be inspected with `errors.Is()` or `errors.As()`

#### Scenario: Error includes context
- **WHEN** a typed error is returned
- **THEN** it includes contextual information (workspace ID, repo name, etc.)
- **AND** the error message is user-friendly

### Requirement: Exit Code Mapping
The CLI SHALL return consistent exit codes mapped from error types.

#### Scenario: Normal success returns 0
- **WHEN** a command completes successfully
- **THEN** the exit code is 0

#### Scenario: Typed errors map to exit codes
- **WHEN** a command returns a typed error
- **THEN** the exit code is determined by the error type
- **AND** `ErrWorkspaceNotFound` returns exit code 1
- **AND** `ErrInvalidArgument` returns exit code 2
- **AND** `ErrConfigInvalid` returns exit code 3
