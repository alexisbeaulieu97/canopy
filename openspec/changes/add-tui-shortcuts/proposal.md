# Change: Add TUI Keyboard Shortcuts

## Why
The TUI previously exposed only basic navigation. This change documents and wires a concise shortcut set so users can act on the selected workspace without leaving the UI.

## What Changes
- `enter`: open workspace details
- `p`: push selected workspace (confirmation shown)
- `o`: open selected workspace in `$VISUAL`/`$EDITOR`
- `s`: toggle stale-only filter
- `/`: enter search mode (Bubble Tea list filter)
- `c`: close selected workspace (confirmation shown)
- `q`/`ctrl+c`: quit

Notes:
- Fetch (`f`), pull (`P`), open-in-browser (`g`), dirty/behind filters (`D`/`B`), help overlay (`?`), and refresh (`r`) are **not implemented** in the current TUI and remain out of scope for this change.

## Impact
- Affected specs: `specs/tui/spec.md`
- Affected code:
  - `internal/tui/tui.go` - shortcut handlers and help strings
