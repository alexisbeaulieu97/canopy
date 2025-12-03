```markdown
# Change: Add Repo Status Command

## Why
Canopy manages canonical repositories (bare clones) in `projects_root`, but there's no command to view their status. Users cannot easily see which canonical repos exist, when they were last fetched, or if they have stale remotes. A `canopy repo status` command would provide visibility into the repo cache health.

## What Changes
- Add `canopy repo status [NAME]` command
- Without NAME: show all canonical repos with last-fetch time
- With NAME: show detailed status for one repo (branches, remotes, size)
- Add `--stale` flag to show only repos not fetched in N days
- Show which workspaces use each repo
- Add `--json` output for scripting

## Impact
- Affected specs: `specs/repository-management/spec.md`
- Affected code:
  - `cmd/canopy/repo.go` - Add new `status` subcommand
  - `internal/workspaces/service.go` - Add `GetCanonicalRepoStatus()` method
  - `internal/gitx/git.go` - Add methods to get repo metadata
```
