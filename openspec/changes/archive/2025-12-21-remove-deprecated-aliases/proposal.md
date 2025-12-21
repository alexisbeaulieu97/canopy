# Change: Remove Deprecated Aliases and Functions

## Why
Several deprecated type aliases and wrapper functions add indirection and maintenance burden. These are internal-only and should be removed.

## What Changes
Remove the following deprecated items (all internal, no external dependents):
- `workspace.ClosedWorkspace` - alias for `domain.ClosedWorkspace`
- `hooks.HookContext` - alias for `domain.HookContext`
- `resolver.isLikelyURL` - wrapper for `giturl.IsURL`
- `resolver.repoNameFromURL` - wrapper for `giturl.ExtractRepoName`
- `service.CalculateDiskUsage` - wrapper for `DiskUsageCalculator.Calculate`

Update any internal callers to use the canonical implementations directly.

## Impact
- Affected specs: `core-architecture` (code hygiene)
- Affected code:
  - `internal/workspace/workspace.go` - Remove ClosedWorkspace alias
  - `internal/hooks/executor.go` - Remove HookContext alias
  - `internal/workspaces/resolver.go` - Remove isLikelyURL, repoNameFromURL
  - `internal/workspaces/service.go` - Remove CalculateDiskUsage wrapper
  - Update any callers to use canonical types/functions
- **Risk**: Very Low - Internal refactor only, no public API changes (verified: all items are in internal/ packages)

