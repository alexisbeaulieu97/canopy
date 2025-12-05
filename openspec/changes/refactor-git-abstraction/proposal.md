# Change: Consolidate Git Abstraction

## Why
`internal/gitx/git.go` mixes two approaches:
1. `go-git` library for some operations (EnsureCanonical, Status)
2. `exec.Command("git", ...)` for others (CreateWorktree, Clone, Fetch, Pull, Push)

The project constraint states "No shelling out to git" for portability and testability, but the code shells out for "robustness". This inconsistency:
- Makes testing harder (some ops need real git, others don't)
- Creates confusion about which approach to use for new features
- May have different error handling patterns

## What Changes
- Document the intentional deviation from the constraint with clear rationale
- Create separate adapters: `GoGitAdapter` (library) and `GitCLIAdapter` (exec)
- Define a facade that chooses the appropriate adapter per operation
- Add configuration option to prefer one adapter over another (for testing)
- Ensure all operations have consistent error wrapping

## Impact
- **Affected specs**: `specs/core-architecture/spec.md`
- **Affected code**:
  - `internal/gitx/git.go` - Refactor into multiple files
  - `internal/gitx/gogit.go` - Pure go-git operations
  - `internal/gitx/cli.go` - CLI-based operations
  - `internal/gitx/facade.go` - Unified interface
  - `internal/ports/git.go` - No changes to interface
- **Risk**: Medium - Core git operations, needs careful testing
