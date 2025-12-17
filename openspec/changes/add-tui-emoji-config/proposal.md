# Change: Add TUI Emoji Configuration

## Why

The TUI currently uses emoji characters directly (🌲, 💾, 📂, ⚠, ✓) which may not render correctly in all terminal environments, particularly:
- Older terminal emulators
- Some SSH sessions
- Windows Command Prompt (pre-Windows Terminal)
- Terminals without Unicode font support

Adding a configuration option allows users to opt for ASCII-only output when emoji rendering is problematic.

## What Changes

- Add `tui.use_emoji` configuration option (default: true for backward compatibility)
- Define ASCII fallbacks for all emoji used in TUI
- Update TUI rendering to conditionally use emoji or ASCII based on config
- Document the option in configuration guide

## Impact

- Affected specs: tui (accessibility)
- Affected code:
  - `internal/config/config.go` - Add UseEmoji field to TUIConfig
  - `internal/tui/view.go` - Use conditional rendering (~10 locations)
  - `internal/tui/styles.go` or new `symbols.go` - Define symbol mappings
  - `docs/configuration.md` - Document new option
- Risk: Low - additive feature, backward compatible default
