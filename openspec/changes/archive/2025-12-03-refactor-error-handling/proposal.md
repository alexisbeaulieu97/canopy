```markdown
# Change: Refactor Error Handling with Custom Error Types

## Why
Current error handling uses `fmt.Errorf` throughout, making it difficult to programmatically distinguish error types. Code like `if strings.Contains(err.Error(), "not found")` is fragile. Custom error types enable:
- Proper error type checking with `errors.Is()` and `errors.As()`
- Consistent user-facing error messages
- Better testability (assert specific error types)
- Cleaner error handling in CLI layer

## What Changes
- Create `internal/errors/errors.go` with domain error types
- Define: `ErrWorkspaceNotFound`, `ErrWorkspaceExists`, `ErrRepoNotFound`, `ErrUncleanWorkspace`, `ErrInvalidConfig`
- Wrap errors with context using `fmt.Errorf("%w", err)` pattern
- Update CLI commands to check error types and show appropriate messages
- Add error codes for JSON output

## Impact
- Affected specs: `specs/core-architecture/spec.md` (new)
- Affected code:
  - `internal/errors/errors.go` (new) - Error type definitions
  - `internal/workspaces/service.go` - Return typed errors
  - `internal/workspace/workspace.go` - Return typed errors
  - `cmd/canopy/*.go` - Handle typed errors appropriately
```
