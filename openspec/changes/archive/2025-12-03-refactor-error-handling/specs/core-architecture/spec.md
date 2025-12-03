# Core Architecture Spec (Delta)

## ADDED Requirements

### Requirement: Typed Error System
The application SHALL use typed errors with error codes for all domain errors.

#### Scenario: Create workspace not found error
- **WHEN** `NewWorkspaceNotFound("my-ws")` is called
- **THEN** returned CanopyError SHALL have Code `WORKSPACE_NOT_FOUND`
- **AND** Message SHALL contain the workspace ID "my-ws"

#### Scenario: Create repo not clean error
- **WHEN** `NewRepoNotClean("/path/to/repo")` is called
- **THEN** returned CanopyError SHALL have Code `REPO_NOT_CLEAN`
- **AND** Context SHALL contain the repo path

### Requirement: Error Wrapping
The application SHALL support wrapping errors to preserve root cause.

#### Scenario: Wrap git error
- **WHEN** `WrapGitError(err, "clone")` is called with underlying error
- **THEN** returned CanopyError SHALL have Code `GIT_OPERATION_FAILED`
- **AND** Cause SHALL contain the original error
- **AND** `errors.Unwrap()` SHALL return the original error

### Requirement: Error Matching
Errors SHALL support standard Go error matching with `errors.Is()` and `errors.As()`.

#### Scenario: Match error by code
- **WHEN** CanopyError is returned and `errors.Is()` is used
- **THEN** matching SHALL succeed for errors with same ErrorCode

#### Scenario: Extract error details
- **WHEN** `errors.As()` is used with `*CanopyError`
- **THEN** full error details SHALL be accessible including Code, Message, and Context

### Requirement: Error Context
Errors SHALL support contextual key-value pairs for debugging.

#### Scenario: Include context in error
- **WHEN** error is created with Context map
- **THEN** Context SHALL contain all provided key-value pairs
- **AND** Context SHALL be accessible for logging and debugging

## Reference

### Error Types
```go
type ErrorCode string

const (
    ErrWorkspaceNotFound    ErrorCode = "WORKSPACE_NOT_FOUND"
    ErrWorkspaceExists      ErrorCode = "WORKSPACE_EXISTS"
    ErrRepoNotFound         ErrorCode = "REPO_NOT_FOUND"
    ErrRepoNotClean         ErrorCode = "REPO_NOT_CLEAN"
    ErrGitOperationFailed   ErrorCode = "GIT_OPERATION_FAILED"
    ErrConfigInvalid        ErrorCode = "CONFIG_INVALID"
)

type CanopyError struct {
    Code      ErrorCode
    Message   string
    Cause     error
    Context   map[string]string
}

func (e *CanopyError) Error() string
func (e *CanopyError) Unwrap() error
func (e *CanopyError) Is(target error) bool
```

### Error Constructors
```go
func NewWorkspaceNotFound(id string) *CanopyError
func NewRepoNotClean(path string) *CanopyError
func WrapGitError(err error, operation string) *CanopyError
```
