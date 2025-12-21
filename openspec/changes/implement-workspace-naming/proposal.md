# Change: Implement workspace_naming template end-to-end

## Why
The `workspace_naming` configuration exists but is not actually used - workspace directory paths are always derived directly from IDs. Users expect the documented template feature to work.

## What Changes
- Wire `workspace_naming` template to actual directory name computation **BREAKING**. Existing workspaces may need to be moved or recreated to match the new template-derived paths.
- Validate template at config load time **BREAKING**. Invalid templates will now fail load; fix the template or roll back to a known-good config.
- Add template preview in `config validate` output
- Support future template variables beyond `{{.ID}}`

## Impact
- Affected specs: workspace-management
- Affected code:
  - `internal/config/config.go` - Template validation
  - `internal/workspaces/create.go` - Use template for directory name
  - `internal/storage/storage.go` - Respect naming template
  - `cmd/canopy/config.go` - Show template preview
  - `docs/configuration.md` - Clarify template usage
