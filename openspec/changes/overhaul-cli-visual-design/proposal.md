# Change: Overhaul CLI Visual Design

## Why

CLI output is currently functional but plain. Improved formatting with tables, colors, and structure makes output more scannable, professional, and pleasant to use. This aligns with modern CLI tools (gh, docker, kubectl) that prioritize readable output.

## What Changes

- **Tables**: Introduce box-drawn tables for list outputs with column alignment
- **Status Indicators**: Colored inline badges for workspace/repo status
- **Headers & Sections**: Clear visual hierarchy for multi-section outputs
- **Progress Output**: Spinners and progress bars for long operations
- **Error Formatting**: Boxed error messages with context and suggestions
- **Success Messages**: Consistent formatting with icons and colors

## Impact

- Affected specs: `cli`
- Affected code:
  - `internal/output/` (new table, progress, error formatters)
  - `cmd/canopy/workspace_list.go`
  - `cmd/canopy/workspace_view.go`
  - `cmd/canopy/workspace_sync.go`
  - `cmd/canopy/presenters.go`
  - All command files with user-facing output
