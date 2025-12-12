## Context
The TUI module's `Model` struct currently manages 20+ fields across multiple concerns: view state, loading indicators, caches, selected items, and UI components. This monolithic design makes it difficult to reason about state transitions and test individual concerns.

## Goals / Non-Goals
- **Goals**:
  - Improve separation of concerns in TUI code
  - Enable easier testing of individual view states
  - Simplify the main `Update()` function
  - Make state transitions explicit and traceable

- **Non-Goals**:
  - Change user-facing TUI behavior
  - Modify keybindings or visual appearance
  - Add new features

## Decisions

### Decision 1: State Pattern for View Management
Use the State pattern to manage different view modes (list, detail, confirm).

**Rationale**: The current boolean flags (`detailView`, `confirming`) create implicit state machines that are hard to follow. Explicit state objects make transitions clear.

```go
type ViewState interface {
    View(m *Model) string
    Update(m *Model, msg tea.Msg) (ViewState, tea.Cmd)
    HandleKey(m *Model, key string) (ViewState, tea.Cmd, bool)
}

type ListViewState struct{}
type DetailViewState struct{ workspace *domain.Workspace }
type ConfirmViewState struct{ action, targetID string }
```

### Decision 2: Composition over Inheritance
Extract data management into `WorkspaceModel` rather than using inheritance.

**Rationale**: Go doesn't support inheritance; composition provides cleaner separation.

```go
type WorkspaceModel struct {
    items       []workspaceItem
    statusCache map[string]*domain.WorkspaceStatus
    totalUsage  int64
    filterStale bool
    staleThreshold int
}

type Model struct {
    viewState   ViewState
    workspaces  *WorkspaceModel
    ui          UIComponents
    svc         *workspaces.Service
    err         error
}
```

### Decision 3: Preserve Backward Compatibility
Keep the same public API (`NewModel`, `Init`, `Update`, `View`) to avoid breaking changes.

## Risks / Trade-offs
- **Risk**: Increased number of files/types
  - **Mitigation**: Group related types in same file; document relationships
- **Risk**: Message routing complexity
  - **Mitigation**: Keep message types unchanged; only change handlers

## Migration Plan
1. Add new types alongside existing code
2. Refactor `Model` to use new types internally
3. Update handlers incrementally
4. Remove obsolete fields once migration complete
5. No rollback needed as this is internal refactoring

## Open Questions
- Should each `ViewState` be in its own file or grouped in `states.go`?
- Should `WorkspaceModel` expose fields or use getters?
