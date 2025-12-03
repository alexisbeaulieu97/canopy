# Change: Add Workspace Pull Command

## Why
The TUI provides push functionality (`p` key), but there's no corresponding pull command. Developers frequently need to pull updates across all repos in a workspace after teammates push changes. Currently, users must manually `cd` into each repo or use `git -C` commands repetitively. A dedicated `canopy workspace pull` command would provide symmetry with push and streamline the sync workflow.

## What Changes
- Add `canopy workspace pull <WORKSPACE-ID>` command
- Pull all repos within the specified workspace
- Add `--rebase` flag to use `git pull --rebase`
- Add `--continue-on-error` flag to continue if one repo fails
- Show per-repo output with clear status indicators
- Add `l` key shortcut in TUI for pull (symmetric with `p` for push)

## Impact
- Affected specs: `specs/cli/spec.md`, `specs/tui/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go` - Add new `pull` subcommand
  - `internal/workspaces/service.go` - Add `PullWorkspace()` method
  - `internal/gitx/git.go` - Already has `Pull()` method
  - `internal/tui/tui.go` - Add `l` key handler
