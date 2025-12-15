# Change: Remove Deprecated Aliases and Functions

## Why
Several deprecated type aliases and wrapper functions exist that add confusion and maintenance burden:
1. `workspace.ClosedWorkspace` - alias for `domain.ClosedWorkspace`
2. `executor.HookContext` - alias for `domain.HookContext`
3. `repo_registry.DeriveAliasFromURL` - wrapper for `giturl.DeriveAlias`
4. `resolver.isLikelyURL` - wrapper for `giturl.IsURL`
5. `resolver.repoNameFromURL` - wrapper for `giturl.ExtractRepoName`
6. `service.CalculateDiskUsage` - wrapper for `DiskUsageCalculator.Calculate`

These were kept for backward compatibility but add indirection without value.

## What Changes
- Remove the deprecated type aliases and update any internal usages
- Remove the deprecated wrapper functions
- Update any callers to use the canonical implementations directly

## Impact
- Affected specs: `core-architecture` (code hygiene)
- Affected code:
  - `internal/workspace/workspace.go` - Remove ClosedWorkspace alias
  - `internal/hooks/executor.go` - Remove HookContext alias
  - `internal/config/repo_registry.go` - Remove DeriveAliasFromURL
  - `internal/workspaces/resolver.go` - Remove isLikelyURL, repoNameFromURL
  - `internal/workspaces/service.go` - Remove CalculateDiskUsage wrapper
  - Update any callers to use canonical types/functions
- **Risk**: Very Low - Internal refactor only, no public API changes

