# Change: Refactor Dependency Injection

## Why
The `App` struct directly instantiates concrete types, making unit testing difficult since tests require real filesystems and git repositories. Injecting `ports.*` interfaces enables fast unit tests with mocks, better separation of concerns, and easier alternative implementations.

## What Changes
- Modify `App` struct to accept interface types via constructor options
- Create `AppOption` functional options pattern for flexible construction
- Update `app.New()` to use default implementations when options not provided
- Ensure existing CLI code continues to work unchanged (backward compatible)

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/app/app.go` - Refactor to use interfaces and options pattern
  - `internal/app/app_test.go` - Add unit tests with mocked dependencies
  - `cmd/canopy/*.go` - No changes needed (uses default construction)
- **Risk**: Low - Internal refactoring, no user-facing changes
