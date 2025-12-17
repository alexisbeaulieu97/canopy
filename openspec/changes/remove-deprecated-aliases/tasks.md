## 1. Type Alias Removal
- [x] 1.1 Remove `workspace.ClosedWorkspace` alias, update callers
- [x] 1.2 Remove `executor.HookContext` alias, update callers

## 2. Wrapper Function Removal
- [x] 2.1 Remove `repo_registry.DeriveAliasFromURL`, update callers
- [x] 2.2 Remove `resolver.isLikelyURL`, verify no callers (private)
- [x] 2.3 Remove `resolver.repoNameFromURL`, verify no callers (private)
- [x] 2.4 Remove `service.CalculateDiskUsage`, update callers

## 3. Verification
- [x] 3.1 Run tests to verify no regressions
- [x] 3.2 Search for remaining "Deprecated:" comments

