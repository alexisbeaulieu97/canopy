# Change: Add Repository Registry System

## Why
Users must either remember full Git URLs or manually configure patterns for each workspace type. The current system hardcodes "github.com/example/" for short names (service.go:58), which fails for actual repositories. A registry provides a user-friendly way to manage frequently-used repositories with aliases, eliminating the need to type full URLs repeatedly.

## What Changes
- Add `~/.yard/repos.yaml` registry file for storing repository aliases with metadata
- Implement `RepoRegistry` struct in `internal/config/` to manage registry operations
- Add new commands: `yard repo register <alias> <url>`, `yard repo unregister <alias>`, `yard repo list-registry`
- Update `ResolveRepos()` to check registry before falling back to URL patterns
- Support additional metadata: default branch, description, tags for filtering
- Auto-sync registry on `yard repo add` to create aliases from cloned repos

## Impact
- Affected specs: `specs/repository-management/spec.md` (new)
- Affected code:
  - `internal/config/repo_registry.go` (new) - Registry struct and persistence
  - `internal/config/config.go` - Integrate registry with Config
  - `internal/workspaces/service.go:36-82` - Update ResolveRepos() to check registry
  - `cmd/yard/repo.go` - Add register/unregister commands
  - `cmd/yard/workspace.go:29` - Accept registry aliases in --repos flag
