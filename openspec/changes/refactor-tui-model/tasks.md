## 1. Design State Types
- [ ] 1.1 Define `ViewState` interface with `View()` and `Update()` methods
- [ ] 1.2 Create `ListViewState` struct for workspace list management
- [ ] 1.3 Create `DetailViewState` struct for workspace detail view
- [ ] 1.4 Create `ConfirmViewState` struct for confirmation dialogs

## 2. Extract Sub-Components
- [ ] 2.1 Create `WorkspaceModel` struct to manage workspace data and caches
- [ ] 2.2 Move `allItems`, `statusCache`, `totalDiskUsage` to `WorkspaceModel`
- [ ] 2.3 Create `UIComponents` struct for `list`, `spinner`, `keybindings`
- [ ] 2.4 Add helper methods to `WorkspaceModel` for data access

## 3. Refactor Main Model
- [ ] 3.1 Update `Model` struct to use new sub-components
- [ ] 3.2 Update `NewModel()` to initialize sub-components
- [ ] 3.3 Refactor `Init()` to work with new structure

## 4. Update Message Handling
- [ ] 4.1 Refactor `Update()` to delegate to current `ViewState`
- [ ] 4.2 Move state-specific key handling to respective state types
- [ ] 4.3 Update message handlers to work with `WorkspaceModel`

## 5. Update View Rendering
- [ ] 5.1 Refactor `View()` to delegate to current `ViewState.View()`
- [ ] 5.2 Move list rendering to `ListViewState`
- [ ] 5.3 Move detail rendering to `DetailViewState`
- [ ] 5.4 Move confirm dialog rendering to `ConfirmViewState`

## 6. Testing
- [ ] 6.1 Add unit tests for `ViewState` transitions
- [ ] 6.2 Add unit tests for `WorkspaceModel` data management
- [ ] 6.3 Update existing TUI tests for new structure

## 7. Documentation
- [ ] 7.1 Update code comments with new architecture
- [ ] 7.2 Add godoc comments to new types
