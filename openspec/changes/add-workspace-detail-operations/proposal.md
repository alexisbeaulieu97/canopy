# Change: Add Workspace Detail View Operations

## Why

When viewing workspace details, users frequently want to perform actions on that specific workspace (push, sync, open in editor). Currently, they must exit the detail view, return to the list, and then perform the operation. This adds friction to a common workflow.

## What Changes

- Add keyboard shortcuts for common operations within the detail view
- Operations available: Push, Sync, Open in Editor, Close workspace
- Confirmation dialogs appear within detail view context
- After operation completion, user remains in detail view (or returns to list if workspace was closed)

## Impact

- Affected specs: `tui`
- Affected code: `internal/tui/update.go`, `internal/tui/view.go`, `internal/tui/states.go`
