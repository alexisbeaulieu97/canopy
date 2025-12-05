# Change: Add Repository Status Command

## Why
While `canopy workspace git` handles git operations within workspaces, there's no way to check the health of **canonical repositories** (the bare clones in `projects_root`). Users need to know:
- When was each canonical repo last fetched?
- How much disk space do canonical repos use?
- Are there orphaned repos not used by any workspace?
- Is a repo outdated compared to remote?

This is operational visibility that `canopy workspace git` cannot provide since it operates on worktrees, not canonical repos.

## What Changes
- Add `canopy repo status [NAME]` command
- Shows: last fetch time, disk usage, workspace usage count
- Without NAME argument, shows status for all repos
- Add `--json` flag for scripting
- Optionally show "stale" indicator if not fetched recently

## Impact
- **Affected specs**: `specs/repository-management/spec.md`
- **Affected code**:
  - `cmd/canopy/repo.go` - Add status subcommand
  - `internal/workspaces/service.go` - Add `GetCanonicalRepoStatus()` method
  - `internal/gitx/git.go` - Add `LastFetchTime()` method
- **Risk**: Low - New additive feature
