# Change: Refactor TUI Model into State-Based Sub-Components

## Why
The current TUI `Model` struct has 20+ fields managing different concerns (view states, loading states, caches, UI components). This "God Struct" pattern makes the code harder to maintain, test, and extend. Extracting state into focused sub-components will improve separation of concerns and enable better testability.

## What Changes
- Extract view state management into a dedicated `ViewState` type with distinct states (List, Detail, Confirm)
- Extract workspace data management into a `WorkspaceModel` sub-component
- Extract UI component configuration into a `UIComponents` struct
- Simplify the main `Model` to coordinate sub-components
- Update `Update()` to delegate to state-specific handlers

## Impact
- Affected specs: `tui`, `tui-interface`
- Affected code:
  - `internal/tui/model.go` - Main model restructuring
  - `internal/tui/update.go` - State-delegated updates
  - `internal/tui/view.go` - View rendering per state
  - `internal/tui/commands.go` - Command creation
  - `internal/tui/messages.go` - Message types
