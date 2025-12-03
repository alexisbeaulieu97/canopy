# Implementation Tasks

## 1. Create New Files
- [x] 1.1 Create `internal/tui/messages.go` - extract all Msg types
- [x] 1.2 Create `internal/tui/styles.go` - extract style definitions
- [x] 1.3 Create `internal/tui/delegate.go` - extract workspaceDelegate
- [x] 1.4 Create `internal/tui/commands.go` - extract tea.Cmd functions
- [x] 1.5 Create `internal/tui/view.go` - extract View and render methods
- [x] 1.6 Create `internal/tui/update.go` - extract Update and handlers
- [x] 1.7 Create `internal/tui/model.go` - keep Model struct and Init

## 2. Extract Messages (messages.go)
- [x] 2.1 Move workspaceListMsg
- [x] 2.2 Move workspaceStatusMsg, workspaceStatusErrMsg
- [x] 2.3 Move pushResultMsg, openEditorResultMsg
- [x] 2.4 Move workspaceDetailsMsg

## 3. Extract Styles (styles.go)
- [x] 3.1 Move statusCleanStyle, statusDirtyStyle, statusWarnStyle
- [x] 3.2 Move subtleTextStyle, badgeStyle
- [x] 3.3 Export styles or keep package-private

## 4. Extract Delegate (delegate.go)
- [x] 4.1 Move workspaceItem type
- [x] 4.2 Move workspaceSummary type
- [x] 4.3 Move workspaceDelegate type and methods
- [x] 4.4 Move healthForWorkspace, renderBadges helpers

## 5. Extract Commands (commands.go)
- [x] 5.1 Move loadWorkspaces command
- [x] 5.2 Move loadWorkspaceStatus command
- [x] 5.3 Move pushWorkspace, closeWorkspace, openWorkspace commands
- [x] 5.4 Move loadWorkspaceDetails command

## 6. Extract View (view.go)
- [x] 6.1 Move View method
- [x] 6.2 Move renderHeader, renderDetailView methods
- [x] 6.3 Keep helpers.go for humanizeBytes, relativeTime

## 7. Extract Update (update.go)
- [x] 7.1 Move Update method
- [x] 7.2 Move handleKey, handleListKey, handleDetailKey, handleConfirmKey
- [x] 7.3 Move handleEnter, handlePushConfirm, handleOpenEditor, handleCloseConfirm

## 8. Cleanup
- [x] 8.1 Remove old tui.go or rename to model.go
- [x] 8.2 Fix any import cycles
- [x] 8.3 Run tests to ensure no regressions
- [x] 8.4 Run linter

## 9. Testing
- [x] 9.1 Verify `go build` succeeds
- [x] 9.2 Verify existing helper tests pass
- [x] 9.3 Manual test TUI functionality
