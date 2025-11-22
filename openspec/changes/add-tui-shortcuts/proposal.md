# Change: Add TUI Keyboard Shortcuts

## Why
The TUI currently has limited keyboard shortcuts (`enter`, `s` sync, `c` close, `q` quit). Power users need quick access to common git operations and navigation without leaving the TUI. Additional shortcuts for fetch, pull, push, browser open, and filtering improve workflow efficiency.

## What Changes
- Add `f` key: fetch all repos in selected workspace
- Add `P` (shift+p) key: pull all repos in selected workspace
- Add `p` key: push all repos (with confirmation)
- Add `g` key: open repo(s) in browser
- Add `o` key: open workspace in editor
- Add `D` key: toggle filter to show only dirty workspaces
- Add `B` key: toggle filter to show only behind-remote workspaces
- Add `?` key: show help overlay with all shortcuts
- Add `r` key: refresh workspace list

## Impact
- Affected specs: `specs/tui/spec.md`
- Affected code:
  - `internal/tui/tui.go:228-296` - Add new key handlers
  - `internal/tui/tui.go:160-204` - Add help overlay view
  - `internal/tui/tui.go:35-51` - Add filter state to Model
