# Change: Add Workspace Open Command

## Why
Users frequently need to open workspaces in their editor or view repos in the browser. Currently this requires manual navigation (`cd $(canopy w path ID)` then `$EDITOR .`). A dedicated `open` command streamlines the workflow for common operations: opening in editor and opening in browser.

## What Changes
- Add `canopy workspace open <WORKSPACE-ID>` command (opens in `$EDITOR`)
- Respect `$VISUAL` over `$EDITOR` when set
- Add `--browser` flag to open repo(s) in default browser
- Browser opens the remote URL for each repo (GitHub/GitLab/etc.)
- Support `--repo` flag to open specific repo instead of workspace root

## Impact
- Affected specs: `specs/cli/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go` - Add new `open` subcommand
  - `internal/workspaces/service.go` - Add `OpenWorkspace()` method
  - `internal/gitx/git.go` - Add `GetRemoteURL()` helper
