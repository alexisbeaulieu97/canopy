# Tasks: Refactor TUI Components

## Implementation Checklist

### 1. Component Infrastructure
- [x] 1.1 Create `internal/tui/components/` directory
- [x] 1.2 Create `internal/tui/components/components.go` with shared types

### 2. Extract StatusBadge
- [x] 2.1 Create `internal/tui/components/badge.go`
- [x] 2.2 Define `StatusBadge` struct with state (dirty, clean, stale, error)
- [x] 2.3 Move badge rendering logic from `delegate.go` and `view.go`
- [x] 2.4 Export `RenderStatusBadge()` function
- [x] 2.5 Update callers to use new component

### 3. Extract ConfirmDialog
- [x] 3.1 Create `internal/tui/components/confirm.go`
- [x] 3.2 Define `ConfirmDialog` model with title, message, callbacks
- [x] 3.3 Move confirmation logic from `update.go` (handleConfirmKey)
- [x] 3.4 Move confirmation rendering from `view.go` (renderConfirmPrompt)
- [x] 3.5 Update model to embed `ConfirmDialog`

### 4. Extract WorkspaceListItem
- [x] 4.1 Create `internal/tui/components/listitem.go`
- [x] 4.2 Move `workspaceItem` type from `helpers.go`
- [x] 4.3 Move item rendering logic from `delegate.go`
- [x] 4.4 Update delegate to use new component

### 5. Integration and Testing
- [x] 5.1 Update `internal/tui/model.go` to use components
- [x] 5.2 Add unit tests for each component
- [x] 5.3 Visual testing of TUI to ensure no regressions
- [x] 5.4 Run `canopy tui` and verify all interactions work
