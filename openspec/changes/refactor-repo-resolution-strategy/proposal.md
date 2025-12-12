# Change: Refactor Repository Resolution to Use Strategy Pattern

## Why
The current `RepoResolver` uses a hard-coded chain of resolution attempts (URL, registry, GitHub shorthand). Adding new resolution strategies (e.g., GitLab shorthand, Bitbucket, enterprise GitHub) requires modifying the core resolver code. The Strategy pattern would make resolution extensible and easier to test.

## What Changes
- Define a `ResolutionStrategy` interface for repository resolution
- Convert existing resolution methods to strategy implementations
- Make the resolver accept a configurable list of strategies
- Enable users to configure which strategies are active (future enhancement)
- Consolidate duplicated `isLikelyURL` and `repoNameFromURL` functions into a shared `giturl` package

## Impact
- Affected specs: `core-architecture`, `repository-management`
- Affected code:
  - `internal/workspaces/resolver.go` - Refactor to use strategies
  - `internal/config/repo_registry.go` - Extract shared URL utilities
  - New package: `internal/giturl/` - Shared URL parsing utilities
