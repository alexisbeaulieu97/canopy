# Change: Add TUI Keyboard Customization

## Why
Power users may want to customize TUI keybindings:
- Conflict with terminal emulator shortcuts
- Personal preference (vim-style, emacs-style)
- Accessibility needs
- Consistency with other tools

Currently, keybindings are hardcoded in `internal/tui/`.

## What Changes
- Add `tui.keybindings` section to config.yaml:
  ```yaml
  tui:
    keybindings:
      quit: ["q", "ctrl+c"]
      search: ["/"]
      push: ["p"]
      close: ["c"]
      open_editor: ["o", "e"]
      toggle_stale: ["s"]
  ```
- Load keybindings from config at TUI startup
- Fall back to defaults if not configured
- Support multiple keys per action

## Impact
- **Affected specs**: `specs/tui-interface/spec.md`
- **Affected code**:
  - `internal/config/config.go` - Add TUI config section
  - `internal/tui/model.go` - Load keybindings
  - `internal/tui/update.go` - Use configurable keybindings
  - `internal/tui/view.go` - Show configured keys in help
- **Risk**: Low - Additive configuration, defaults unchanged
