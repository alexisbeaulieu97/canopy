## ADDED Requirements

### Requirement: No Deprecated Code Aliases
The codebase SHALL NOT contain the following deprecated type aliases or wrapper functions. All code SHALL use canonical implementations directly.

Deprecated items (to be removed):
- `workspace.ClosedWorkspace` (use `domain.ClosedWorkspace`)
- `hooks.HookContext` (use `domain.HookContext`)
- `resolver.isLikelyURL` (use `giturl.IsURL`)
- `resolver.repoNameFromURL` (use `giturl.ExtractRepoName`)
- `service.CalculateDiskUsage` (use `DiskUsageCalculator.Calculate`)

#### Scenario: Domain types used directly
- **WHEN** code needs to reference domain types like Workspace, ClosedWorkspace, or HookContext, **THEN** the code SHALL import and use `domain.TypeName` directly

#### Scenario: Utility functions used directly
- **WHEN** code needs URL parsing utilities, **THEN** the code SHALL import and use `giturl.FunctionName` directly

