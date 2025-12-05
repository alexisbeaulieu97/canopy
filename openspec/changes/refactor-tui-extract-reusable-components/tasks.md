# Tasks: Refactor TUI Components

## Implementation Checklist

### 1. Component Infrastructure
- [ ] 1.1 Create `internal/tui/components/` directory
- [ ] 1.2 Create `internal/tui/components/components.go` with shared types

### 2. Extract StatusBadge
- [ ] 2.1 Create `internal/tui/components/badge.go`
- [ ] 2.2 Define `StatusBadge` struct with state (dirty, clean, stale, error)
- [ ] 2.3 Move badge rendering logic from `delegate.go` and `view.go`
- [ ] 2.4 Export `RenderStatusBadge()` function
- [ ] 2.5 Update callers to use new component

### 3. Extract ConfirmDialog
- [ ] 3.1 Create `internal/tui/components/confirm.go`
- [ ] 3.2 Define `ConfirmDialog` model with title, message, callbacks
- [ ] 3.3 Move confirmation logic from `update.go` (handleConfirmKey)
- [ ] 3.4 Move confirmation rendering from `view.go` (renderConfirmPrompt)
- [ ] 3.5 Update model to embed `ConfirmDialog`

### 4. Extract WorkspaceListItem
- [ ] 4.1 Create `internal/tui/components/listitem.go`
- [ ] 4.2 Move `workspaceItem` type from `helpers.go`
- [ ] 4.3 Move item rendering logic from `delegate.go`
- [ ] 4.4 Update delegate to use new component

### 5. Integration and Testing
- [ ] 5.1 Update `internal/tui/model.go` to use components
- [ ] 5.2 Add unit tests for each component
- [ ] 5.3 Visual testing of TUI to ensure no regressions
- [ ] 5.4 Run `canopy tui` and verify all interactions work
