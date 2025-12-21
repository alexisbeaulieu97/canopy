# Change: Implement workspace_naming template end-to-end

## Why
The `workspace_naming` configuration exists but is not actually used - workspace directory paths are always derived directly from IDs. Users expect the documented template feature to work.

## What Changes
- Wire `workspace_naming` template to actual directory name computation
- Validate template at config load time
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
