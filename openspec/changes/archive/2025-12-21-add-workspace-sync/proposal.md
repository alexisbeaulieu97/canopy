# Change: Add Workspace Sync Command

## Why
The `workspace git <ID> pull` command outputs raw git output, making it hard to quickly assess sync status across multiple repos. A dedicated `sync` command provides curated, summarized output for the common "pull all repos" workflow.

## What Changes
- Add `canopy workspace sync <ID>` command
- Display formatted summary table instead of raw git output
- Show per-repo status (updated, up-to-date, conflict, error)
- Add timeout handling for slow/unresponsive repos

## Impact
- Affected specs: `cli`
- Affected code: `cmd/canopy/workspace.go`, `internal/workspaces/service.go`
