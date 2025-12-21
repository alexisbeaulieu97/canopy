# Change: Improve test coverage for critical packages

## Why
Several critical packages have low or missing test coverage:
- TUI components: 17.6% coverage, update.go has `//nolint:gocyclo`
- Workspace operations: create.go, close.go, sync.go lack unit tests
- Storage layer: 33.2% coverage

This increases regression risk and makes refactoring harder.

## What Changes
- Add unit tests for TUI update logic
- Add unit tests for workspace lifecycle operations
- Improve storage layer test coverage
- Remove `//nolint:gocyclo` by refactoring or proper testing

## Impact
- Affected specs: core-architecture
- Affected code:
  - `internal/tui/*_test.go` - Add TUI tests
  - `internal/workspaces/*_test.go` - Add workspace operation tests
  - `internal/storage/*_test.go` - Expand storage tests
