```markdown
# Change: Refactor TUI into Modular Components

## Why
The current `internal/tui/tui.go` is 880 lines, handling models, views, updates, delegates, helpers, and messages in a single file. This makes the code difficult to navigate, test, and maintain. Splitting into focused modules improves:
- Code organization and discoverability
- Independent testing of components
- Easier onboarding for contributors
- Separation of concerns

## What Changes
- Split `tui.go` into multiple focused files:
  - `model.go` - Model struct, Init, constructor
  - `update.go` - Update function and message handling
  - `view.go` - View function and rendering helpers
  - `delegate.go` - workspaceDelegate and item rendering
  - `messages.go` - All tea.Msg types
  - `commands.go` - tea.Cmd factory functions
  - `styles.go` - Lipgloss style definitions
  - `helpers.go` - Utility functions (humanizeBytes, relativeTime)
- Keep package as `tui` (no API changes)
- Ensure all tests still pass

## Impact
- Affected specs: None (internal refactor)
- Affected code:
  - `internal/tui/tui.go` → split into multiple files
  - No API changes, only file organization
```
