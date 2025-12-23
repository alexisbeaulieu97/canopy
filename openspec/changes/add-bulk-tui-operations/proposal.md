# Change: Add Bulk Operations to TUI

## Why
The TUI currently only allows operating on one workspace at a time. Users managing many workspaces need to select multiple and perform bulk actions (sync, close, push) without leaving the TUI or using CLI pattern matching.

## What Changes
- Add multi-select mode with Space key
- Add select all/deselect all shortcuts
- Support bulk sync, close, push operations on selected workspaces
- Show selection count in status bar

## Impact
- Affected specs: tui
- Affected code:
  - `internal/tui/model.go` (selection state)
  - `internal/tui/keys.go` (new keybindings)
  - `internal/tui/view.go` (selection indicators)
  - `internal/tui/update.go` (bulk operation handlers)
- No breaking changes
