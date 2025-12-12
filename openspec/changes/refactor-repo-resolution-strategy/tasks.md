## 1. Create Shared URL Utilities Package
- [ ] 1.1 Create `internal/giturl/` package
- [ ] 1.2 Move `isLikelyURL()` to `giturl.IsURL()`
- [ ] 1.3 Move `repoNameFromURL()` to `giturl.ExtractRepoName()`
- [ ] 1.4 Add `giturl.DeriveAlias()` (consolidate with `DeriveAliasFromURL`)
- [ ] 1.5 Add comprehensive URL scheme support (http, https, ssh, git, file, git@)
- [ ] 1.6 Write unit tests for URL utilities

## 2. Define Strategy Interface
- [ ] 2.1 Create `ResolutionStrategy` interface in `internal/workspaces/`
- [ ] 2.2 Define `Resolve(input string) (domain.Repo, bool)` method
- [ ] 2.3 Define `Name() string` method for debugging/logging

## 3. Implement Built-in Strategies
- [ ] 3.1 Create `URLStrategy` for direct URL resolution
- [ ] 3.2 Create `RegistryStrategy` for alias lookup
- [ ] 3.3 Create `GitHubShorthandStrategy` for `owner/repo` format
- [ ] 3.4 Write unit tests for each strategy

## 4. Refactor RepoResolver
- [ ] 4.1 Update `RepoResolver` to hold a slice of strategies
- [ ] 4.2 Update `NewRepoResolver()` to accept strategy configuration
- [ ] 4.3 Create default strategy chain in constructor
- [ ] 4.4 Refactor `Resolve()` to iterate through strategies
- [ ] 4.5 Update error handling for unknown repositories

## 5. Update Callers
- [ ] 5.1 Update `workspaces.NewService()` to configure resolver
- [ ] 5.2 Update `config/repo_registry.go` to use `giturl` package
- [ ] 5.3 Verify all existing tests pass

## 6. Documentation
- [ ] 6.1 Add godoc comments to new types
- [ ] 6.2 Document strategy order and precedence
