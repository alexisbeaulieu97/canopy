# Change: Consolidate Git Abstraction

## Why
`internal/gitx/git.go` mixes `go-git` library calls with `exec.Command("git", ...)` invocations, violating the project constraint against shelling out and causing inconsistent testing and error handling patterns.

## What Changes
- Migrate all operations to pure `go-git` library implementations
- Remove `exec.Command("git", ...)` calls from the codebase
- Consolidate error handling to use consistent `cerrors.WrapGitError()` patterns
- Add comprehensive go-git implementations for: CreateWorktree, Clone, Fetch, Pull, Push, Checkout
- Document any go-git limitations discovered during migration

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/gitx/git.go` - Migrate CLI calls to go-git
  - `internal/ports/git.go` - No changes to interface
- **Risk**: Medium - Core git operations, requires thorough testing of go-git equivalents
