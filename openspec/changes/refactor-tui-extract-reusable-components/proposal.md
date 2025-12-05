# Change: Refactor TUI Components

## Why
TUI code is tightly coupled across files (`model.go`, `update.go`, `view.go`, `commands.go`) with implicit shared state, making feature additions and testing cumbersome. Extracting reusable components improves extensibility and testability.

## What Changes
- Extract `StatusBadge` component for dirty/clean/stale indicators
- Extract `ConfirmDialog` component for confirmation prompts
- Extract `WorkspaceListItem` component for list rendering
- Create `components/` subdirectory for reusable pieces
- Define clear interfaces between components and main model

## Impact
- **Affected specs**: `specs/tui-interface/spec.md`
- **Affected code**:
  - `internal/tui/components/` - New directory for components
  - `internal/tui/components/badge.go` - Status badge component
  - `internal/tui/components/confirm.go` - Confirmation dialog
  - `internal/tui/components/listitem.go` - Workspace list item
  - `internal/tui/view.go` - Use extracted components
  - `internal/tui/delegate.go` - Use extracted list item
- **Risk**: Medium - UI refactoring, needs visual testing
