# Tasks: Refactor TUI Components

## Implementation Checklist

### Phase 1: Component Infrastructure
- [ ] Create `internal/tui/components/` directory
- [ ] Create `internal/tui/components/components.go` with shared types

### Phase 2: Extract StatusBadge
- [ ] Create `internal/tui/components/badge.go`
- [ ] Define `StatusBadge` struct with state (dirty, clean, stale, error)
- [ ] Move badge rendering logic from `delegate.go` and `view.go`
- [ ] Export `RenderStatusBadge()` function
- [ ] Update callers to use new component

### Phase 3: Extract ConfirmDialog
- [ ] Create `internal/tui/components/confirm.go`
- [ ] Define `ConfirmDialog` model with title, message, callbacks
- [ ] Move confirmation logic from `update.go` (handleConfirmKey)
- [ ] Move confirmation rendering from `view.go` (renderConfirmPrompt)
- [ ] Update model to embed `ConfirmDialog`

### Phase 4: Extract WorkspaceListItem
- [ ] Create `internal/tui/components/listitem.go`
- [ ] Move `workspaceItem` type from `helpers.go`
- [ ] Move item rendering logic from `delegate.go`
- [ ] Update delegate to use new component

### Phase 5: Integration and Testing
- [ ] Update `internal/tui/model.go` to use components
- [ ] Add unit tests for each component
- [ ] Visual testing of TUI to ensure no regressions
- [ ] Run `canopy tui` and verify all interactions work
