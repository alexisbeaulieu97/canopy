```markdown
# Implementation Tasks

## 1. Create New Files
- [ ] 1.1 Create `internal/tui/messages.go` - extract all Msg types
- [ ] 1.2 Create `internal/tui/styles.go` - extract style definitions
- [ ] 1.3 Create `internal/tui/delegate.go` - extract workspaceDelegate
- [ ] 1.4 Create `internal/tui/commands.go` - extract tea.Cmd functions
- [ ] 1.5 Create `internal/tui/view.go` - extract View and render methods
- [ ] 1.6 Create `internal/tui/update.go` - extract Update and handlers
- [ ] 1.7 Create `internal/tui/model.go` - keep Model struct and Init

## 2. Extract Messages (messages.go)
- [ ] 2.1 Move workspaceListMsg
- [ ] 2.2 Move workspaceStatusMsg, workspaceStatusErrMsg
- [ ] 2.3 Move pushResultMsg, openEditorResultMsg
- [ ] 2.4 Move workspaceDetailsMsg

## 3. Extract Styles (styles.go)
- [ ] 3.1 Move statusCleanStyle, statusDirtyStyle, statusWarnStyle
- [ ] 3.2 Move subtleTextStyle, badgeStyle
- [ ] 3.3 Export styles or keep package-private

## 4. Extract Delegate (delegate.go)
- [ ] 4.1 Move workspaceItem type
- [ ] 4.2 Move workspaceSummary type
- [ ] 4.3 Move workspaceDelegate type and methods
- [ ] 4.4 Move healthForWorkspace, renderBadges helpers

## 5. Extract Commands (commands.go)
- [ ] 5.1 Move loadWorkspaces command
- [ ] 5.2 Move loadWorkspaceStatus command
- [ ] 5.3 Move pushWorkspace, closeWorkspace, openWorkspace commands
- [ ] 5.4 Move loadWorkspaceDetails command

## 6. Extract View (view.go)
- [ ] 6.1 Move View method
- [ ] 6.2 Move renderHeader, renderDetailView methods
- [ ] 6.3 Keep helpers.go for humanizeBytes, relativeTime

## 7. Extract Update (update.go)
- [ ] 7.1 Move Update method
- [ ] 7.2 Move handleKey, handleListKey, handleDetailKey, handleConfirmKey
- [ ] 7.3 Move handleEnter, handlePushConfirm, handleOpenEditor, handleCloseConfirm

## 8. Cleanup
- [ ] 8.1 Remove old tui.go or rename to model.go
- [ ] 8.2 Fix any import cycles
- [ ] 8.3 Run tests to ensure no regressions
- [ ] 8.4 Run linter

## 9. Testing
- [ ] 9.1 Verify `go build` succeeds
- [ ] 9.2 Verify existing helper tests pass
- [ ] 9.3 Manual test TUI functionality
```
