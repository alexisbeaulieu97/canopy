# Change: Add Workspace Export/Import

## Why
For multi-machine workflows, users need to:
- Work on the same set of repos from different computers
- Share workspace configurations with teammates
- Backup workspace definitions before system changes
- Quickly recreate workspaces after OS reinstall

Current `closed` metadata provides partial support, but there's no standard export format or explicit import flow.

## What Changes
- Add `canopy workspace export <ID>` - outputs workspace definition as YAML/JSON
- Add `canopy workspace import <file>` - creates workspace from exported definition
- Export includes: ID, branch, repo list (with URLs), optional notes
- Import resolves repos via registry or URLs and creates workspace

## Impact
- **Affected specs**: `specs/workspace-management/spec.md`
- **Affected code**:
  - `cmd/canopy/workspace.go` - Add export/import subcommands
  - `internal/workspaces/service.go` - Add export/import methods
  - `internal/domain/domain.go` - Define exportable workspace schema
- **Risk**: Low - New feature, doesn't change existing workflows
