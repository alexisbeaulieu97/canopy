# Core Architecture Spec (Delta)

## Purpose
Error handling patterns and type system for Canopy.

## New Error Types

### Domain Errors
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

## Error Handling Guidelines
- Use typed errors for recoverable conditions
- Include context for debugging
- Chain errors with `Wrap` for root cause
- Check with `errors.Is()` or `errors.As()`
