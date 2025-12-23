# Change: Add Progress Indicators for Long-Running Operations

## Why
Operations like bulk sync, workspace creation with many repos, and multi-workspace close can take significant time without any feedback. Users have no way to know if the operation is progressing or stuck.

## What Changes
- Add progress bars for multi-repository operations
- Show current operation status during bulk sync/close
- Add ETA estimates where feasible
- Support cancellation during progress display

## Impact
- Affected specs: cli
- Affected code:
  - `internal/output/` (new progress module)
  - `cmd/canopy/workspace_sync.go`
  - `cmd/canopy/workspace_new.go`
  - `cmd/canopy/workspace_close.go`
  - `internal/workspaces/parallel.go`
- No breaking changes
