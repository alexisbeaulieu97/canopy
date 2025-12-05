# Change: Standardize Error Handling

## Why
The codebase has a good `CanopyError` type system in `internal/errors/errors.go` with:
- Error codes (WORKSPACE_NOT_FOUND, REPO_NOT_CLEAN, etc.)
- Context maps for debugging
- Proper `Is()` and `Unwrap()` implementations

However, many places still use raw `fmt.Errorf()`:
- `cmd/canopy/status.go` returns `fmt.Errorf("not inside a workspace")`
- `cmd/canopy/check.go` returns `fmt.Errorf("configuration is invalid: %w", err)`
- Various service methods return untyped errors

This inconsistency hurts:
- Programmatic error handling (can't switch on error type)
- User experience (inconsistent error messages)
- Debugging (no context maps)

## What Changes
- Audit all error returns in `cmd/` and `internal/`
- Replace `fmt.Errorf()` with appropriate `cerrors.New*()` constructors
- Add new error types as needed (e.g., `ErrNotInWorkspace`, `ErrCommandFailed`)
- Ensure CLI commands use typed errors for exit codes

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/errors/errors.go` - Add missing error types
  - `cmd/canopy/status.go` - Use typed errors
  - `cmd/canopy/check.go` - Use typed errors
  - `internal/workspaces/service.go` - Audit and fix
  - `internal/gitx/git.go` - Audit and fix
- **Risk**: Low - Error handling changes, backward compatible
