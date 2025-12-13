# Change: Add Input Validation Hardening

## Why
Input validation is currently inconsistent across the codebase. Some inputs are validated while others are not, creating potential security risks (path traversal, injection) and poor UX when invalid inputs are accepted silently or produce confusing errors.

## What Changes
- Add workspace ID format validation (pattern matching)
- Validate branch names against git ref rules
- Prevent path traversal attacks in workspace/repo names
- Add input length limits for workspace IDs, repo names, branch names
- Centralize validation functions in a `validation` package
- Return typed errors for all validation failures

## Impact
- Affected specs: core
- Affected code:
  - New `internal/validation/` package
  - `internal/workspaces/service.go` (add validation calls)
  - `internal/workspace/workspace.go` (add validation calls)
  - CLI commands (validate before service calls)

