# Change: Remove Deprecated Code

## Why
`internal/workspace/workspace.go:37` has a deprecated type alias:
```go
// ClosedWorkspace is an alias for domain.ClosedWorkspace for backward compatibility.
//
// Deprecated: Use domain.ClosedWorkspace directly.
type ClosedWorkspace = domain.ClosedWorkspace
```

Deprecated code should be removed to:
- Reduce confusion for new contributors
- Simplify the codebase
- Avoid accidental use of deprecated types

## What Changes
- Remove the `ClosedWorkspace` type alias from `workspace/workspace.go`
- Update any code still using `workspace.ClosedWorkspace` to use `domain.ClosedWorkspace`
- Search for other deprecated markers and remove if safe

## Impact
- **Affected specs**: None (internal cleanup)
- **Affected code**:
  - `internal/workspace/workspace.go` - Remove alias
  - Any files importing `workspace.ClosedWorkspace` - Update imports
- **Risk**: Very Low - Simple type alias removal
