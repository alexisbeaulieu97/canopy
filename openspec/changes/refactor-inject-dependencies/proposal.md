# Change: Refactor Dependency Injection

## Why
The `App` struct in `internal/app/app.go` directly instantiates concrete types (`gitx.New()`, `workspace.New()`), making unit testing difficult. Tests currently require real filesystems and git repositories. By injecting `ports.*` interfaces instead of concrete implementations, we enable:
- Fast unit tests with mocks
- Better separation of concerns
- Easier extension with alternative implementations

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
