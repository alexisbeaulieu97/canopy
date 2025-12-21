# Change: Add dedicated error field to RepoStatus

## Why
Currently, errors in repository status are encoded in the `Branch` string field (e.g., "timeout", "error"), which conflates data with error states. This makes JSON/TUI/CLI output inconsistent and harder to parse programmatically.

## What Changes
- Add `Error` field to `domain.RepoStatus` struct
- Update status retrieval to populate error field instead of encoding in Branch
- Update TUI, CLI, and JSON output to display errors consistently
- **BREAKING**: JSON output format changes for repo status

## Impact
- Affected specs: core
- Affected code:
  - `internal/domain/domain.go` - Add Error field to RepoStatus
  - `internal/workspaces/status.go` - Populate error field
  - `cmd/canopy/status.go` - Update CLI output
  - `cmd/canopy/workspace_list.go` - Update status display
  - `internal/tui/view.go` - Update TUI rendering
