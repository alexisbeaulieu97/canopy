# Change: Replace TUI Magic Strings with Typed Action Constants

## Why
The TUI uses magic strings like `"close"` and `"push"` in `ConfirmViewState.Action`. This is error-prone and lacks IDE/type-safety support.

## What Changes
- Define `Action` type as `type Action string`
- Define constants: `ActionClose Action = "close"`, `ActionPush Action = "push"`
- **BREAKING (internal)**: Update `ConfirmViewState.Action` from `string` to `Action`
- Update all action comparisons to use constants

## Impact
- Affected specs: `tui`
- Affected code:
  - `internal/tui/states.go` - Change Action field type
  - `internal/tui/update.go` - Use constants in handlers
  - Any other files comparing action strings
- **Risk**: Very Low - Internal API change only, simple type alias with constants

