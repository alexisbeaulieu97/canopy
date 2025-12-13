## 1. Design State Types
- [x] 1.1 Define `ViewState` interface with `View()` and `Update()` methods
- [x] 1.2 Create `ListViewState` struct for workspace list management
- [x] 1.3 Create `DetailViewState` struct for workspace detail view
- [x] 1.4 Create `ConfirmViewState` struct for confirmation dialogs

## 2. Extract Sub-Components
- [x] 2.1 Create `WorkspaceModel` struct to manage workspace data and caches
- [x] 2.2 Move `allItems`, `statusCache`, `totalDiskUsage` to `WorkspaceModel`
- [x] 2.3 Create `UIComponents` struct for `list`, `spinner`, `keybindings`
- [x] 2.4 Add helper methods to `WorkspaceModel` for data access

## 3. Refactor Main Model
- [x] 3.1 Update `Model` struct to use new sub-components
- [x] 3.2 Update `NewModel()` to initialize sub-components
- [x] 3.3 Refactor `Init()` to work with new structure

## 4. Update Message Handling
- [x] 4.1 Refactor `Update()` to delegate to current `ViewState`
- [x] 4.2 Move state-specific key handling to respective state types
- [x] 4.3 Update message handlers to work with `WorkspaceModel`

## 5. Update View Rendering
- [x] 5.1 Refactor `View()` to delegate to current `ViewState.View()`
- [x] 5.2 Move list rendering to `ListViewState`
- [x] 5.3 Move detail rendering to `DetailViewState`
- [x] 5.4 Move confirm dialog rendering to `ConfirmViewState`

## 6. Testing
- [x] 6.1 Add unit tests for `ViewState` transitions
- [x] 6.2 Add unit tests for `WorkspaceModel` data management
- [x] 6.3 Update existing TUI tests for new structure

## 7. Documentation
- [x] 7.1 Update code comments with new architecture
- [x] 7.2 Add godoc comments to new types
