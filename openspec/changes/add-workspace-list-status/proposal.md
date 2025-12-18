# Change: Add Status Flag to Workspace List

## Why
Currently `workspace list` only shows workspace IDs and branch names. Users frequently need git status (dirty/clean, ahead/behind) to decide which workspace to work on. The TUI shows this, but CLI users lack a quick way to see status without entering each workspace.

## What Changes
- Add `--status` flag to `canopy workspace list`
- Show git status indicators per repo when flag is set
- Opt-in to avoid expensive git calls by default
- Add timeout handling for slow/unresponsive repos

## Impact
- Affected specs: `cli`
- Affected code: `cmd/canopy/workspace.go`
