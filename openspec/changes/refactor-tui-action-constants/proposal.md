# Change: Replace TUI Magic Strings with Typed Action Constants

## Why
The TUI uses magic strings like `"close"` and `"push"` in `ConfirmViewState.Action`. This is error-prone:
1. **Typos compile but fail at runtime**: `"colse"` would compile but not match any handler
2. **No IDE support**: No autocomplete or refactoring support for string literals
3. **Documentation gap**: Unclear what valid actions exist

## What Changes
- Define `Action` type as `type Action string`
- Define constants: `ActionClose Action = "close"`, `ActionPush Action = "push"`, etc.
- Update `ConfirmViewState.Action` from `string` to `Action`
- Update all action comparisons to use constants

## Impact
- Affected specs: `tui`
- Affected code:
  - `internal/tui/states.go` - Change Action field type
  - `internal/tui/update.go` - Use constants in handlers
  - Any other files comparing action strings
- **Risk**: Very Low - Simple type alias with constants

