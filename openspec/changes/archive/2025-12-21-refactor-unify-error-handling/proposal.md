# Change: Unify Error Handling with Typed Errors

## Why
The codebase has 81 occurrences of `fmt.Errorf` scattered across internal packages, despite having a well-designed typed error system in `internal/errors`. This inconsistency makes it impossible to reliably match errors with `errors.Is()` for error handling, creates unclear error messages for users, and violates the architectural principle documented in `core-architecture/spec.md`.

## What Changes
- Convert all `fmt.Errorf` calls in `internal/workspace/workspace.go` (27 instances) to use typed errors
- Convert all `fmt.Errorf` calls in `internal/config/config.go` (26 instances) to use typed errors
- Convert all `fmt.Errorf` calls in `internal/config/repo_registry.go` (10 instances) to use typed errors
- Convert `fmt.Errorf` calls in `internal/tui/commands.go` (3 instances) to use typed errors
- Convert `fmt.Errorf` calls in `internal/workspaces/service.go` (2 instances) to use typed errors
- Add new error types as needed: `ErrConfigValidation`, `ErrPathInvalid`, `ErrPathNotDirectory`
- Add corresponding constructor functions to `internal/errors/errors.go`

## Impact
- Affected specs: core-architecture
- Affected code:
  - `internal/errors/errors.go` (new error types and constructors)
  - `internal/workspace/workspace.go` (27 conversions)
  - `internal/config/config.go` (26 conversions)
  - `internal/config/repo_registry.go` (10 conversions)
  - `internal/tui/commands.go` (3 conversions)
  - `internal/workspaces/service.go` (2 conversions)
  - Test files that depend on error messages

