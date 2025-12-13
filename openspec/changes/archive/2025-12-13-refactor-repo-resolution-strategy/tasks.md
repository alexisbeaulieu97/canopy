## 1. Create Shared URL Utilities Package
- [x] 1.1 Create `internal/giturl/` package
- [x] 1.2 Move `isLikelyURL()` to `giturl.IsURL()`
- [x] 1.3 Move `repoNameFromURL()` to `giturl.ExtractRepoName()`
- [x] 1.4 Add `giturl.DeriveAlias()` (consolidate with `DeriveAliasFromURL`)
- [x] 1.5 Add comprehensive URL scheme support (http, https, ssh, git, file, git@)
- [x] 1.6 Write unit tests for URL utilities

## 2. Define Strategy Interface
- [x] 2.1 Create `ResolutionStrategy` interface in `internal/workspaces/`
- [x] 2.2 Define `Resolve(input string) (domain.Repo, bool)` method
- [x] 2.3 Define `Name() string` method for debugging/logging

## 3. Implement Built-in Strategies
- [x] 3.1 Create `URLStrategy` for direct URL resolution
- [x] 3.2 Create `RegistryStrategy` for alias lookup
- [x] 3.3 Create `GitHubShorthandStrategy` for `owner/repo` format
- [x] 3.4 Write unit tests for each strategy

## 4. Refactor RepoResolver
- [x] 4.1 Update `RepoResolver` to hold a slice of strategies
- [x] 4.2 Update `NewRepoResolver()` to accept strategy configuration
- [x] 4.3 Create default strategy chain in constructor
- [x] 4.4 Refactor `Resolve()` to iterate through strategies
- [x] 4.5 Update error handling for unknown repositories

## 5. Update Callers
- [x] 5.1 Update `workspaces.NewService()` to configure resolver
- [x] 5.2 Update `config/repo_registry.go` to use `giturl` package
- [x] 5.3 Verify all existing tests pass

## 6. Documentation
- [x] 6.1 Add godoc comments to new types
- [x] 6.2 Document strategy order and precedence
