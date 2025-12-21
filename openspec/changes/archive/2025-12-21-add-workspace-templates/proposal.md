# Change: Add Workspace Templates

## Why
Users often create similar workspace configurations repeatedly (e.g., all backend work needs "backend" and "common" repos; frontend work needs "frontend", "ui-kit", and "design-system"). Currently, they must remember and type the repo list each time or configure regex patterns that are brittle and limited. Templates provide a simple, user-defined way to create workspaces with predefined repository sets and settings.

## What Changes
- Add template definitions to `~/.canopy/config.yaml` under `templates:` section
- Implement template resolution in workspace creation command
- Add `--template <name>` flag to `canopy workspace new` command
- Add `canopy template list` command to show available templates
- Support template metadata: repos, default branch pattern, environment setup commands
- Templates can be combined with explicit `--repos` flag (additive)

## Impact
- Affected specs: `specs/workspace-management/spec.md` (new)
- Affected code:
  - `internal/config/config.go` - Add Template type and parsing
  - `internal/workspaces/service.go:36-82` - Update CreateWorkspace to apply templates
  - `cmd/canopy/workspace.go:23-70` - Add --template flag to new command
  - `cmd/canopy/main.go` - Add `canopy template` command group (optional)
- Affected config: `config.yaml` - Add templates section to example
