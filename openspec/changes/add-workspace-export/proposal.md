# Change: Add Workspace Export/Import

## Why
Users need a standard, portable workspace export/import to sync, share, and restore workspace definitions across machines. Current closed metadata provides only partial support with no standard export format or explicit import flow.

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
