# Change: Standardize Error Handling

## Why
The codebase has a `CanopyError` type system but many paths still use raw `fmt.Errorf()`, preventing programmatic error handling, creating inconsistent user messages, and stripping debugging context.

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
