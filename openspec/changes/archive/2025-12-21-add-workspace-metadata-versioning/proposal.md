# Change: Add Workspace Metadata Versioning

## Why
The `workspace.yaml` file has no schema versioning. This means:
1. **Silent breakage**: Structural changes (new fields, renamed fields) can silently break restores and exports
2. **No migration path**: Users can't upgrade old workspaces to new formats
3. **Import/export fragility**: Workspaces exported from one version may fail on another

## What Changes
- Add `version: 1` field to workspace metadata schema
- Enforce version validation during workspace load
- Add migration hooks for future schema changes
- Update export/import to include and validate version
- Default missing version to `0` for backward compatibility with existing workspaces

## Impact
- Affected specs: `workspace-management`
- Affected code:
  - `internal/domain/domain.go` - Add Version field to Workspace struct
  - `internal/workspace/workspace.go` - Validate version on load, default missing version
  - `internal/workspaces/export_service.go` - Include version in exports
  - Existing `workspace.yaml` files without version treated as version 0
- **No breaking changes** - Existing workspaces continue to work

