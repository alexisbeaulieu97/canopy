## 1. Type Alias Removal
- [ ] 1.1 Remove `workspace.ClosedWorkspace` alias, update callers
- [ ] 1.2 Remove `executor.HookContext` alias, update callers

## 2. Wrapper Function Removal
- [ ] 2.1 Remove `repo_registry.DeriveAliasFromURL`, update callers
- [ ] 2.2 Remove `resolver.isLikelyURL`, verify no callers (private)
- [ ] 2.3 Remove `resolver.repoNameFromURL`, verify no callers (private)
- [ ] 2.4 Remove `service.CalculateDiskUsage`, update callers

## 3. Verification
- [ ] 3.1 Run tests to verify no regressions
- [ ] 3.2 Search for remaining "Deprecated:" comments

