```markdown
# Change: Add Workspace Sync Command

## Why
Developers often need to bring all repos in a workspace up-to-date with their remotes. This typically requires running `git fetch` followed by `git pull` in each repo. A `canopy workspace sync` command provides a single command to fetch and pull all repos, making it the go-to command for "make my workspace current."

## What Changes
- Add `canopy workspace sync <WORKSPACE-ID>` command
- Execute `git fetch --all` then `git pull` for each repo
- Add `--fetch-only` flag to only fetch without pulling
- Add `--rebase` flag to use rebase instead of merge
- Show progress and per-repo status
- Add `r` key shortcut in TUI for sync/refresh

## Impact
- Affected specs: `specs/cli/spec.md`, `specs/tui/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go` - Add new `sync` subcommand
  - `internal/workspaces/service.go` - Add `SyncWorkspace()` method
  - `internal/gitx/git.go` - Combine Fetch + Pull operations
  - `internal/tui/tui.go` - Add `r` key handler
```
