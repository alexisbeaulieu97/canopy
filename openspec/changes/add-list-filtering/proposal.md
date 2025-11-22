# Change: Add Flexible List Filtering

## Why
The current `workspace list` command has no filtering capabilities. As workspaces accumulate, users need to quickly filter by status (stale, dirty, behind-remote, clean). Instead of adding individual flags (`--stale`, `--dirty`, etc.), a composable `--include`/`--exclude` system allows flexible filtering without flag explosion.

## What Changes
- Add `--include` flag to `workspace list` (comma-separated status values)
- Add `--exclude` flag to `workspace list` (comma-separated status values)
- Support filter values: `stale`, `dirty`, `behind`, `clean`, `archived`
- Add config option `workspace.list.include` and `workspace.list.exclude` for defaults
- Special `*` value means "all statuses" (default for include)
- Filters are additive: include first, then exclude

## Impact
- Affected specs: `specs/cli/spec.md`
- Affected code:
  - `cmd/canopy/workspace.go:114-179` - Add flags and filtering logic to list command
  - `internal/workspaces/service.go` - Add filtering methods
  - `internal/config/config.go` - Add list filter defaults
